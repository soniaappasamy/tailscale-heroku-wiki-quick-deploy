package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	name       = flag.String("name", "visitor", "The name to say hello to.")
	publicAddr = flag.String("public", ":80", "The address to bind to.")
)

func main() {
	flag.Parse()

	if err := startTailscale(context.Background()); err != nil {
		log.Fatal(err)
	}
	if err := startWikiServer(context.Background()); err != nil {
		log.Fatal(err)
	}
	if err := startPublicDummyServer(); err != nil {
		log.Fatal(err)
	}
}

const (
	stateFilePath = "/wiki/ts.state"
	socketPath    = "/wiki/ts.sock"
)

const tableQuery = `
CREATE TABLE IF NOT EXISTS "tailscale_data" (
	"id"    serial primary key,
	"state" text not null
);
`

func startTailscale(ctx context.Context) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	tsAuthKey := os.Getenv("TAILSCALE_AUTHKEY")
	if tsAuthKey == "" {
		return fmt.Errorf("TAILSCALE_AUTHKEY is not set")
	}

	db, err := sql.Open("postgres", databaseURL+"?sslmode=require")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, tableQuery)
	if err != nil {
		return fmt.Errorf("failed to create tailscale_data table: %w", err)
	}

	var tsState string
	err = db.QueryRowContext(ctx, "SELECT state FROM tailscale_data ORDER BY id DESC LIMIT 1").Scan(&tsState)
	if errors.Is(err, sql.ErrNoRows) {
	} else if err != nil {
		return fmt.Errorf("failed to read state file from database: %w", err)
	}
	if tsState != "" {
		err := os.WriteFile(stateFilePath, []byte(tsState), 0644)
		if err != nil {
			return fmt.Errorf("failed to write state file: %w", err)
		}
	}
	if tsAuthKey == "" && tsState == "" {
		return fmt.Errorf("TAILSCALE_AUTHKEY or state file must be present")
	}

	// Start `tailscaled`
	daemoncmd := exec.CommandContext(ctx, "/app/tailscaled", "--socket", socketPath, "--state", stateFilePath, "--tun", "userspace-networking")
	if err = daemoncmd.Start(); err != nil {
		return fmt.Errorf("failed to start tailscaled: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Start `tailscale`
	args := []string{"--socket", socketPath, "up", "--hostname", "wiki-server"}
	if tsAuthKey != "" {
		args = append(args, "--authkey", tsAuthKey)
	}
	cmd := exec.CommandContext(ctx, "/app/tailscale", args...)
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("failed to start tailscale: %w", err)
	}

	b, err := os.ReadFile(stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}
	if string(b) != tsState {
		_, err := db.ExecContext(ctx, "INSERT INTO tailscale_data(state) VALUES($1)", string(b))
		if err != nil {
			return fmt.Errorf("failed to update state file in database: %w", err)
		}
	}

	return nil
}

func startWikiServer(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "node", "/wiki/server")
	// Without specifying the port here, wiki.js use a random port Heroku gives it.
	// We force it to port 3000 so our go server knows where the service is running.
	cmd.Env = append(os.Environ(), "PORT=3000")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start wiki server: %w", err)
	}
	return nil
}

// startPublicDummyServer starts a go webserver that displays a simple welcome prompt to viewers.
// Heroku requires something to be running at it's public port, otherwise it shuts down the instance.
// We only want our wiki server acessible over Tailscale, though so we don't want to serve that over
// the public Heroku port. So we instead use this dummy server to satisfy Heroku.
func startPublicDummyServer() error {
	// Grab the Heroku random public port assigned to our instance.
	if port := os.Getenv("PORT"); port != "" {
		*publicAddr = ":" + port
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome, %s! Hello from %s", *name, hostname)
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		http.ListenAndServe(*publicAddr, publicMux)
		wg.Done()
	}()

	wg.Wait()

	return nil
}
