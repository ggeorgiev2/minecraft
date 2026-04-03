package client

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	SERVER_ADDRESS = "localhost:12345"
)

// MinecraftClient defines the interface for interacting with a Minecraft server
type MinecraftClient interface {
	Connect() error
	GetVersion() (string, error)
	Disconnect() error
}

// MinecraftClientImpl implements MinecraftClient and holds the TCP connection
type MinecraftClientImpl struct {
	address string
	conn    net.Conn
}

// NewMinecraftClient creates a new Minecraft client instance
func NewMinecraftClient(address string) MinecraftClient {
	return &MinecraftClientImpl{
		address: address,
	}
}

// Connect establishes a TCP connection to the Minecraft server
func (c *MinecraftClientImpl) Connect() error {
	if c.conn != nil {
		return fmt.Errorf("already connected")
	}

	conn, err := net.DialTimeout("tcp", c.address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.conn = conn
	return nil
}

// GetVersion retrieves the server version using the Minecraft status protocol
func (c *MinecraftClientImpl) GetVersion() (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected")
	}

	// Send handshake packet (next state = 1 for status)
	if err := c.sendHandshake(); err != nil {
		return "", fmt.Errorf("failed to send handshake: %w", err)
	}

	// Send status request
	if err := c.sendStatusRequest(); err != nil {
		return "", fmt.Errorf("failed to send status request: %w", err)
	}

	// Read status response
	version, err := c.readStatusResponse()
	if err != nil {
		return "", fmt.Errorf("failed to read status response: %w", err)
	}

	return version, nil
}

// Disconnect closes the TCP connection
func (c *MinecraftClientImpl) Disconnect() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// sendHandshake sends the initial handshake packet with next state = 1 (status)
func (c *MinecraftClientImpl) sendHandshake() error {
	// Handshake packet structure:
	// Packet ID: 0x00 (VarInt)
	// Protocol Version: 767 (1.21.1+, VarInt)
	// Server Address: "localhost" (String)
	// Server Port: 12345 (Unsigned Short)
	// Next State: 1 (Status, VarInt)

	packet := []byte{}

	// Packet ID (0x00)
	packet = append(packet, 0x00)

	// Protocol version (767 for 1.21.1+)
	packet = append(packet, writeVarInt(767)...)

	// Server address
	address := "localhost"
	packet = append(packet, writeVarInt(int32(len(address)))...)
	packet = append(packet, []byte(address)...)

	// Server port (12345)
	port := uint16(12345)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, port)
	packet = append(packet, portBytes...)

	// Next state (1 for status)
	packet = append(packet, 0x01)

	// Packet length prefix
	length := writeVarInt(int32(len(packet)))
	packet = append(length, packet...)

	_, err := c.conn.Write(packet)
	return err
}

// sendStatusRequest sends the status request packet
func (c *MinecraftClientImpl) sendStatusRequest() error {
	// Status request packet: just packet ID 0x00 with length 1
	packet := []byte{0x01, 0x00} // length 1, packet ID 0x00
	_, err := c.conn.Write(packet)
	return err
}

// readStatusResponse reads and parses the status response
func (c *MinecraftClientImpl) readStatusResponse() (string, error) {
	// Read packet length (VarInt)
	_, err := readVarInt(c.conn)
	if err != nil {
		return "", err
	}

	// Read packet ID (should be 0x00)
	packetID, err := readVarInt(c.conn)
	if err != nil {
		return "", err
	}
	if packetID != 0 {
		return "", fmt.Errorf("unexpected packet ID: %d", packetID)
	}

	// Read JSON string length
	jsonLength, err := readVarInt(c.conn)
	if err != nil {
		return "", err
	}

	// Read JSON string
	jsonBytes := make([]byte, jsonLength)
	_, err = io.ReadFull(c.conn, jsonBytes)
	if err != nil {
		return "", err
	}

	// Parse JSON
	var response struct {
		Version struct {
			Name string `json:"name"`
		} `json:"version"`
	}

	if err := json.Unmarshal(jsonBytes, &response); err != nil {
		return "", err
	}

	return response.Version.Name, nil
}

// writeVarInt encodes an int32 as a VarInt
func writeVarInt(value int32) []byte {
	var buf []byte
	for {
		temp := byte(value & 0x7F)
		value >>= 7
		if value != 0 {
			temp |= 0x80
		}
		buf = append(buf, temp)
		if value == 0 {
			break
		}
	}
	return buf
}

// readVarInt decodes a VarInt from the connection
func readVarInt(r io.Reader) (int32, error) {
	var value int32
	var shift uint
	for {
		var b byte
		if err := binary.Read(r, binary.BigEndian, &b); err != nil {
			return 0, err
		}
		value |= int32(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
		if shift >= 35 {
			return 0, fmt.Errorf("VarInt too big")
		}
	}
	return value, nil
}

// ConnectToServer maintains backward compatibility
func ConnectToServer() error {
	client := NewMinecraftClient(SERVER_ADDRESS)
	if err := client.Connect(); err != nil {
		return err
	}
	defer client.Disconnect()

	fmt.Printf("Successfully connected to Minecraft server at %s\n", SERVER_ADDRESS)

	// Try to get version
	version, err := client.GetVersion()
	if err != nil {
		fmt.Printf("Could not retrieve server version: %v\n", err)
	} else {
		fmt.Printf("Server version: %s\n", version)
	}

	return nil
}
