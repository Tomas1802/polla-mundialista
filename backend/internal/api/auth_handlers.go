package api

import (
	"encoding/json"
	"net/http"

	"polla/internal/auth"
)

// sessionDTO is the shape returned by /login and /me.
type sessionDTO struct {
	IsAdmin       bool   `json:"isAdmin"`
	PlayerID      *int64 `json:"playerId"`
	PlayerName    string `json:"playerName"`
	MustChangePin bool   `json:"mustChangePin"`
}

type loginRequest struct {
	PlayerID int64  `json:"playerId"`
	Pin      string `json:"pin"`
}

// handleLogin authenticates a player by id + PIN.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlayerID == 0 {
		writeError(w, http.StatusBadRequest, "Selecciona tu nombre y escribe tu PIN.")
		return
	}
	player, err := s.store.GetPlayer(r.Context(), req.PlayerID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Jugador o PIN incorrecto.")
		return
	}
	if player.PinHash == "" {
		writeError(w, http.StatusUnauthorized, "Aún no tienes un PIN asignado. Pídeselo al organizador.")
		return
	}
	if !auth.CheckPin(player.PinHash, req.Pin) {
		writeError(w, http.StatusUnauthorized, "Jugador o PIN incorrecto.")
		return
	}

	token, _, err := s.sessions.Issue(player.ID, false, player.SessionEpoch)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos crear tu sesión.")
		return
	}
	s.setSessionCookie(w, token)
	id := player.ID
	writeJSON(w, http.StatusOK, sessionDTO{
		PlayerID: &id, PlayerName: player.Name, MustChangePin: player.MustChangePin,
	})
}

type adminLoginRequest struct {
	Pin string `json:"pin"`
}

// handleAdminLogin authenticates the master admin via the configured PIN.
func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if s.cfg.AdminPin == "" {
		writeError(w, http.StatusForbidden, "El acceso de administrador no está configurado.")
		return
	}
	var req adminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Solicitud inválida.")
		return
	}
	if !auth.ConstantTimeEqual(req.Pin, s.cfg.AdminPin) {
		writeError(w, http.StatusUnauthorized, "PIN de administrador incorrecto.")
		return
	}
	token, _, err := s.sessions.Issue(0, true, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos crear tu sesión.")
		return
	}
	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, sessionDTO{IsAdmin: true})
}

type changePinRequest struct {
	NewPin string `json:"newPin"`
}

// handleChangePin lets a logged-in player set a new PIN (required on first login).
func (s *Server) handleChangePin(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.IdentityFrom(r.Context())
	if id.PlayerID == 0 {
		writeError(w, http.StatusForbidden, "Solo los jugadores cambian PIN.")
		return
	}
	var req changePinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Solicitud inválida.")
		return
	}
	if !auth.ValidPin(req.NewPin) {
		writeError(w, http.StatusBadRequest, "El PIN debe tener 4 dígitos.")
		return
	}
	hash, err := auth.HashPin(req.NewPin)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos guardar el PIN.")
		return
	}
	if err := s.store.ChangePin(r.Context(), id.PlayerID, hash); err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos guardar el PIN.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleLogout invalidates the player's sessions and clears the cookie.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.CookieName); err == nil {
		if claims, err := s.sessions.Parse(cookie.Value); err == nil && claims.PlayerID != 0 {
			if player, err := s.store.GetPlayer(r.Context(), claims.PlayerID); err == nil && claims.Epoch == player.SessionEpoch {
				_, _ = s.store.IncrementPlayerEpoch(r.Context(), player.ID)
			}
		}
	}
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleMe returns the current session.
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.IdentityFrom(r.Context())
	if id.IsAdmin {
		writeJSON(w, http.StatusOK, sessionDTO{IsAdmin: true})
		return
	}
	player, err := s.store.GetPlayer(r.Context(), id.PlayerID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Sesión inválida.")
		return
	}
	pid := player.ID
	writeJSON(w, http.StatusOK, sessionDTO{
		PlayerID: &pid, PlayerName: player.Name, MustChangePin: player.MustChangePin,
	})
}
