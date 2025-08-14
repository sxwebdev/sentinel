package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/sxwebdev/sentinel/internal/utils"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error:", err)
	}
}

func run() error {
	listener, err := net.Listen("tcp", "127.0.0.1:12345")
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer listener.Close()

	fmt.Println("Server is listening on 127.0.0.1:12345")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}
		fmt.Println("Client connected:", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		fmt.Printf("Connection closed: %s\n\n", conn.RemoteAddr())
	}()

	// Set connection timeout - close if no data received in 5 seconds
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	var accumulated []byte
	reader := bufio.NewReader(conn)

	for {
		buffer := make([]byte, 1024)
		n, err := reader.Read(buffer)
		if err != nil {
			if utils.IsErrTimeout(err) {
				break
			}

			fmt.Println("failed to read from connection:", err)
			return
		}

		if n == 0 {
			break
		}

		// Accumulate received data
		accumulated = append(accumulated, buffer[:n]...)

		// If no more data buffered, client likely finished sending
		if reader.Buffered() == 0 {
			break
		}
	}

	// Now process the complete message
	if len(accumulated) == 0 {
		fmt.Println("No data received from client")
		return
	}

	msg := string(accumulated)
	fmt.Println("Complete message received:", msg)

	// Simple ping-pong protocol - exact match
	switch msg {
	case "ping":
		fmt.Println("Sending pong")
		_, err := conn.Write([]byte("pong"))
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	case "noresponse":
		fmt.Println("No response expected")
	default:
		fmt.Printf("Unknown message '%s', sending ok\n", msg)
		_, err := conn.Write([]byte("ok"))
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
