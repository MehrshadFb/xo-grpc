package grpc

import (
	"context"
	"errors"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/metrics"
	"github.com/MehrshadFb/xo-grpc/internal/realtime"
	gamesvc "github.com/MehrshadFb/xo-grpc/internal/service/game"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GameHandler struct {
	xov1.UnimplementedGameServiceServer

	service *gamesvc.Service
	hub     *realtime.Hub
}

func NewGameHandler(service *gamesvc.Service, hub *realtime.Hub) *GameHandler {
	return &GameHandler{
		service: service,
		hub:     hub,
	}
}

func (h *GameHandler) GetState(ctx context.Context, req *xov1.GetStateRequest) (*xov1.GetStateResponse, error) {
	result, err := h.service.GetState(req.GetGameId(), req.GetPlayerToken())
	if err != nil {
		return nil, gameError(err)
	}

	return &xov1.GetStateResponse{
		State: toProtoGameState(result.Game),
	}, nil
}

func (h *GameHandler) MakeMove(ctx context.Context, req *xov1.MakeMoveRequest) (*xov1.MakeMoveResponse, error) {
	result, err := h.service.MakeMove(
		req.GetGameId(),
		req.GetPlayerToken(),
		int(req.GetCellIndex()),
	)
	if err != nil {
		return nil, gameError(err)
	}

	return &xov1.MakeMoveResponse{
		State: toProtoGameState(result.Game),
	}, nil
}

func (h *GameHandler) WatchGame(req *xov1.WatchGameRequest, stream xov1.GameService_WatchGameServer) error {
	if h.hub == nil {
		return status.Error(codes.Internal, "realtime hub is not configured")
	}

	result, err := h.service.GetState(req.GetGameId(), req.GetPlayerToken())
	if err != nil {
		return gameError(err)
	}

	if result.Game.Version > req.GetAfterVersion() {
		if err := stream.Send(&xov1.GameEvent{
			Type:    xov1.GameEventType_GAME_EVENT_TYPE_STATE_SNAPSHOT,
			EventId: result.Game.Version,
			State:   toProtoGameState(result.Game),
			Message: "initial game state",
		}); err != nil {
			return err
		}
	}

	metrics.ActiveWatchStreams.Inc()       // increment the number of active watch streams
	defer metrics.ActiveWatchStreams.Dec() // decrement the number of active watch streams when the client disconnects

	sub := h.hub.Subscribe(req.GetGameId())       // dedicated a private channel for this client to watch this game
	defer h.hub.Unsubscribe(req.GetGameId(), sub) // unsubscribe when the client disconnects

	for {
		select {
		case <-stream.Context().Done(): // client context is done = client disconnected
			return stream.Context().Err()

		case event, ok := <-sub: // receive game state updates from the realtime hub
			if !ok {
				return nil
			}

			if err := stream.Send(&xov1.GameEvent{
				Type:           toProtoEventType(event.Type),
				EventId:        event.Game.Version,
				State:          toProtoGameState(event.Game),
				GameOverReason: toProtoGameOverReason(event.GameOverReason),
				Message:        "game state updated",
			}); err != nil {
				return err
			}
		}
	}
}

func gameError(err error) error {
	switch {
	case errors.Is(err, gamesvc.ErrInvalidGameID):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, gamesvc.ErrInvalidToken):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, session.ErrSessionNotFound):
		return status.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, session.ErrSessionGameMismatch):
		return status.Error(codes.PermissionDenied, err.Error())

	case errors.Is(err, memory.ErrGameNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domaingame.ErrInvalidCellIndex):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domaingame.ErrCellOccupied):
		return status.Error(codes.FailedPrecondition, err.Error())

	case errors.Is(err, domaingame.ErrNotPlayersTurn):
		return status.Error(codes.FailedPrecondition, err.Error())

	case errors.Is(err, domaingame.ErrGameNotInProgress):
		return status.Error(codes.FailedPrecondition, err.Error())

	case errors.Is(err, domaingame.ErrGameFinished):
		return status.Error(codes.FailedPrecondition, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
