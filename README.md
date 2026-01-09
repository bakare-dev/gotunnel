# GoTunnel

GoTunnel is a secure, CLI-based tunneling system written in Go, inspired by ngrok.

## Features

-   Secure TLS tunnels
-   Token-based authentication
-   Connection multiplexing
-   Automatic tunnel expiration
-   Metrics & monitoring
-   No frontend, CLI-first

## Usage

```bash
gotunnel server
gotunnel client --server <ip:port> --local localhost:8080
```
