package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <port>\n", os.Args[0])
		os.Exit(1)
	}

	port := os.Args[1]

	// Validate port number
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		fmt.Fprintf(os.Stderr, "Invalid port number: %s\n", port)
		os.Exit(1)
	}

	// Check if port is in use by attempting to listen on it
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		// Port is in use
		fmt.Printf("Port %s is in use\n", port)
		os.Exit(1)
	}

	// Port is available, close the listener
	if err := listener.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to close listener: %v\n", err)
	}
	fmt.Printf("Port %s is available\n", port)
	os.Exit(0)
}
