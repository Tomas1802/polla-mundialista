// Package scores imports the cartón CSV files into players, cards and card
// predictions. Each cartón CSV has rows "numero,local,visitante" where numero
// is the official FIFA match number (1..72); schedule.csv maps that number to
// the football-data match id, so predictions land on the correct match
// regardless of chronological/timezone ordering.
//
// File name format: GRUPOS_Fase1_<PlayerName>[<cardNo>].csv
//   - the player name is the token after the last underscore
//   - an optional trailing digit is the card number (1..3); none means card 1
package scores

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"polla/internal/db"
	"polla/internal/model"
)

// Result reports what the import did.
type Result struct {
	Players     int `json:"players"`
	Cards       int `json:"cards"`
	Predictions int `json:"predictions"`
}

const scheduleFile = "schedule.csv"

// Import reads schedule.csv plus every GRUPOS_*.csv in dir and upserts the
// corresponding players, cards and predictions (mapped by FIFA match number).
func Import(ctx context.Context, store *db.DB, dir string) (Result, error) {
	schedule, err := loadSchedule(filepath.Join(dir, scheduleFile))
	if err != nil {
		return Result{}, err
	}

	// Clear existing predictions for the scheduled (group-stage) matches so a
	// re-import cleanly replaces them instead of leaving stale rows.
	matchIDs := make([]int64, 0, len(schedule))
	for _, id := range schedule {
		matchIDs = append(matchIDs, id)
	}
	if err := store.DeleteCardPredictionsForMatches(ctx, matchIDs); err != nil {
		return Result{}, fmt.Errorf("limpiar pronósticos previos: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return Result{}, fmt.Errorf("no se pudo leer la carpeta %q: %w", dir, err)
	}

	var res Result
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "GRUPOS") || !strings.HasSuffix(strings.ToLower(name), ".csv") {
			continue
		}
		playerName, cardNo := parseName(name)
		if playerName == "" {
			continue
		}
		rows, err := readCardCSV(filepath.Join(dir, name))
		if err != nil {
			return res, fmt.Errorf("leer %s: %w", name, err)
		}

		playerID, err := store.UpsertPlayer(ctx, playerName)
		if err != nil {
			return res, err
		}
		cardID, err := store.UpsertCard(ctx, playerID, cardNo, fmt.Sprintf("Cartón %d", cardNo))
		if err != nil {
			return res, err
		}
		res.Cards++

		for _, row := range rows {
			matchID, ok := schedule[row.number]
			if !ok {
				continue // number outside the group stage / not mapped
			}
			if err := store.UpsertCardPrediction(ctx, model.CardPrediction{
				CardID:  cardID,
				MatchID: matchID,
				Home:    row.home,
				Away:    row.away,
			}); err != nil {
				return res, err
			}
			res.Predictions++
		}
	}

	players, err := store.ListPlayers(ctx)
	if err != nil {
		return res, err
	}
	res.Players = len(players)
	return res, nil
}

// loadSchedule reads "numero,matchId" rows into a map.
func loadSchedule(path string) (map[int]int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir %s (necesario para el orden de partidos): %w", scheduleFile, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("leer %s: %w", scheduleFile, err)
	}
	out := map[int]int64{}
	for _, rec := range records {
		if len(rec) < 2 {
			continue
		}
		num, err1 := strconv.Atoi(strings.TrimSpace(rec[0]))
		id, err2 := strconv.ParseInt(strings.TrimSpace(rec[1]), 10, 64)
		if err1 != nil || err2 != nil {
			continue // header or blank line
		}
		out[num] = id
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s no tiene mapeos válidos", scheduleFile)
	}
	return out, nil
}

type cardRow struct {
	number int
	home   *int
	away   *int
}

// readCardCSV reads "numero,local,visitante" rows; missing values stay nil.
func readCardCSV(path string) ([]cardRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	var out []cardRow
	for _, rec := range records {
		if len(rec) == 0 {
			continue
		}
		num, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil {
			continue // header or blank
		}
		row := cardRow{number: num}
		if len(rec) > 1 {
			if v, e := strconv.Atoi(strings.TrimSpace(rec[1])); e == nil {
				row.home = &v
			}
		}
		if len(rec) > 2 {
			if v, e := strconv.Atoi(strings.TrimSpace(rec[2])); e == nil {
				row.away = &v
			}
		}
		out = append(out, row)
	}
	return out, nil
}

// parseName extracts the player display name and card number from a filename.
func parseName(filename string) (string, int) {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	if i := strings.LastIndex(base, "_"); i >= 0 {
		base = base[i+1:]
	}
	cardNo := 1
	if n := len(base); n > 0 && unicode.IsDigit(rune(base[n-1])) {
		cardNo = int(base[n-1] - '0')
		base = base[:n-1]
	}
	return prettify(base), cardNo
}

// prettify turns "MariaPaulaBuitrago" into "Maria Paula Buitrago".
func prettify(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}
