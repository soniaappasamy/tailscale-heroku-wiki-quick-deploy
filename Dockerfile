FROM golang:1.16.2-alpine3.13 as builder
WORKDIR /app
COPY . ./
RUN go build -o app ./main.go

FROM alpine:latest as tailscale
WORKDIR /app
ENV TSFILE=tailscale_1.14.0_amd64.tgz
RUN wget https://pkgs.tailscale.com/stable/${TSFILE} && \
  tar xzf ${TSFILE} --strip-components=1

FROM requarks/wiki:2 as wiki
WORKDIR /wiki
ENV DB_TYPE postgres
ENV DB_SSL 1
ENV HEROKU 1
ENV PGSSLMODE no-verify
COPY . .

# Copy binary to production image
COPY --from=builder /app/app /app/app
COPY --from=tailscale /app/tailscaled /app/tailscaled
COPY --from=tailscale /app/tailscale /app/tailscale

# Run on container startup.
CMD ["/app/app"]
