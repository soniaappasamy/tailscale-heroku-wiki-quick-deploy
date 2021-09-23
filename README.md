# Heroku Deploy for Tailscale Wiki.js server

This repo is an Heroku app definition for deploying a private [Wiki.js](https://wiki.js.org/) server to Heroku that is only accessible over [Tailscale](https://tailscale.com/).

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/soniaappasamy/tailscale-heroku-wiki-quick-deploy)

After deploying, your wiki server will be accessible at `100.x.y.z:3000`, where `100.x.y.z` is the Tailscale IP of your new server (you can find this on your [admin panel](https://login.tailscale.com/admin/machines)).

The first time you visit your wiki, you'll be asked a couple of configuration questions. The email and password fields allow you to create an initial administrator for your wiki (this user gets added to your wiki's private DB). Use `http://<Tailscale 100.x.y.z IP>:3000` as the site URL.

<img width="400" alt="Screen Shot 2021-09-23 at 1 27 13 PM" src="https://user-images.githubusercontent.com/9019214/134555588-6ce0f47b-fa86-479f-8298-bbca15080691.png">
