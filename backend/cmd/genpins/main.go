// Command genpins assigns a random 4-digit PIN to every player that does NOT
// already have one, and writes the newly-generated ones to a CSV (Jugador,PIN)
// for the organizer to hand out. Players who already have a PIN are left
// untouched, so it is safe to re-run when new players are added.
//
// The plaintext PINs live ONLY in that CSV (and stdout); the database stores
// just the bcrypt hash, and each player must change their PIN on first login.
//
// Usage (from the backend/ folder, with DATABASE_URL set to the target DB):
//
//	go run ./cmd/genpins                 # writes pins_nuevos.csv
//	go run ./cmd/genpins salida.csv      # writes salida.csv
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

	type pinned struct{ name, pin string }
	var generated []pinned
	for _, p := range players {
		full, err := database.GetPlayer(ctx, p.ID)
		if err != nil {
			return err
		}
		if full.PinHash != "" {
			continue // already has a PIN — leave it untouched
		}
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
		generated = append(generated, pinned{p.Name, pin})
		fmt.Printf("PIN nuevo: %s = %s\n", p.Name, pin)
	}

	if len(generated) == 0 {
		fmt.Println("Todos los jugadores ya tienen PIN; nada que generar.")
		return nil
	}

	outPath := "pins_nuevos.csv"
	if len(os.Args) > 1 {
		outPath = os.Args[1]
	}
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("crear %s: %w", outPath, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	_ = w.Write([]string{"Jugador", "PIN"})
	for _, g := range generated {
		if err := w.Write([]string{g.name, g.pin}); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	abs, _ := filepath.Abs(outPath)
	fmt.Printf("Generados %d PIN(s) nuevo(s). Archivo: %s\n", len(generated), abs)
	return nil
}
