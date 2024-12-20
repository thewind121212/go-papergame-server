package services

import (
	"encoding/json"
	"go-socket/types"
	"go-socket/utils"
	"net/http"
)

func CreateRoom(games map[string]*types.Game) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		roomId, _ := utils.GenerateRoomId()

		for games[roomId] != nil {
			roomId, _ = utils.GenerateRoomId()
		}
		games[roomId] = &types.Game{ID: roomId, Status: "Waiting for Player"}
		_ = json.NewEncoder(w).Encode(map[string]string{"roomId": roomId})
	}
}

func CheckRoom(games map[string]*types.Game) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomId := r.URL.Query().Get("roomId")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if games[roomId] == nil {
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "Room Not Found"})
			return
		}

		if games[roomId].Status == "Game Start" {
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "Game Full"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "Waiting for Player"})
	}

}
