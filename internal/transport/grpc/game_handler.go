package grpc

import (
	"context"
	"errors"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
	gamesvc "github.com/MehrshadFb/xo-grpc/internal/service/game"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GameHandler struct {
	xov1.UnimplementedGameServiceServer

	service *gamesvc.Service
}

func NewGameHandler(service *gamesvc.Service) *GameHandler {
	return &GameHandler{
		service: service,
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