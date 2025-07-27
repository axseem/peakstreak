package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axseem/peakstreak/internal/api"
	"github.com/axseem/peakstreak/internal/config"
	"github.com/axseem/peakstreak/internal/repository"
	"github.com/axseem/peakstreak/internal/service"
	"github.com/axseem/peakstreak/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig(".")
	if err != nil {
		slog.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	dbpool, err := pgxpool.New(context.Background(), cfg.DBUrl)
	if err != nil {
		slog.Error("unable to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	postgresRepo := repository.NewPostgresRepository(dbpool)
	avatarStoragePath := "./uploads/avatars"
	fileStorage := storage.NewLocalStorage(avatarStoragePath, "/uploads/avatars")
	appService := service.New(postgresRepo, fileStorage)
	apiHandler := api.NewAPIHandler(appService, &cfg)
	router := api.NewRouter(apiHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		slog.Info("shutdown signal received")

		shutdownCtx, stopShutdownCtx := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				slog.Error("graceful shutdown timed out.. forcing exit.")
				os.Exit(1)
			}
		}()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown failed", "error", err)
		}
		serverStopCtx()
		stopShutdownCtx()
	}()

	slog.Info("server starting", "port", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}

	<-serverCtx.Done()
	slog.Info("server stopped")
}
