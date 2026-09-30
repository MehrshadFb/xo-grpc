package e2e

import (
	"context"
	"testing"
	"time"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestWatchGameNotifiesWhenOpponentLeavesAndReturns(t *testing.T) {
	const grace = 100 * time.Millisecond
	addr := startTestServerWithPresence(t, grace)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	defer conn.Close()

	lobbyClient := xov1.NewLobbyServiceClient(conn)
	gameClient := xov1.NewGameServiceClient(conn)

	createResp, err := lobbyClient.CreateGame(ctx, &xov1.CreateGameRequest{DisplayName: "Alice"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	gameID := createResp.GetState().GetGameId()

	joinResp, err := lobbyClient.JoinGame(ctx, &xov1.JoinGameRequest{
		JoinCode:    createResp.GetState().GetJoinCode(),
		DisplayName: "Bob",
	})
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}

	watch := func(ctx context.Context, token string) xov1.GameService_WatchGameClient {
		t.Helper()
		stream, err := gameClient.WatchGame(ctx, &xov1.WatchGameRequest{GameId: gameID, PlayerToken: token})
		if err != nil {
			t.Fatalf("WatchGame: %v", err)
		}
		recvType(t, stream, xov1.GameEventType_GAME_EVENT_TYPE_STATE_SNAPSHOT)
		return stream
	}

	aliceStream := watch(ctx, createResp.GetPlayerToken())

	bobCtx, bobLeaves := context.WithCancel(ctx)
	watch(bobCtx, joinResp.GetPlayerToken())
	bobLeaves()

	left := recvType(t, aliceStream, xov1.GameEventType_GAME_EVENT_TYPE_PLAYER_LEFT)
	if left.GetPlayerMark() != xov1.Mark_MARK_O {
		t.Fatalf("expected PLAYER_LEFT for O, got %v", left.GetPlayerMark())
	}

	aliceReloaded := watch(ctx, createResp.GetPlayerToken())
	recvType(t, aliceReloaded, xov1.GameEventType_GAME_EVENT_TYPE_PLAYER_LEFT)

	watch(ctx, joinResp.GetPlayerToken())

	returned := recvType(t, aliceStream, xov1.GameEventType_GAME_EVENT_TYPE_PLAYER_RETURNED)
	if returned.GetPlayerMark() != xov1.Mark_MARK_O {
		t.Fatalf("expected PLAYER_RETURNED for O, got %v", returned.GetPlayerMark())
	}
}

func recvType(t *testing.T, stream xov1.GameService_WatchGameClient, want xov1.GameEventType) *xov1.GameEvent {
	t.Helper()

	event, err := stream.Recv()
	if err != nil {
		t.Fatalf("waiting for %v: %v", want, err)
	}
	if event.GetType() != want {
		t.Fatalf("expected %v, got %v", want, event.GetType())
	}
	return event
}
