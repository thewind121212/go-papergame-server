package utils

import (
	"crypto/rand"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/big"
	"strconv"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func SplitString(input string) (int, int, error) {
	parts := strings.Split(input, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid format: %s", input)
	}

	x, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid number: %v", err)
	}

	y, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid number: %v", err)
	}

	return x, y, nil
}

func GenerateRoomId() (string, error) {
	length := 6
	roomID := make([]byte, length)
	for i := range roomID {
		randomByte, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		roomID[i] = charset[randomByte.Int64()]
	}
	return string(roomID), nil
}

func RemoveConnection(gameID string, conn *websocket.Conn, connections map[string][]*websocket.Conn) {
	if conns, ok := connections[gameID]; ok {
		for i, clientConn := range conns {
			if clientConn == conn {
				connections[gameID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
	}
}

func ClearConnections(gameID string, Connections map[string][]*websocket.Conn) {
	// Check if the gameID exists in the Connections map
	if conns, exists := Connections[gameID]; exists {
		for i, conn := range conns {
			// Close each connection
			err := conn.Close()
			if err != nil {
				log.Printf("Error closing connection %d for game %s: %v", i, gameID, err)
			}
		}
		// Remove the gameID from the Connections map
		delete(Connections, gameID)
		log.Printf("Cleared all connections for game %s", gameID)
	} else {
		log.Printf("No connections found for game %s", gameID)
	}
}
