// Unsynced marcadores are mirrored to localStorage the instant a user edits
// them, so a flaky mobile network or a page reload can never lose a prediction.
// An entry is removed only once the server confirms the save; a surviving entry
// on load means "edited but never synced" — the UI shows it and lets the player
// confirm/retry. This is the safety net behind the optimistic auto-save.
const KEY = 'polla_pending_predictions'

function readAll() {
  try {
    return JSON.parse(localStorage.getItem(KEY) || '{}')
  } catch {
    return {}
  }
}

function writeAll(map) {
  try {
    localStorage.setItem(KEY, JSON.stringify(map))
  } catch {
    // ignore storage errors (private mode, quota exceeded)
  }
}

const k = (cardId, matchId) => `${cardId}:${matchId}`

export function getPending(cardId, matchId) {
  return readAll()[k(cardId, matchId)] || null
}

export function setPending(cardId, matchId, draft) {
  const map = readAll()
  map[k(cardId, matchId)] = draft
  writeAll(map)
}

export function clearPending(cardId, matchId) {
  const map = readAll()
  delete map[k(cardId, matchId)]
  writeAll(map)
}
