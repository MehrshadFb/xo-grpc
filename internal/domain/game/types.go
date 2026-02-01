package game

type Mark int
const (
	MarkUnspecified Mark = iota
	MarkEmpty
	MarkX
	MarkO
)

type GameStatus int
const (
	StatusUnspecified GameStatus = iota
	StatusWaiting
	StatusInProgress
	StatusFinished
	StatusAborted
)

type Game struct {
	ID       string
	JoinCode string
	Status   GameStatus
	Board [9]Mark
	NextTurn Mark
	Winner Mark
	IsDraw bool
	MoveNumber int64
	Version    int64
}
