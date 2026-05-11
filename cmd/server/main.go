package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	"github.com/MehrshadFb/xo-grpc/internal/config"
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
	cfg := config.Load()

	// Infrastructure
	store := memory.NewStore()
	sessions := session.NewManager()
	hub := realtime.NewHub()

	// Services
	lobbyService := lobby.NewService(store, sessions, hub)
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
	lis, err := net.Listen("tcp", cfg.Address())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	serverErr := make(chan error, 1) // channel to receive server errors

	// start the server in a goroutine
	go func() {
		log.Printf("gRPC server listening on %s", cfg.Address())
		serverErr <- grpcServer.Serve(lis)
	}()

	shutdown := make(chan os.Signal, 1)                      // channel to receive shutdown signals
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM) // notify when SIGINT or SIGTERM is received

	select {
	case err := <-serverErr: // receive server error
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}

	case sig := <-shutdown: // receive shutdown signal
		log.Printf("received shutdown signal: %s", sig)
		log.Println("gracefully stopping gRPC server")

		grpcServer.GracefulStop() // gracefully stop the server

		log.Println("server stopped")
	}
}
