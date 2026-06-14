package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CookieName is the name of the session cookie.
const CookieName = "polla_session"

// Claims is the payload of our session JWT. For a player session PlayerID is
// set and Epoch is matched against the player's session_epoch (logout bumps it).
// For the master admin session IsAdmin is true and PlayerID is 0.
type Claims struct {
	PlayerID int64 `json:"pid"`
	IsAdmin  bool  `json:"adm"`
	Epoch    int   `json:"epoch"`
	jwt.RegisteredClaims
}

// Sessions issues and parses session JWTs signed with HMAC-SHA256.
type Sessions struct {
	secret []byte
	ttl    time.Duration
}

// NewSessions builds a session signer/verifier.
func NewSessions(secret string, ttl time.Duration) *Sessions {
	return &Sessions{secret: []byte(secret), ttl: ttl}
}

// TTL is how long an issued session lasts.
func (s *Sessions) TTL() time.Duration { return s.ttl }

// Issue mints a signed session token.
func (s *Sessions) Issue(playerID int64, isAdmin bool, epoch int) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.ttl)
	claims := Claims{
		PlayerID: playerID,
		IsAdmin:  isAdmin,
		Epoch:    epoch,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign session: %w", err)
	}
	return signed, expiresAt, nil
}

// Parse validates a token string and returns its claims.
func (s *Sessions) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse session: %w", err)
	}
	return claims, nil
}
