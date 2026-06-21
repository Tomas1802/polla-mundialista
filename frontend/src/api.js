// Thin fetch wrapper. The session token is kept in localStorage and sent as a
// Bearer header (cookies are blocked cross-site by iOS Safari).
const BASE = import.meta.env.VITE_API_BASE || ''
const TOKEN_KEY = 'polla_token'

export function getToken() {
  try {
    return localStorage.getItem(TOKEN_KEY) || ''
  } catch {
    return ''
  }
}

export function setToken(token) {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token)
    else localStorage.removeItem(TOKEN_KEY)
  } catch {
    // ignore storage errors (private mode, etc.)
  }
}

async function request(path, options = {}) {
  const url = BASE + path
  const headers = { 'Content-Type': 'application/json', ...(options.headers || {}) }
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`
  let res
  try {
    res = await fetch(url, {
      credentials: 'include',
      ...options,
      headers,
    })
  } catch (e) {
    // Network / CORS / DNS failure (fetch threw before any response).
    throw new Error(`No se pudo conectar a ${url} — ${e.name}: ${e.message}`)
  }
  if (!res.ok) {
    let bodyText = ''
    try {
      bodyText = await res.text()
    } catch {
      // ignore: no readable body
    }
    let message = ''
    try {
      const body = JSON.parse(bodyText)
      if (body && body.error) message = body.error
    } catch {
      // body was not JSON
    }
    if (!message) {
      const snippet = bodyText ? ` — ${bodyText.slice(0, 300)}` : ''
      message = `HTTP ${res.status} ${res.statusText} en ${url}${snippet}`
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
  login: async (playerId, pin) => {
    const data = await request('/api/auth/login', { method: 'POST', body: JSON.stringify({ playerId, pin }) })
    if (data && data.token) setToken(data.token)
    return data
  },
  adminLogin: async (pin) => {
    const data = await request('/api/auth/admin-login', { method: 'POST', body: JSON.stringify({ pin }) })
    if (data && data.token) setToken(data.token)
    return data
  },
  changePin: (newPin) =>
    request('/api/auth/change-pin', { method: 'POST', body: JSON.stringify({ newPin }) }),
  logout: async () => {
    try {
      await request('/api/auth/logout', { method: 'POST' })
    } finally {
      setToken('')
    }
  },

  cards: () => request('/api/cards'),
  matches: (cardId) => request(`/api/matches?cardId=${cardId}`),
  putPrediction: (cardId, matchId, body) =>
    request(`/api/cards/${cardId}/predictions/${matchId}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  ranking: () => request('/api/ranking'),
  scorecard: (cardId) => request(`/api/cards/${cardId}/scorecard`),
  teams: () => request('/api/teams'),
  tables: (cardId) => request(`/api/tables?cardId=${cardId}`),

  adminPlayers: () => request('/api/admin/players'),
  adminMissingToday: () => request('/api/admin/missing-today'),
  adminMatches: () => request('/api/admin/matches'),
  adminSetResult: (matchId, body) =>
    request(`/api/admin/matches/${matchId}/result`, { method: 'PUT', body: JSON.stringify(body) }),
  adminConfirmResult: (matchId) =>
    request(`/api/admin/matches/${matchId}/confirm`, { method: 'POST' }),
  adminClearResult: (matchId) =>
    request(`/api/admin/matches/${matchId}/result`, { method: 'DELETE' }),
  adminSettings: () => request('/api/admin/settings'),
  adminSetSettings: (matchId) =>
    request('/api/admin/settings', { method: 'PUT', body: JSON.stringify({ editLockUntilMatchId: matchId }) }),
  adminCards: () => request('/api/admin/cards'),
  adminEditPrediction: (cardId, matchId, body) =>
    request(`/api/admin/cards/${cardId}/predictions/${matchId}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  adminImport: () => request('/api/admin/import-scores', { method: 'POST' }),
}
