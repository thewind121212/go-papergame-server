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
var GameChannel = make(map[string]chan bool)
var Connections = struct {
	Player    map[string][]*websocket.Conn
	Spectator map[string][]*websocket.Conn
}{
	Player:    make(map[string][]*websocket.Conn),
	Spectator: make(map[string][]*websocket.Conn),
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameID")

	//reject upgrade if room not found
	if _, exists := Games[gameID]; !exists {
		_ = json.NewEncoder(w).Encode(map[string]string{"Status": "Room Not Found"})
		return
	}

	//if current socket > 2 then reject upgrade

	if len(Connections.Player[gameID])+1 > 2 {
		_ = json.NewEncoder(w).Encode(map[string]string{"Status": "Game Full"})
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}

	if len(Connections.Player[gameID])+1 <= 2 {

		fmt.Println(len(Connections.Player[gameID]) + 1)

		if Games[gameID].Status == "One Player Disconnected" && len(Connections.Player[gameID])+1 == 2 {
			fmt.Println("Reconnect")
			go games.ClearTimeout(gameID, GameChannel)
			if Games[gameID].P2ID != "" && Games[gameID].P1ID != "" && Games[gameID].Grid != nil {
				Games[gameID].Status = "Game Start"
			}

		}

		if Games[gameID].Status == "One Player Disconnected" && len(Connections.Player[gameID])+1 == 1 {
			conn.WriteJSON(map[string]string{"Status": "Room Disconnected"})
		}
	}

	//create room info

	romInfo := types.GameStatus{
		Id:            gameID,
		CurrentPlayer: 0,
		Player1Id:     "",
		Player2Id:     "",
		Status:        "Waiting for Player",
	}

	var roomInfo, _ = json.Marshal(romInfo)

	err = utils.SetKey(gameID, string(roomInfo), 0)
	if err != nil {
		log.Println("Error setting key abort upgrade socket", err)
		return
	}

	Connections.Player[gameID] = append(Connections.Player[gameID], conn)

	GameChannel[gameID] = make(chan bool)

	defer func() {
		_ = conn.Close()
		utils.RemoveConnection(gameID, conn, Connections.Player)
		utils.RemoveConnection(gameID, conn, Connections.Spectator)
	}()

	err = conn.WriteJSON(Games[gameID])
	if err != nil {
		log.Println("Error sending game status:", err)
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		//parse msg to json
		var msgJSON types.GameMessage
		_ = json.Unmarshal([]byte(msg), &msgJSON)
		if err != nil {
			log.Println("Error reading message:", err)
			if Games[gameID].Status != "One Player Left" && Games[gameID].Status != "Waiting for Player" {
				games.Disconnect(gameID, Games, Connections.Player, GameChannel)
			}
			//break
		}
		//check connndisconnect

		fmt.Printf("Received message for game %s: %s\n", gameID, msg)

		//handle msg to game
		games.Carogame(msgJSON, conn, Games, Connections.Player, GameChannel)

		//send msg to all socket client in room
		for i, clientConn := range Connections.Player[gameID] {
			err := clientConn.WriteJSON(Games[gameID])
			if err != nil {
				fmt.Println("cant send msg to clien with gameID", gameID)
				clientConn.Close()
				// Remove the disconnected client
				Connections.Player[gameID] = append(Connections.Player[gameID][:i], Connections.Player[gameID][i+1:]...)
			}
		}

		if err != nil {
			log.Println("Error sending updated game status:", err)
			break
		}
	}
}

func main() {

	// if redis client not connect then exit
	utils.InitRedisClient("localhost:6379", "", 1)
	err := utils.PingRedis()
	if err != nil {
		log.Fatal("Error connecting to Redis:", err)
		return
	}
	http.HandleFunc("/game", handleConnection)
	http.HandleFunc("/create-room", services.CreateRoom(Games))
	http.HandleFunc("/check-room", services.CheckRoom(Games))
	http.HandleFunc("/get-all-room", services.GetAllRoom(Games))

	log.Println("Starting WebSocket server on port 4296")
	err = http.ListenAndServe(":4296", nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
