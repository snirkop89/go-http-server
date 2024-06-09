package main

import (
	"fmt"
	// Uncomment this block to pass the first stage
	"net"
	"os"
)

// Only supported version
const HTTPVersion = "HTTP/1.1"

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	if err := run(); err != nil {
		fmt.Println("run error:", err)
		os.Exit(1)
	}
}

func run() error {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		return fmt.Errorf("failed to bind to port 4221: %w", err)
	}
	defer func() { _ = l.Close() }()

	// TODO: handle signals, then loop on connections
	conn, err := l.Accept()
	if err != nil {
		return fmt.Errorf("error accepting connection: %w", err)
	}
	handleConn(conn)

	return nil
}

func handleConn(conn net.Conn) {
	err := respond(conn)
	if err != nil {
		fmt.Println("ERROR:,", err)
	}
	conn.Close()
}

func respond(conn net.Conn) error {
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	return err
}

// func parseRequest(data []byte) error {
//     // Split by end markers
//     bytes.Split()
// }
