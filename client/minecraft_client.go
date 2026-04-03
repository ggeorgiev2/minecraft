package client

import (
	"fmt"
	"net"
	"time"
)

const (
	// Minecraft server version 1.21.11
	MINECRAFT_SERVER_VERSION = "1.21.11"
	SERVER_ADDRESS           = "localhost:12345"
)

func ConnectToServer() error {
	fmt.Printf("Attempting to connect to Minecraft server (version %s) at %s...\n", MINECRAFT_SERVER_VERSION, SERVER_ADDRESS)

	conn, err := net.DialTimeout("tcp", SERVER_ADDRESS, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	fmt.Printf("Successfully connected to Minecraft server at %s\n", SERVER_ADDRESS)

	// In a real client, you would now proceed with the Minecraft handshake protocol.
	// This involves sending various packets to the server.
	// For this task, we are only establishing the TCP connection.

	return nil
}
