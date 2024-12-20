package types

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
	NextTurnTo   string
	WinnerID     string
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
