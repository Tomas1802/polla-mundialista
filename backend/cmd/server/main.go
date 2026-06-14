// Command server is the Polla Mundialista 2026 backend: it migrates the
// database, starts the background football-data sync, and serves the REST API.
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

	"polla/internal/api"
	"polla/internal/auth"
	"polla/internal/config"
	"polla/internal/db"
	"polla/internal/football"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	if err := db.Migrate(ctx, database.Pool); err != nil {
		return err
	}
	log.Info("database migrated")

	sessions := auth.NewSessions(cfg.JWTSecret, cfg.SessionTTL)

	// Background football-data sync, only when a token is configured.
	if cfg.FootballDataToken != "" {
		client := football.NewClient(cfg.FootballDataToken, cfg.CompetitionCode)
		syncer := football.NewService(client, database, log)
		go syncer.Run(ctx, time.Duration(cfg.SyncIntervalMinutes)*time.Minute)
	} else {
		log.Warn("FOOTBALL_DATA_TOKEN not set; match sync disabled")
	}

	server := api.NewServer(cfg, database, sessions, log)
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("graceful shutdown failed", "err", err)
		}
	}()

	log.Info("server listening", "port", cfg.Port)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("server stopped")
	return nil
}
