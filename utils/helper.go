package utils

import (
	"crypto/rand"
	"fmt"
	"github.com/gorilla/websocket"
	"math/big"
	"strconv"
	"strings"
)

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
