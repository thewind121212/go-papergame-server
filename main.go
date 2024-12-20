package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go-socket/games"
	"go-socket/services"
	"go-socket/types"
	"go-socket/utils"
	"log"
	"net/http"
)

const GRID_SIZE = 225

var Games = make(map[string]*types.Game)
var connections = make(map[string][]*websocket.Conn)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameID")

	//reject upgrade if room not found
	if _, exists := Games[gameID]; !exists {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "Room Not Found"})
		return
	}

	//if current socket > 2 then reject upgrade

	if len(connections[gameID])+1 > 2 {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "Game Full"})
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}

	connections[gameID] = append(connections[gameID], conn)

	defer func() {
		_ = conn.Close()
		utils.RemoveConnection(gameID, conn, connections)
	}()

	err = conn.WriteJSON(Games[gameID])
	if err != nil {
		log.Println("Error sending game status:", err)
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		fmt.Printf("Received message for game %s: %s\n", gameID, msg)

		//parse msg to json
		var msgJSON types.GameMessage
		_ = json.Unmarshal([]byte(msg), &msgJSON)

		//handle msg to game
		games.Carogame(msgJSON, conn, Games)

		//send msg to all socket client in room
		for i, clientConn := range connections[gameID] {
			err := clientConn.WriteJSON(Games[gameID])
			if err != nil {
				log.Printf("Error sending update to client %d: %v", i, err)
				clientConn.Close()
				// Remove the disconnected client
				connections[gameID] = append(connections[gameID][:i], connections[gameID][i+1:]...)
			}
		}

		if err != nil {
			log.Println("Error sending updated game status:", err)
			break
		}
	}
}

func main() {

	http.HandleFunc("/game", handleConnection)
	http.HandleFunc("/create-room", services.CreateRoom(Games))
	http.HandleFunc("/check-room", services.CheckRoom(Games))

	log.Println("Starting WebSocket server on port 4296")
	err := http.ListenAndServe(":4296", nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
