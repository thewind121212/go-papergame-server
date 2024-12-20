package games

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go-socket/types"
	"go-socket/utils"
	"math/rand"
)

func Carogame(msgJSON types.GameMessage, conn *websocket.Conn, Games map[string]*types.Game) {
	switch msgJSON.Type {
	case "join":
		joinGame(msgJSON, conn, Games)
	case "move":
		gameMove(msgJSON, Games)
	case "leave":
		leaveGame(msgJSON, Games, conn)
	case "rematch":
		rematch(msgJSON.GameID, Games, conn)
	}

}

func joinGame(msgJSON types.GameMessage, conn *websocket.Conn, Games map[string]*types.Game) {

	if Games[msgJSON.GameID].P1ID != "" && Games[msgJSON.GameID].P2ID != "" && msgJSON.Data.PlayerID != Games[msgJSON.GameID].P1ID && msgJSON.Data.PlayerID != Games[msgJSON.GameID].P2ID {
		_ = conn.WriteJSON(map[string]string{"status": "Game Full"})
		return
	}

	if Games[msgJSON.GameID].P1ID == "" {
		Games[msgJSON.GameID].P1Name = msgJSON.Data.Name
		Games[msgJSON.GameID].P1ID = msgJSON.Data.PlayerID
	} else if Games[msgJSON.GameID].P2ID == "" && Games[msgJSON.GameID].P1ID != msgJSON.Data.PlayerID {
		Games[msgJSON.GameID].P2Name = msgJSON.Data.Name
		Games[msgJSON.GameID].P2ID = msgJSON.Data.PlayerID
	}

	if Games[msgJSON.GameID].P2ID != "" && Games[msgJSON.GameID].P1ID != "" {
		Games[msgJSON.GameID].Status = "All Players Joined"
	}

	if Games[msgJSON.GameID].P1Name != "" && Games[msgJSON.GameID].P2Name != "" && Games[msgJSON.GameID].Status == "All Players Joined" {
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
		_ = conn.WriteJSON(map[string]string{"status": "Game Start"})
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
}
