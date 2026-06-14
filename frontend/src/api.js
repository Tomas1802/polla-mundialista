// Thin fetch wrapper. credentials:'include' sends the session cookie on every
// request so the backend keeps the session across visits.
const BASE = import.meta.env.VITE_API_BASE || ''

async function request(path, options = {}) {
  const res = await fetch(BASE + path, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    let message = 'Ocurrió un error. Intenta de nuevo.'
    try {
      const body = await res.json()
      if (body && body.error) message = body.error
    } catch {
      // non-JSON error body; keep the default message
    }
    const err = new Error(message)
    err.status = res.status
    throw err
  }
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  me: () => request('/api/me'),
  players: () => request('/api/players'),
  login: (playerId, pin) =>
    request('/api/auth/login', { method: 'POST', body: JSON.stringify({ playerId, pin }) }),
  adminLogin: (pin) =>
    request('/api/auth/admin-login', { method: 'POST', body: JSON.stringify({ pin }) }),
  changePin: (newPin) =>
    request('/api/auth/change-pin', { method: 'POST', body: JSON.stringify({ newPin }) }),
  logout: () => request('/api/auth/logout', { method: 'POST' }),

  cards: () => request('/api/cards'),
  matches: (cardId) => request(`/api/matches?cardId=${cardId}`),
  putPrediction: (cardId, matchId, body) =>
    request(`/api/cards/${cardId}/predictions/${matchId}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  ranking: () => request('/api/ranking'),
  tables: (cardId) => request(`/api/tables?cardId=${cardId}`),

  adminPlayers: () => request('/api/admin/players'),
  adminImport: () => request('/api/admin/import-scores', { method: 'POST' }),
}
