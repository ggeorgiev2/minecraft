package client

import (
	"net"
	"testing"
	"time"
)

func TestConnectToServer(t *testing.T) {
	// This is a dummy server for testing purposes.
	// In a real scenario, you would have a running Minecraft server.
	listener, err := net.Listen("tcp", SERVER_ADDRESS)
	if err != nil {
		t.Fatalf("Could not start dummy server: %v", err)
	}
	defer listener.Close()

	// Allow a moment for the listener to be ready
	time.Sleep(100 * time.Millisecond)

	err = ConnectToServer()
	if err != nil {
		t.Errorf("ConnectToServer failed: %v", err)
	}
}
