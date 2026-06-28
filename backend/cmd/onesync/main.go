// Command onesync runs a single football-data FullSync against the configured
// database, then exits. Throwaway tool to force-refresh the cache out of band.
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"polla/internal/config"
	"polla/internal/db"
	"polla/internal/football"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}
	if cfg.FootballDataToken == "" {
		log.Error("FOOTBALL_DATA_TOKEN not set")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("connect", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	client := football.NewClient(cfg.FootballDataToken, cfg.CompetitionCode)
	syncer := football.NewService(client, database, log)
	if err := syncer.FullSync(ctx); err != nil {
		log.Error("full sync", "err", err)
		os.Exit(1)
	}
	log.Info("one-off full sync complete")
}
