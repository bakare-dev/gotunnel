package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		runServer(os.Args[2:])
	case "client":
		runClient(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("GoTunnel v%s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:

		runClient(os.Args[1:])
	}
}

func printUsage() {
	usage := `GoTunnel v%s - Self-hosted TCP Tunneling

Usage:
  gotunnel <command> [options]

Commands:
  server          Start tunnel server
  client          Start tunnel client (default)
  version         Show version information
  help            Show this help message

Server Options:
  gotunnel server [options]
    --addr string           Listen address (default ":9000")
    --start-port int        Starting port for public listeners (default 10000)
    --tls                   Enable TLS encryption
    --tls-cert string       Path to TLS certificate (default "certs/server-cert.pem")
    --tls-key string        Path to TLS private key (default "certs/server-key.pem")

Client Options:
  gotunnel client [options]
  gotunnel [options]        (client is default)
    --server string         Tunnel server address (default "localhost:9000")
    --local string          Local service to expose (required, e.g. localhost:8080)
    --token string          Authentication token (default "dev-token")
    --tls                   Enable TLS encryption
    --tls-ca string         Path to CA certificate (default "certs/ca-cert.pem")
    --no-reconnect          Disable auto-reconnect on connection loss

Examples:
  # Start server
  gotunnel server --addr=:9000

  # Start server with TLS
  gotunnel server --tls --tls-cert=certs/server.pem --tls-key=certs/key.pem

  # Expose local service
  gotunnel client --server=tunnel.example.com:9000 --local=localhost:3000

  # Shorthand (client is default)
  gotunnel --server=tunnel.example.com:9000 --local=localhost:3000

  # With TLS
  gotunnel --server=tunnel.example.com:9000 --local=localhost:3000 --tls

Documentation: https://github.com/bakare-dev/gotunnel
Report bugs: https://github.com/bakare-dev/gotunnel/issues
`
	fmt.Printf(usage, version)
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)

	addr := fs.String("addr", ":9000", "Listen address")
	startPort := fs.Int("start-port", 10000, "Starting port for public listeners")
	tlsEnabled := fs.Bool("tls", false, "Enable TLS encryption")
	tlsCert := fs.String("tls-cert", "certs/server-cert.pem", "Path to TLS certificate")
	tlsKey := fs.String("tls-key", "certs/server-key.pem", "Path to TLS private key")

	fs.Parse(args)

	serverMain(*addr, *startPort, *tlsEnabled, *tlsCert, *tlsKey)
}

func runClient(args []string) {
	fs := flag.NewFlagSet("client", flag.ExitOnError)

	serverAddr := fs.String("server", "localhost:9000", "Tunnel server address")
	localAddr := fs.String("local", "", "Local service to expose (required)")
	token := fs.String("token", "dev-token", "Authentication token")
	tlsEnabled := fs.Bool("tls", false, "Enable TLS encryption")
	tlsCA := fs.String("tls-ca", "certs/ca-cert.pem", "Path to CA certificate")
	noReconnect := fs.Bool("no-reconnect", false, "Disable auto-reconnect")

	fs.Parse(args)

	if *localAddr == "" {
		fmt.Println("Error: --local flag is required")
		fmt.Println("\nUsage: gotunnel client --local localhost:3000 [options]")
		fmt.Println("   or: gotunnel --local localhost:3000 [options]")
		os.Exit(1)
	}

	clientMain(*serverAddr, *localAddr, *token, *tlsEnabled, *tlsCA, *noReconnect)
}
