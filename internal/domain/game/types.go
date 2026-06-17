package game

type Player struct {
	ID          string
	DisplayName string
	Mark        Mark // X or O
}

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
	ID                string
	JoinCode          string
	Status            GameStatus
	Board             [9]Mark
	NextTurn          Mark
	Winner            Mark
	IsDraw            bool
	MoveNumber        int64
	Version           int64
	PlayerX           *Player
	PlayerO           *Player
	XWins             int64
	OWins             int64
	Draws             int64
	RoundNumber       int64
	RematchXRequested bool
	RematchORequested bool
}
