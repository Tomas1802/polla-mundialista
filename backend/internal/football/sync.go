package football

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"polla/internal/db"
	"polla/internal/model"
)

// Service keeps the local match cache fresh while making as few API calls as
// possible, per the requirement: one sync per day, plus one extra sync once the
// next match's kickoff has passed (to capture its result and advance to the
// following match).
type Service struct {
	client *Client
	store  *db.DB
	log    *slog.Logger
}

// NewService wires the API client and the database store together.
func NewService(client *Client, store *db.DB, log *slog.Logger) *Service {
	return &Service{client: client, store: store, log: log}
}

// FullSync fetches every match once, caches fixtures + results, and recomputes
// which match is next. This is the only method that calls the API.
func (s *Service) FullSync(ctx context.Context) error {
	matches, teams, err := s.client.Matches(ctx)
	if err != nil {
		return err
	}

	// Chronological order, then assign a stable sequence index.
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].UTCDate.Equal(matches[j].UTCDate) {
			return matches[i].ID < matches[j].ID
		}
		return matches[i].UTCDate.Before(matches[j].UTCDate)
	})
	for i := range matches {
		matches[i].Seq = i
	}

	if err := s.store.UpsertTeams(ctx, teams); err != nil {
		return err
	}
	if err := s.store.UpsertMatches(ctx, matches); err != nil {
		return err
	}

	now := time.Now().UTC()
	state := db.SyncState{LastFullSyncAt: &now, NextMatchSynced: true}
	if next, ok := nextUpcoming(matches, now); ok {
		state.NextMatchUTC = &next.UTCDate
		id := next.ID
		state.NextMatchID = &id
	}
	if err := s.store.SaveSyncState(ctx, state); err != nil {
		return err
	}
	s.log.Info("football full sync done", "matches", len(matches), "teams", len(teams))
	return nil
}

// Tick decides whether a sync is due and performs at most one API call.
func (s *Service) Tick(ctx context.Context) error {
	state, err := s.store.GetSyncState(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()

	// Daily refresh (also covers the very first run).
	if state.LastFullSyncAt == nil || now.Sub(*state.LastFullSyncAt) >= 24*time.Hour {
		return s.FullSync(ctx)
	}
	// The next match has kicked off: pull its result and advance.
	if state.NextMatchUTC != nil && !now.Before(*state.NextMatchUTC) {
		return s.FullSync(ctx)
	}
	// Between stages the next-match pointer is empty: when a phase ends the API
	// has no upcoming fixture yet (e.g. the knockout bracket is published only
	// once the groups finish). Re-sync on the short gap so the new round's
	// fixtures appear promptly, bounded to a few days past the last match so a
	// finished tournament doesn't poll forever (the daily sync still covers it).
	if state.NextMatchUTC == nil && now.Sub(*state.LastFullSyncAt) >= minPollGap {
		if pending, err := s.store.AwaitingNextStage(ctx, now, 3*24*time.Hour); err == nil && pending {
			return s.FullSync(ctx)
		}
	}
	// Any match whose kickoff has passed but isn't FINISHED yet (e.g. the service
	// was asleep when it ended): keep polling until its result lands, regardless
	// of how long ago it kicked off. Bounded by minPollGap to limit API calls.
	if now.Sub(*state.LastFullSyncAt) >= minPollGap {
		if unsettled, err := s.store.HasUnsettledMatches(ctx, now); err == nil && unsettled {
			return s.FullSync(ctx)
		}
	}
	return nil
}

// minPollGap is the shortest interval between syncs while matches are live.
const minPollGap = 5 * time.Minute

// Run performs an initial Tick and then re-checks on the given interval until
// the context is cancelled. The interval only governs how often we *check*; the
// number of actual API calls is bounded by the policy in Tick.
func (s *Service) Run(ctx context.Context, interval time.Duration) {
	if err := s.Tick(ctx); err != nil {
		s.log.Error("initial football sync failed", "err", err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Tick(ctx); err != nil {
				s.log.Error("football sync tick failed", "err", err)
			}
		}
	}
}

// nextUpcoming returns the earliest match whose kickoff is still in the future.
func nextUpcoming(matches []model.Match, now time.Time) (model.Match, bool) {
	for _, m := range matches {
		if m.UTCDate.After(now) {
			return m, true
		}
	}
	return model.Match{}, false
}
