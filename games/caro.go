package games

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go-socket/types"
	"go-socket/utils"
	"math/rand"
	"time"
)

func Carogame(msgJSON types.GameMessage, conn *websocket.Conn, Games map[string]*types.Game, Connections map[string][]*websocket.Conn, GameChannel map[string]chan bool) {
	switch msgJSON.Type {
	case "join":
		joinGame(msgJSON, conn, Games, GameChannel)
	case "move":
		gameMove(msgJSON, Games)
	case "leave":
		leaveGame(msgJSON, Games, conn)
	case "rematch":
		rematch(msgJSON.GameID, Games, conn)
	case "disconnect":
		disconnect(msgJSON, Games, Connections, GameChannel)
	}

}

func joinGame(msgJSON types.GameMessage, conn *websocket.Conn, Games map[string]*types.Game, GameChannel map[string]chan bool) {

	if Games[msgJSON.GameID].P1ID != "" && Games[msgJSON.GameID].P2ID != "" && msgJSON.Data.PlayerID != Games[msgJSON.GameID].P1ID && msgJSON.Data.PlayerID != Games[msgJSON.GameID].P2ID {
		_ = conn.WriteJSON(map[string]string{"status": "Game Full"})
		return
	}

	if Games[msgJSON.GameID].P1ID == "" {
		Games[msgJSON.GameID].P1Name = msgJSON.Data.Name
		Games[msgJSON.GameID].P1ID = msgJSON.Data.PlayerID
		roomInfo := types.GameStatus{
			Id:            msgJSON.GameID,
			CurrentPlayer: 1,
			Player1Id:     Games[msgJSON.GameID].P1ID,
			Player2Id:     "",
			Status:        "Waiting for Player",
		}

		roomInfoJson, _ := json.Marshal(roomInfo)
		_ = utils.SetKey(msgJSON.GameID, string(roomInfoJson), 0)
	} else if Games[msgJSON.GameID].P2ID == "" && Games[msgJSON.GameID].P1ID != msgJSON.Data.PlayerID {
		Games[msgJSON.GameID].P2Name = msgJSON.Data.Name
		Games[msgJSON.GameID].P2ID = msgJSON.Data.PlayerID
		moveToken, _ := utils.EncryptAES(Games[msgJSON.GameID].P1ID+Games[msgJSON.GameID].P2ID, "mysecretaeskey12")
		Games[msgJSON.GameID].MoveToken = moveToken
		roomInfo := types.GameStatus{
			Id:            msgJSON.GameID,
			CurrentPlayer: 2,
			Player1Id:     Games[msgJSON.GameID].P1ID,
			Player2Id:     Games[msgJSON.GameID].P2ID,
			Status:        "All Players Joined",
		}

		roomInfoJson, _ := json.Marshal(roomInfo)

		_ = utils.SetKey(msgJSON.GameID, string(roomInfoJson), 0)
	}

	if Games[msgJSON.GameID].P2ID != "" && Games[msgJSON.GameID].P1ID != "" {
		Games[msgJSON.GameID].Status = "All Players Joined"
	}

	if Games[msgJSON.GameID].P1Name != "" && Games[msgJSON.GameID].P2Name != "" && Games[msgJSON.GameID].Status == "All Players Joined" {
		go ClearTimeout(msgJSON.GameID, GameChannel)
		Games[msgJSON.GameID].Status = "Game Initialize"
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

		//decide who goes first by random number
		if Games[msgJSON.GameID].PlayerTurn == "" {
			randomNumber := rand.Intn(101)
			fmt.Println("Random Number: ", randomNumber)
			if randomNumber > 50 {
				Games[msgJSON.GameID].PlayerTurn = "P1"
				Games[msgJSON.GameID].NextTurnTo = "P2"
			} else {
				Games[msgJSON.GameID].PlayerTurn = "P2"
				Games[msgJSON.GameID].NextTurnTo = "P1"
			}
		}

		Games[msgJSON.GameID].Grid = grid
		Games[msgJSON.GameID].Status = "Game Start"

		roomInfo := types.GameStatus{
			Id:            msgJSON.GameID,
			CurrentPlayer: 2,
			Player1Id:     Games[msgJSON.GameID].P1ID,
			Player2Id:     Games[msgJSON.GameID].P2ID,
			Status:        "Game Start",
		}
		roomInfoJson, _ := json.Marshal(roomInfo)
		_ = utils.SetKey(msgJSON.GameID, string(roomInfoJson), 0)

		_ = conn.WriteJSON(map[string]string{"status": "Game Started"})
	}

}

func checkWin(row int, col int, grid [15][15]string, player string) bool {
	// Check row
	count := 0
	for i := 0; i < 5; i++ {
		if grid[row][i] == player {
			count++
			if count == 5 {
				return true
			}
		} else {
			count = 0
		}
	}

	// Check column
	count = 0
	for i := 0; i < 15; i++ {
		if grid[i][col] == player {
			count++
			if count == 5 {
				return true
			}
		} else {
			count = 0
		}
	}

	//Check diagonal

	count = 1
	for i := 1; i < 5; i++ {
		topRight := 0
		bottomLeft := 0
		if col+i < 15 && row-i >= 0 {
			if grid[row-i][col+i] == player {
				topRight++
			}
		}
		if col-i >= 0 && row+i < 15 {
			if grid[row+i][col-i] == player {
				bottomLeft++
			}
		}
		count += topRight + bottomLeft

		if count == 5 {
			return true
		}
	}

	count = 1
	for i := 1; i < 5; i++ {
		topLeft := 0
		bottomRight := 0
		if col-i >= 0 && row-i >= 0 {
			if grid[row-i][col-i] == player {
				topLeft++
			}
		}
		if col+i < 15 && row+i < 15 {
			if grid[row+i][col+i] == player {
				bottomRight++
			}
		}
		count += topLeft + bottomRight

		if count == 5 {
			return true
		}
	}

	return false
}

func gameMove(msgJSON types.GameMessage, Games map[string]*types.Game) {
	if Games[msgJSON.GameID].Status == "Game Start" {

		de, _ := utils.DecryptAES(msgJSON.Data.Coordinate, "mysecretaeskey12")
		// Split the coordinate 1-4 to 1, 4

		row, col, _ := utils.SplitString(de)
		fmt.Println("Row: ", row-1, "Col: ", col-1)
		//assign the interact grid
		Games[msgJSON.GameID].InteractGrid[row-1][col-1] = Games[msgJSON.GameID].PlayerTurn

		iswin := checkWin(row-1, col-1, Games[msgJSON.GameID].InteractGrid, Games[msgJSON.GameID].PlayerTurn)
		if iswin {
			Games[msgJSON.GameID].IsFinished = true
			Games[msgJSON.GameID].Winner = Games[msgJSON.GameID].PlayerTurn
			//reset game

			Games[msgJSON.GameID].Status = "Game Finished"
			if Games[msgJSON.GameID].PlayerTurn == "P1" {
				Games[msgJSON.GameID].WinnerID = Games[msgJSON.GameID].P1ID
			} else {
				Games[msgJSON.GameID].WinnerID = Games[msgJSON.GameID].P2ID
			}
			roomInfo := types.GameStatus{
				Id:            msgJSON.GameID,
				CurrentPlayer: 2,
				Player1Id:     Games[msgJSON.GameID].P1ID,
				Player2Id:     Games[msgJSON.GameID].P2ID,
				Status:        "Game Finished",
			}
			roomInfoJson, _ := json.Marshal(roomInfo)
			_ = utils.SetKey(msgJSON.GameID, string(roomInfoJson), 0)
			return
		}

		if Games[msgJSON.GameID].PlayerTurn == "P1" {
			Games[msgJSON.GameID].PlayerTurn = "P2"
		} else {
			Games[msgJSON.GameID].PlayerTurn = "P1"
		}

	}
}

func resetGame(GameID string, Games map[string]*types.Game) {
	Games[GameID].PlayerTurn = ""
	Games[GameID].Grid = nil
	Games[GameID].InteractGrid = [15][15]string{}
	Games[GameID].IsFinished = false
	Games[GameID].Winner = ""
	Games[GameID].NextTurnTo = ""
	Games[GameID].WinnerID = ""
}

func leaveGame(msgJSON types.GameMessage, Games map[string]*types.Game, conn *websocket.Conn) {
	if Games[msgJSON.GameID].P1ID == msgJSON.Data.PlayerID {
		Games[msgJSON.GameID].P1ID = ""
		Games[msgJSON.GameID].P1Name = ""
		Games[msgJSON.GameID].Status = "One Player Left"
		resetGame(msgJSON.GameID, Games)
	} else if Games[msgJSON.GameID].P2ID == msgJSON.Data.PlayerID {
		Games[msgJSON.GameID].P2ID = ""
		Games[msgJSON.GameID].P2Name = ""
		Games[msgJSON.GameID].Status = "One Player Left"
		resetGame(msgJSON.GameID, Games)
	}

	if Games[msgJSON.GameID].P1ID == "" && Games[msgJSON.GameID].P2ID == "" {
		delete(Games, msgJSON.GameID)
	}

	roomInfo := types.GameStatus{
		Id:            msgJSON.GameID,
		CurrentPlayer: 1,
		Player1Id:     Games[msgJSON.GameID].P1ID,
		Player2Id:     Games[msgJSON.GameID].P2ID,
		Status:        "Waiting for Player",
	}
	roomInfoJson, _ := json.Marshal(roomInfo)
	_ = utils.SetKey(msgJSON.GameID, string(roomInfoJson), 0)

	_ = conn.WriteJSON(map[string]string{"Status": "Player left the game", "gameID": msgJSON.GameID, "PlayerID": msgJSON.Data.PlayerID})
}

func rematch(GameID string, Games map[string]*types.Game, conn *websocket.Conn) {
	if !Games[GameID].IsFinished {
		_ = conn.WriteJSON(map[string]string{"Status": "Game is not finished"})
		return
	}

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

	Games[GameID].PlayerTurn = Games[GameID].NextTurnTo
	Games[GameID].Grid = grid
	Games[GameID].InteractGrid = [15][15]string{}
	Games[GameID].IsFinished = false
	Games[GameID].Winner = ""
	Games[GameID].WinnerID = ""
	Games[GameID].Status = "Game Start"

	roomInfo := types.GameStatus{
		Id:            GameID,
		CurrentPlayer: 2,
		Player1Id:     Games[GameID].P1ID,
		Player2Id:     Games[GameID].P2ID,
		Status:        "Game Started",
	}
	roomInfoJson, _ := json.Marshal(roomInfo)
	_ = utils.SetKey(GameID, string(roomInfoJson), 0)
}

func disconnect(msgJSON types.GameMessage, Games map[string]*types.Game, Connections map[string][]*websocket.Conn, GameChannel map[string]chan bool) {
	if Games[msgJSON.GameID].P1ID == msgJSON.Data.PlayerID {
		Games[msgJSON.GameID].P1ID = ""
		Games[msgJSON.GameID].P1Name = ""
		Games[msgJSON.GameID].Status = "One Player Left"
		resetGame(msgJSON.GameID, Games)
	} else if Games[msgJSON.GameID].P2ID == msgJSON.Data.PlayerID {
		Games[msgJSON.GameID].P2ID = ""
		Games[msgJSON.GameID].P2Name = ""
		Games[msgJSON.GameID].Status = "One Player Left"
		resetGame(msgJSON.GameID, Games)
	}

	fmt.Println("run disconnect")

	go SetTimeout(GameChannel, msgJSON.GameID, 5*time.Second, func() {
		delete(Games, msgJSON.GameID)
	_:
		utils.ClearConnections(msgJSON.GameID, Connections)
		_ = utils.DelKey(msgJSON.GameID)
	})

	if Games[msgJSON.GameID].P1ID == "" && Games[msgJSON.GameID].P2ID == "" {
		delete(Games, msgJSON.GameID)
		utils.ClearConnections(msgJSON.GameID, Connections)
	_:
		_ = utils.DelKey(msgJSON.GameID)
	}

}

func SetTimeout(GameChannel map[string]chan bool, gameID string, duration time.Duration, callback func()) {
	game, exists := GameChannel[gameID]
	if !exists {
		fmt.Println("Game not found:", gameID)
		return
	}

	go func() {
		select {
		case <-time.After(duration): // Timeout expires
			callback()
		case <-game:
			fmt.Println("Timeout cleared for game:", gameID)
		}
	}()
}

// ClearTimeout clears the timeout for a specific game ID
func ClearTimeout(gameID string, GameChannel map[string]chan bool) {
	game, exists := GameChannel[gameID]
	if !exists {
		fmt.Println("Game not found:", gameID)
		return
	}
	game <- true // Send a signal to cancel the timeout
}
