package e2e

import (
	"context"
	"net"
	"testing"
	"time"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	"github.com/MehrshadFb/xo-grpc/internal/realtime"
	gamesvc "github.com/MehrshadFb/xo-grpc/internal/service/game"
	"github.com/MehrshadFb/xo-grpc/internal/service/lobby"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
	transportgrpc "github.com/MehrshadFb/xo-grpc/internal/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func startTestServer(t *testing.T) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	store := memory.NewStore()
	sessions := session.NewManager()
	hub := realtime.NewHub()

	lobbySvc := lobby.NewService(store, sessions)
	gameSvc := gamesvc.NewService(store, sessions, hub)

	server := grpc.NewServer()

	xov1.RegisterLobbyServiceServer(server, transportgrpc.NewLobbyHandler(lobbySvc))
	xov1.RegisterGameServiceServer(server, transportgrpc.NewGameHandler(gameSvc, hub))

	go func() {
		_ = server.Serve(lis)
	}()

	t.Cleanup(func() {
		server.Stop()
		_ = lis.Close()
	})

	return lis.Addr().String()
}

func TestCreateJoinMoveAndWatchGame(t *testing.T) {
	addr := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	defer conn.Close()

	lobbyClient := xov1.NewLobbyServiceClient(conn)
	gameClient := xov1.NewGameServiceClient(conn)

	createResp, err := lobbyClient.CreateGame(ctx, &xov1.CreateGameRequest{
		DisplayName: "Alice",
	})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	if createResp.GetState().GetGameId() == "" {
		t.Fatalf("expected game id")
	}
	if createResp.GetPlayerToken() == "" {
		t.Fatalf("expected Alice token")
	}

	joinResp, err := lobbyClient.JoinGame(ctx, &xov1.JoinGameRequest{
		JoinCode:    createResp.GetState().GetJoinCode(),
		DisplayName: "Bob",
	})
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}

	if joinResp.GetPlayerToken() == "" {
		t.Fatalf("expected Bob token")
	}
	if joinResp.GetState().GetStatus() != xov1.GameStatus_GAME_STATUS_IN_PROGRESS {
		t.Fatalf("expected game in progress, got %v", joinResp.GetState().GetStatus())
	}

	stream, err := gameClient.WatchGame(ctx, &xov1.WatchGameRequest{
		GameId:      createResp.GetState().GetGameId(),
		PlayerToken: createResp.GetPlayerToken(),
	})
	if err != nil {
		t.Fatalf("WatchGame: %v", err)
	}

	snapshot, err := stream.Recv()
	if err != nil {
		t.Fatalf("WatchGame snapshot recv: %v", err)
	}

	if snapshot.GetType() != xov1.GameEventType_GAME_EVENT_TYPE_STATE_SNAPSHOT {
		t.Fatalf("expected snapshot event, got %v", snapshot.GetType())
	}

	moveResp, err := gameClient.MakeMove(ctx, &xov1.MakeMoveRequest{
		GameId:      createResp.GetState().GetGameId(),
		PlayerToken: createResp.GetPlayerToken(),
		CellIndex:   4,
	})
	if err != nil {
		t.Fatalf("MakeMove: %v", err)
	}

	if moveResp.GetState().GetBoard()[4] != xov1.Mark_MARK_X {
		t.Fatalf("expected cell 4 to be X, got %v", moveResp.GetState().GetBoard()[4])
	}

	event, err := stream.Recv()
	if err != nil {
		t.Fatalf("WatchGame move event recv: %v", err)
	}

	if event.GetType() != xov1.GameEventType_GAME_EVENT_TYPE_MOVE_MADE {
		t.Fatalf("expected move made event, got %v", event.GetType())
	}
	if event.GetState().GetBoard()[4] != xov1.Mark_MARK_X {
		t.Fatalf("expected streamed cell 4 to be X, got %v", event.GetState().GetBoard()[4])
	}
	if event.GetState().GetNextTurn() != xov1.Mark_MARK_O {
		t.Fatalf("expected next turn O, got %v", event.GetState().GetNextTurn())
	}
}