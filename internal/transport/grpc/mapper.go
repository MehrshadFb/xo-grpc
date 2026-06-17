package grpc

import (
	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/realtime"
)

func toProtoGameState(g *domaingame.Game) *xov1.GameState {
	if g == nil {
		return nil
	}

	board := make([]xov1.Mark, len(g.Board))
	for i, mark := range g.Board {
		board[i] = toProtoMark(mark)
	}

	return &xov1.GameState{
		GameId:   g.ID,
		JoinCode: g.JoinCode,
		Status:   toProtoStatus(g.Status),
		Board:    board,

		PlayerX: toProtoPlayer(g.PlayerX),
		PlayerO: toProtoPlayer(g.PlayerO),

		NextTurn: toProtoMark(g.NextTurn),
		Winner:   toProtoMark(g.Winner),
		IsDraw:   g.IsDraw,
		Version:  g.Version,
		Score: &xov1.MatchScore{
			XWins: g.XWins,
			OWins: g.OWins,
			Draws: g.Draws,
		},
		Rematch: &xov1.RematchState{
			XRequested: g.RematchXRequested,
			ORequested: g.RematchORequested,
		},
		RoundNumber: g.RoundNumber,
	}
}

func toProtoPlayer(p *domaingame.Player) *xov1.Player {
	if p == nil {
		return nil
	}

	return &xov1.Player{
		PlayerId:    p.ID,
		DisplayName: p.DisplayName,
		Mark:        toProtoMark(p.Mark),
	}
}

func toProtoMark(mark domaingame.Mark) xov1.Mark {
	switch mark {
	case domaingame.MarkEmpty:
		return xov1.Mark_MARK_EMPTY
	case domaingame.MarkX:
		return xov1.Mark_MARK_X
	case domaingame.MarkO:
		return xov1.Mark_MARK_O
	default:
		return xov1.Mark_MARK_UNSPECIFIED
	}
}

func toProtoStatus(status domaingame.GameStatus) xov1.GameStatus {
	switch status {
	case domaingame.StatusWaiting:
		return xov1.GameStatus_GAME_STATUS_WAITING
	case domaingame.StatusInProgress:
		return xov1.GameStatus_GAME_STATUS_IN_PROGRESS
	case domaingame.StatusFinished:
		return xov1.GameStatus_GAME_STATUS_FINISHED
	case domaingame.StatusAborted:
		return xov1.GameStatus_GAME_STATUS_ABORTED
	default:
		return xov1.GameStatus_GAME_STATUS_UNSPECIFIED
	}
}

func toProtoEventType(t realtime.EventType) xov1.GameEventType {
	switch t {
	case realtime.EventTypePlayerJoined:
		return xov1.GameEventType_GAME_EVENT_TYPE_PLAYER_JOINED
	case realtime.EventTypeMoveMade:
		return xov1.GameEventType_GAME_EVENT_TYPE_MOVE_MADE
	case realtime.EventTypeGameOver:
		return xov1.GameEventType_GAME_EVENT_TYPE_GAME_OVER
	case realtime.EventTypeRematchRequested:
		return xov1.GameEventType_GAME_EVENT_TYPE_REMATCH_REQUESTED
	case realtime.EventTypeRoundStarted:
		return xov1.GameEventType_GAME_EVENT_TYPE_ROUND_STARTED
	default:
		return xov1.GameEventType_GAME_EVENT_TYPE_UNSPECIFIED
	}
}

func toProtoGameOverReason(reason string) xov1.GameOverReason {
	switch reason {
	case "win":
		return xov1.GameOverReason_GAME_OVER_REASON_WIN
	case "draw":
		return xov1.GameOverReason_GAME_OVER_REASON_DRAW
	default:
		return xov1.GameOverReason_GAME_OVER_REASON_UNSPECIFIED
	}
}
