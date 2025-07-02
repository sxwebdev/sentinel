package main

import (
	"bufio"
	"fmt"
	"net"
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
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		fmt.Println("Received:", msg)
		_, err := conn.Write([]byte("OK\n"))
		if err != nil {
			fmt.Println("Failed to send response:", err)
			return
		}
	}
}
