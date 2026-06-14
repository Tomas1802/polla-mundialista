// Command genpins assigns a fresh random 4-digit PIN to every player and writes
// a CSV (Jugador,PIN) for the organizer to hand out. The plaintext PINs live
// ONLY in that CSV; the database stores just the bcrypt hash, and each player is
// forced to change their PIN on first login.
//
// Usage (from the backend/ folder, with the DB running and cartones imported):
//
//	go run ./cmd/genpins            # writes pins.csv
//	go run ./cmd/genpins out.csv    # writes out.csv
//
// WARNING: this resets EVERY player's PIN. Run it once at setup (or again only
// if you intend to reissue all PINs).
package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"polla/internal/auth"
	"polla/internal/config"
	"polla/internal/db"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	players, err := database.ListPlayers(ctx)
	if err != nil {
		return err
	}
	if len(players) == 0 {
		return fmt.Errorf("no hay jugadores; importa los cartones primero (pestaña Admin)")
	}

	outPath := "pins.csv"
	if len(os.Args) > 1 {
		outPath = os.Args[1]
	}
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("crear %s: %w", outPath, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"Jugador", "PIN"}); err != nil {
		return err
	}

	for _, p := range players {
		pin, err := auth.GeneratePin()
		if err != nil {
			return err
		}
		hash, err := auth.HashPin(pin)
		if err != nil {
			return err
		}
		if err := database.SetPlayerPin(ctx, p.ID, hash); err != nil {
			return fmt.Errorf("guardar PIN de %s: %w", p.Name, err)
		}
		if err := w.Write([]string{p.Name, pin}); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	abs, _ := filepath.Abs(outPath)
	fmt.Printf("Generados %d PINs.\nArchivo para entregar al admin: %s\n", len(players), abs)
	return nil
}
