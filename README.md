# Heroku Deploy for Tailscale Wiki.js server

This repo is an Heroku app definition for deploying a private [Wiki.js](https://wiki.js.org/) server to Heroku that is only accessible over [Tailscale](https://tailscale.com/).

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/soniaappasamy/tailscale-heroku-wiki-quick-deploy)

After deploying, your wiki server will be accessible at 100.x.y.z:3000, where 100.x.y.z is the Tailscale IP of your new server (you can find this on your [admin panel](https://login.tailscale.com/admin/machines)).
