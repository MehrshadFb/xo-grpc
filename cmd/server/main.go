package main

import (
	"log"
	"net"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	"github.com/MehrshadFb/xo-grpc/internal/realtime"
	gamesvc "github.com/MehrshadFb/xo-grpc/internal/service/game"
	"github.com/MehrshadFb/xo-grpc/internal/service/lobby"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
	transportgrpc "github.com/MehrshadFb/xo-grpc/internal/transport/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Infrastructure
	store := memory.NewStore()
	sessions := session.NewManager()
	hub := realtime.NewHub()

	// Services
	lobbyService := lobby.NewService(store, sessions)
	gameService := gamesvc.NewService(store, sessions, hub)

	// gRPC handlers
	lobbyHandler := transportgrpc.NewLobbyHandler(lobbyService)
	gameHandler := transportgrpc.NewGameHandler(gameService, hub)

	// gRPC server
	grpcServer := grpc.NewServer()

	// Enable grpcurl reflection support and register services
	reflection.Register(grpcServer)
	xov1.RegisterLobbyServiceServer(grpcServer, lobbyHandler)
	xov1.RegisterGameServiceServer(grpcServer, gameHandler)

	// Listen
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server listening on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}