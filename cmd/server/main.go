package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	xov1 "github.com/MehrshadFb/xo-grpc/gen/go/xo/v1"
	"github.com/MehrshadFb/xo-grpc/internal/config"
	"github.com/MehrshadFb/xo-grpc/internal/database"
	"github.com/MehrshadFb/xo-grpc/internal/realtime"
	"github.com/MehrshadFb/xo-grpc/internal/repository"
	gamesvc "github.com/MehrshadFb/xo-grpc/internal/service/game"
	healthsvc "github.com/MehrshadFb/xo-grpc/internal/service/health"
	"github.com/MehrshadFb/xo-grpc/internal/service/lobby"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
	postgresstore "github.com/MehrshadFb/xo-grpc/internal/store/postgres"
	transportgrpc "github.com/MehrshadFb/xo-grpc/internal/transport/grpc"
	"github.com/jackc/pgx/v5/pgxpool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// Infrastructure
	memStore := memory.NewStore()
	memorySessionRepo := memory.NewSessionRepository()
	hub := realtime.NewHub()

	var (
		gameRepo    repository.GameRepository
		sessionRepo repository.SessionRepository
	)

	// Default to in-memory repositories
	gameRepo = memStore
	sessionRepo = memorySessionRepo

	// Database
	var (
		dbPool *pgxpool.Pool
		err    error
	)
	if cfg.DatabaseURL != "" {
		dbPool, err = database.NewPostgresPool(ctx, cfg.DatabaseURL)
		if err != nil {
			slog.Error("failed to connect to postgres", "error", err)
			os.Exit(1)
		}
		defer dbPool.Close()

		gameRepo = postgresstore.NewGameRepository(dbPool)
		sessionRepo = postgresstore.NewSessionRepository(dbPool)
		slog.Info("connected to postgres; using postgres repositories")
	} else {
		slog.Info("database not configured; using in-memory repositories")
	}

	// Session manager
	sessions := session.NewManager(sessionRepo)

	// Services
	lobbyService := lobby.NewService(gameRepo, sessions, hub)
	gameService := gamesvc.NewService(gameRepo, sessions, hub)
	healthService := healthsvc.NewService(dbPool)

	// gRPC handlers
	lobbyHandler := transportgrpc.NewLobbyHandler(lobbyService)
	gameHandler := transportgrpc.NewGameHandler(gameService, hub)
	healthHandler := transportgrpc.NewHealthHandler(healthService)

	// gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(transportgrpc.LoggingUnaryInterceptor),
		grpc.StreamInterceptor(transportgrpc.LoggingStreamInterceptor),
	)

	// Enable grpcurl reflection support and register services
	reflection.Register(grpcServer)
	xov1.RegisterLobbyServiceServer(grpcServer, lobbyHandler)
	xov1.RegisterGameServiceServer(grpcServer, gameHandler)
	xov1.RegisterHealthServiceServer(grpcServer, healthHandler)

	// Listen
	lis, err := net.Listen("tcp", cfg.Address())
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	serverErr := make(chan error, 1) // channel to receive server errors

	// start the server in a goroutine
	go func() {
		slog.Info("gRPC server listening", "addr", cfg.Address())
		serverErr <- grpcServer.Serve(lis)
	}()

	shutdown := make(chan os.Signal, 1)                      // channel to receive shutdown signals
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM) // notify when SIGINT or SIGTERM is received

	select {
	case err := <-serverErr: // receive server error
		if err != nil {
			slog.Error("failed to serve", "error", err)
			os.Exit(1)
		}

	case sig := <-shutdown: // receive shutdown signal
		slog.Info("received shutdown signal", "signal", sig.String())
		slog.Info("gracefully stopping gRPC server")

		grpcServer.GracefulStop() // gracefully stop the server

		slog.Info("server stopped")
	}
}
