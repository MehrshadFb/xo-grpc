package e2e

import (
	"context"
	"testing"
	"time"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestWatchGameResumeSkipsDuplicateSnapshot(t *testing.T) {
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

	joinResp, err := lobbyClient.JoinGame(ctx, &xov1.JoinGameRequest{
		JoinCode:    joinCode,
		DisplayName: "Bob",
	})
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}

	bobToken := joinResp.GetPlayerToken()

	moveResp, err := gameClient.MakeMove(ctx, &xov1.MakeMoveRequest{
		GameId:      gameID,
		PlayerToken: aliceToken,
		CellIndex:   0,
	})
	if err != nil {
		t.Fatalf("MakeMove first: %v", err)
	}

	currentVersion := moveResp.GetState().GetVersion()

	snapshotStream, err := gameClient.WatchGame(ctx, &xov1.WatchGameRequest{
		GameId:       gameID,
		PlayerToken:  aliceToken,
		AfterVersion: 0,
	})
	if err != nil {
		t.Fatalf("WatchGame snapshot stream: %v", err)
	}

	snapshot, err := snapshotStream.Recv()
	if err != nil {
		t.Fatalf("snapshot recv: %v", err)
	}

	if snapshot.GetType() != xov1.GameEventType_GAME_EVENT_TYPE_STATE_SNAPSHOT {
		t.Fatalf("expected snapshot event, got %v", snapshot.GetType())
	}
	if snapshot.GetState().GetVersion() != currentVersion {
		t.Fatalf("expected snapshot version %d, got %d", currentVersion, snapshot.GetState().GetVersion())
	}

	resumeCtx, resumeCancel := context.WithCancel(ctx)
	defer resumeCancel()

	resumeStream, err := gameClient.WatchGame(resumeCtx, &xov1.WatchGameRequest{
		GameId:       gameID,
		PlayerToken:  aliceToken,
		AfterVersion: currentVersion,
	})
	if err != nil {
		t.Fatalf("WatchGame resume stream: %v", err)
	}

	eventCh := make(chan *xov1.GameEvent, 1)
	errCh := make(chan error, 1)

	go func() {
		event, err := resumeStream.Recv()
		if err != nil {
			errCh <- err
			return
		}

		eventCh <- event
	}()

	select {
	case event := <-eventCh:
		t.Fatalf("expected no immediate event on resume, got %v", event.GetType())

	case err := <-errCh:
		t.Fatalf("unexpected resume recv error before next move: %v", err)

	case <-time.After(200 * time.Millisecond):
		// Expected: no duplicate snapshot.
	}

	_, err = gameClient.MakeMove(ctx, &xov1.MakeMoveRequest{
		GameId:      gameID,
		PlayerToken: bobToken,
		CellIndex:   4,
	})
	if err != nil {
		t.Fatalf("MakeMove second: %v", err)
	}

	select {
	case event := <-eventCh:
		if event.GetType() != xov1.GameEventType_GAME_EVENT_TYPE_MOVE_MADE {
			t.Fatalf("expected move made event, got %v", event.GetType())
		}
		if event.GetState().GetBoard()[4] != xov1.Mark_MARK_O {
			t.Fatalf("expected streamed cell 4 to be O, got %v", event.GetState().GetBoard()[4])
		}

	case err := <-errCh:
		t.Fatalf("resume stream recv error after move: %v", err)

	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for resumed stream event")
	}
}
