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

	lobbySvc := lobby.NewService(store, sessions, hub)
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

	gameID := createResp.GetState().GetGameId()
	aliceToken := createResp.GetPlayerToken()
	joinCode := createResp.GetState().GetJoinCode()

	if gameID == "" {
		t.Fatalf("expected game id")
	}
	if aliceToken == "" {
		t.Fatalf("expected Alice token")
	}
	if joinCode == "" {
		t.Fatalf("expected join code")
	}

	stream, err := gameClient.WatchGame(ctx, &xov1.WatchGameRequest{
		GameId:      gameID,
		PlayerToken: aliceToken,
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
	if snapshot.GetState().GetStatus() != xov1.GameStatus_GAME_STATUS_WAITING {
		t.Fatalf("expected waiting snapshot, got %v", snapshot.GetState().GetStatus())
	}

	joinResp, err := lobbyClient.JoinGame(ctx, &xov1.JoinGameRequest{
		JoinCode:    joinCode,
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

	playerJoinedEvent, err := stream.Recv()
	if err != nil {
		t.Fatalf("WatchGame player joined recv: %v", err)
	}
	if playerJoinedEvent.GetType() != xov1.GameEventType_GAME_EVENT_TYPE_PLAYER_JOINED {
		t.Fatalf("expected player joined event, got %v", playerJoinedEvent.GetType())
	}
	if playerJoinedEvent.GetState().GetPlayerO().GetDisplayName() != "Bob" {
		t.Fatalf("expected PlayerO Bob, got %q", playerJoinedEvent.GetState().GetPlayerO().GetDisplayName())
	}
	if playerJoinedEvent.GetState().GetStatus() != xov1.GameStatus_GAME_STATUS_IN_PROGRESS {
		t.Fatalf("expected in progress after join, got %v", playerJoinedEvent.GetState().GetStatus())
	}

	moveResp, err := gameClient.MakeMove(ctx, &xov1.MakeMoveRequest{
		GameId:      gameID,
		PlayerToken: aliceToken,
		CellIndex:   4,
	})
	if err != nil {
		t.Fatalf("MakeMove: %v", err)
	}

	if moveResp.GetState().GetBoard()[4] != xov1.Mark_MARK_X {
		t.Fatalf("expected cell 4 to be X, got %v", moveResp.GetState().GetBoard()[4])
	}

	moveEvent, err := stream.Recv()
	if err != nil {
		t.Fatalf("WatchGame move event recv: %v", err)
	}
	if moveEvent.GetType() != xov1.GameEventType_GAME_EVENT_TYPE_MOVE_MADE {
		t.Fatalf("expected move made event, got %v", moveEvent.GetType())
	}
	if moveEvent.GetState().GetBoard()[4] != xov1.Mark_MARK_X {
		t.Fatalf("expected streamed cell 4 to be X, got %v", moveEvent.GetState().GetBoard()[4])
	}
	if moveEvent.GetState().GetNextTurn() != xov1.Mark_MARK_O {
		t.Fatalf("expected next turn O, got %v", moveEvent.GetState().GetNextTurn())
	}
}

func TestWatchGameReceivesGameOverEvent(t *testing.T) {
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

	gameID := createResp.GetState().GetGameId()
	aliceToken := createResp.GetPlayerToken()
	joinCode := createResp.GetState().GetJoinCode()

	stream, err := gameClient.WatchGame(ctx, &xov1.WatchGameRequest{
		GameId:      gameID,
		PlayerToken: aliceToken,
	})
	if err != nil {
		t.Fatalf("WatchGame: %v", err)
	}

	snapshot, err := stream.Recv()
	if err != nil {
		t.Fatalf("snapshot recv: %v", err)
	}
	if snapshot.GetType() != xov1.GameEventType_GAME_EVENT_TYPE_STATE_SNAPSHOT {
		t.Fatalf("expected snapshot, got %v", snapshot.GetType())
	}

	joinResp, err := lobbyClient.JoinGame(ctx, &xov1.JoinGameRequest{
		JoinCode:    joinCode,
		DisplayName: "Bob",
	})
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}

	bobToken := joinResp.GetPlayerToken()

	playerJoined, err := stream.Recv()
	if err != nil {
		t.Fatalf("player joined recv: %v", err)
	}
	if playerJoined.GetType() != xov1.GameEventType_GAME_EVENT_TYPE_PLAYER_JOINED {
		t.Fatalf("expected player joined, got %v", playerJoined.GetType())
	}

	moves := []struct {
		token string
		cell  int32
	}{
		{aliceToken, 0}, // X
		{bobToken, 3},   // O
		{aliceToken, 1}, // X
		{bobToken, 4},   // O
		{aliceToken, 2}, // X wins top row
	}

	var lastEvent *xov1.GameEvent

	for i, move := range moves {
		_, err := gameClient.MakeMove(ctx, &xov1.MakeMoveRequest{
			GameId:      gameID,
			PlayerToken: move.token,
			CellIndex:   move.cell,
		})
		if err != nil {
			t.Fatalf("MakeMove #%d: %v", i+1, err)
		}

		lastEvent, err = stream.Recv()
		if err != nil {
			t.Fatalf("stream recv after move #%d: %v", i+1, err)
		}
	}

	if lastEvent.GetType() != xov1.GameEventType_GAME_EVENT_TYPE_GAME_OVER {
		t.Fatalf("expected GAME_OVER event, got %v", lastEvent.GetType())
	}
	if lastEvent.GetGameOverReason() != xov1.GameOverReason_GAME_OVER_REASON_WIN {
		t.Fatalf("expected WIN reason, got %v", lastEvent.GetGameOverReason())
	}
	if lastEvent.GetState().GetWinner() != xov1.Mark_MARK_X {
		t.Fatalf("expected winner X, got %v", lastEvent.GetState().GetWinner())
	}
	if lastEvent.GetState().GetStatus() != xov1.GameStatus_GAME_STATUS_FINISHED {
		t.Fatalf("expected finished status, got %v", lastEvent.GetState().GetStatus())
	}
}