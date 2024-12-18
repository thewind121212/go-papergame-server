package main

import (
	"encoding/json"
	"fmt"
	"go-socket/utils"
	"log"
	"math/rand"

	"net/http"

	"github.com/gorilla/websocket"
)

type Game struct {
	ID           string
	Status       string
	P1ID         string
	P2ID         string
	P1Name       string
	P2Name       string
	PlayerTurn   string
	Grid         [][]string
	InteractGrid [15][15]string
	IsFinished   bool
	Winner       string
}

type GameMessage struct {
	GameID string `json:"gameID"`
	Type   string `json:"type"`
	Data   DataMessage
}

type DataMessage struct {
	Name       string `json:"name"`
	Coordinate string `json:"coordinate"`
	PlayerID   string `json:"playerID"`
}

const GRID_SIZE = 225

var GRID_VERTICAL = 15
var GRID_HORIZONTAL = 15
var GRID = [GRID_SIZE]int{}

var games = make(map[string]*Game)
var connections = make(map[string][]*websocket.Conn)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameID")

	//if exist := games[gameID]; exist == nil {
	//	http.Error(w, "Room Not Found", http.StatusBadRequest)
	//	return
	//}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	if _, exists := games[gameID]; !exists {
		games[gameID] = &Game{ID: gameID, Status: "Waiting for Player"}
	}

	connections[gameID] = append(connections[gameID], conn)

	err = conn.WriteJSON(games[gameID])
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

		// assgin P1Name
		var msgJSON GameMessage
		_ = json.Unmarshal([]byte(msg), &msgJSON)

		if msgJSON.Type == "join" && games[msgJSON.GameID].P1ID != "" && games[msgJSON.GameID].P2ID != "" {
			_ = conn.WriteJSON(map[string]string{"status": "Game Full"})
			return
		}

		if msgJSON.Type == "join" {
			if games[msgJSON.GameID].P1ID == "" {
				games[msgJSON.GameID].P1Name = msgJSON.Data.Name
				games[msgJSON.GameID].P1ID = msgJSON.Data.PlayerID
			} else if games[msgJSON.GameID].P2ID == "" && games[msgJSON.GameID].P1ID != msgJSON.Data.PlayerID {
				games[msgJSON.GameID].P2Name = msgJSON.Data.Name
				games[msgJSON.GameID].P2ID = msgJSON.Data.PlayerID
				games[msgJSON.GameID].Status = "All Players Joined"
			}
		}

		if games[msgJSON.GameID].P1Name != "" && games[msgJSON.GameID].P2Name != "" && games[msgJSON.GameID].Status == "All Players Joined" {
			games[msgJSON.GameID].Status = "Game Initialize"
			_ = conn.WriteJSON(map[string]string{"status": "Game Initialize"})
			rows, cols := 15, 15
			grid := make([][]string, rows)
			// Create grid and mark rows
			for i := 0; i < rows; i++ {
				rowData := make([]string, cols)
				for j := 0; j < cols; j++ {
					location := fmt.Sprintf("%d-%d", i+1, j+1)
					en, _ := utils.EncryptAES(location, "mysecretaeskey12")
					rowData[j] = en
				}
				grid[i] = rowData

			}

			randomNumber := rand.Intn(101)
			if randomNumber > 50 {
				games[msgJSON.GameID].PlayerTurn = "P1"
			} else {
				games[msgJSON.GameID].PlayerTurn = "P2"
			}
			//decide who goes first
			games[msgJSON.GameID].Grid = grid
			games[msgJSON.GameID].Status = "Game Start"
			_ = conn.WriteJSON(map[string]string{"status": "Game Start"})
		}

		if (msgJSON.Type == "move") && (games[msgJSON.GameID].Status == "Game Start") {
			// Decrypt the coordinate
			de, _ := utils.DecryptAES(msgJSON.Data.Coordinate, "mysecretaeskey12")
			fmt.Println("Decrypted Coordinate: ", de)
			// Split the coordinate 1-4 to 1, 4
			row, col, _ := utils.SplitString(de)

			//assign the interact grid
			games[msgJSON.GameID].InteractGrid[row-1][col-1] = "0"

			fmt.Println("Interact Grid: ", games[msgJSON.GameID].InteractGrid[row-1][col-1])

		}

		for i, clientConn := range connections[gameID] {
			err := clientConn.WriteJSON(games[gameID])
			if err != nil {
				log.Printf("Error sending update to client %d: %v", i, err)
				clientConn.Close()
				// Remove the disconnected client
				connections[gameID] = append(connections[gameID][:i], connections[gameID][i+1:]...)
			}
		}

		//if socket is close

		if err != nil {
			log.Println("Error sending updated game status:", err)
			break
		}
	}
}

func createRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	roomId, _ := utils.GenerateRoomId()

	for games[roomId] != nil {
		roomId, _ = utils.GenerateRoomId()
	}
	games[roomId] = &Game{ID: roomId, Status: "Waiting for Player"}
	_ = json.NewEncoder(w).Encode(map[string]string{"roomId": roomId})
}

func checkRoom(w http.ResponseWriter, r *http.Request) {
	roomId := r.URL.Query().Get("roomId")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if games[roomId] == nil {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "Room Not Found"})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "Room Found"})

}

func main() {

	http.HandleFunc("/game", handleConnection)
	http.HandleFunc("/create-room", createRoom)
	http.HandleFunc("/check-room", checkRoom)

	log.Println("Starting WebSocket server on port 4296")
	err := http.ListenAndServe(":4296", nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
