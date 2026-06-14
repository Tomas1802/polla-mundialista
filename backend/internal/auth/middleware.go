package auth

import (
	"context"
	"errors"
	"net/http"

	"polla/internal/db"
)

type ctxKey int

const identityCtxKey ctxKey = iota

// Identity is the authenticated subject: either a player or the master admin.
type Identity struct {
	PlayerID int64
	IsAdmin  bool
}

// Authenticator validates session cookies and loads the current identity.
type Authenticator struct {
	Sessions *Sessions
	store    *db.DB
}

// NewAuthenticator wires the session verifier and the player store together.
func NewAuthenticator(sessions *Sessions, store *db.DB) *Authenticator {
	return &Authenticator{Sessions: sessions, store: store}
}

// Middleware rejects requests without a valid session and attaches the identity.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := a.authenticate(r)
		if err != nil {
			http.Error(w, "no autorizado", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), identityCtxKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Authenticator) authenticate(r *http.Request) (Identity, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return Identity{}, err
	}
	claims, err := a.Sessions.Parse(cookie.Value)
	if err != nil {
		return Identity{}, err
	}
	if claims.IsAdmin && claims.PlayerID == 0 {
		return Identity{IsAdmin: true}, nil
	}
	player, err := a.store.GetPlayer(r.Context(), claims.PlayerID)
	if err != nil {
		return Identity{}, err
	}
	if claims.Epoch != player.SessionEpoch {
		return Identity{}, errors.New("session invalidated")
	}
	return Identity{PlayerID: player.ID}, nil
}

// IdentityFrom returns the authenticated identity attached by Middleware.
func IdentityFrom(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityCtxKey).(Identity)
	return id, ok
}
