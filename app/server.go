package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

// Only supported version
const (
	HTTPVersion = "HTTP/1.1"
)

var crlf = []byte("\r\n")

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

var routes = map[string]struct{}{
	"/": {},
}

func handleConn(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				fmt.Println("connection was closed")
				break
			}
			fmt.Println("failed reading bytes from conn:", err)
			continue
		}

		request, err := parseRequest(buf)
		if err != nil {
			fmt.Println("parsing request:", err)
			// TODO: return 404
			continue
		}

		if !routeExists(request.urlPath) {
			err = respond(conn, Response{status: 404, statusText: "Not Found"})
			if err != nil {
				fmt.Println("ERROR:,", err)
			}
		}

		err = respond(conn, Response{200, "OK"})
		if err != nil {
			fmt.Println("ERROR:,", err)
		}
	}
	conn.Close()
}

type Request struct {
	method  string
	urlPath string
	headers map[string]string
	body    []byte
	version string
}

func parseRequest(data []byte) (*Request, error) {
	var (
		request Request
		err     error
	)
	headers := make(map[string]string)

	// Split by end markers
	for _, line := range bytes.Split(data, crlf) {
		switch {
		case bytes.Contains(line, []byte(HTTPVersion)):
			request.method, request.urlPath, request.version, err = parseRequestLine(line)
			if err != nil {
				return nil, err
			}

		case bytes.Contains(line, []byte{':'}):
			k, v, err := parseHeader(line)
			if err != nil {
				return nil, fmt.Errorf("bad request")
			}
			headers[k] = v

		case bytes.Equal(line, crlf):
		// Blank line separator

		default:
			request.body = line
		}
	}
	return &request, nil
}

func parseRequestLine(line []byte) (method string, urlPath string, version string, err error) {
	// Split by fields the status line
	rline := string(line)
	fmt.Printf("parsing request line: %s\n", rline)
	parts := strings.Fields(rline)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("bad request")
	}
	method, urlPath, version = parts[0], parts[1], parts[2]
	if version != HTTPVersion {
		return "", "", "", fmt.Errorf("http version not supported: %s", version)
	}
	return method, urlPath, version, nil
}

func parseHeader(line []byte) (string, string, error) {
	sline := string(line)

	fmt.Printf("parsing header: %s\n", string(sline))

	parts := strings.SplitN(sline, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid header format")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func routeExists(urlPath string) bool {
	_, ok := routes[urlPath]
	return ok
}

type Response struct {
	status     int
	statusText string
}

func (r Response) String() string {
	return fmt.Sprintf("%s %d %s\r\n\r\n", HTTPVersion, r.status, r.statusText)
}

func respond(conn net.Conn, resp Response) error {
	_, err := conn.Write([]byte(resp.String()))
	return err
}
