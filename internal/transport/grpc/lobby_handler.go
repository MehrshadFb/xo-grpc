package grpc

import (
	"context"
	"errors"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	"github.com/MehrshadFb/xo-grpc/internal/service/lobby"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LobbyHandler struct {
	xov1.UnimplementedLobbyServiceServer

	service *lobby.Service
}

func NewLobbyHandler(service *lobby.Service) *LobbyHandler {
	return &LobbyHandler{
		service: service,
	}
}

func (h *LobbyHandler) CreateGame(ctx context.Context, req *xov1.CreateGameRequest) (*xov1.CreateGameResponse, error) {
	result, err := h.service.CreateGame(req.GetDisplayName())
	if err != nil {
		return nil, lobbyError(err)
	}

	return &xov1.CreateGameResponse{
		State:       toProtoGameState(result.Game),
		PlayerToken: result.PlayerToken,
	}, nil
}

func (h *LobbyHandler) JoinGame(ctx context.Context, req *xov1.JoinGameRequest) (*xov1.JoinGameResponse, error) {
	result, err := h.service.JoinGame(req.GetJoinCode(), req.GetDisplayName())
	if err != nil {
		return nil, lobbyError(err)
	}

	return &xov1.JoinGameResponse{
		State:       toProtoGameState(result.Game),
		PlayerToken: result.PlayerToken,
	}, nil
}

func lobbyError(err error) error {
	switch {
	case errors.Is(err, lobby.ErrEmptyDisplayName):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, lobby.ErrInvalidJoinCode):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, lobby.ErrGameFull):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, lobby.ErrGameNotWaiting):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}