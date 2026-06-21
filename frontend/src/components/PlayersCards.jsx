import { useEffect, useState } from 'react'
import { api } from '../api.js'

// Admin view (read-only): each player is a collapsible block; inside, each card
// is collapsible; opening a card lazily loads its matches and shows the player's
// marcadores together with the official result. No editing here.
export default function PlayersCards() {
  const [players, setPlayers] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api.adminCards()
      .then((d) => {
        const byPlayer = new Map()
        for (const c of d.cards || []) {
          if (!byPlayer.has(c.playerName)) byPlayer.set(c.playerName, [])
          byPlayer.get(c.playerName).push(c)
        }
        setPlayers([...byPlayer.entries()].map(([name, cards]) => ({ name, cards })))
      })
      .catch((e) => setError(e.message))
  }, [])

  if (error) return <p className="error">{error}</p>
  if (!players) return <div className="spinner" />

  return (
    <section className="admin-section">
      <h2>Jugadores y cartones ({players.length})</h2>
      <p className="tablas-intro">
        Abre un jugador, luego un cartón, para ver sus marcadores y el resultado real (solo lectura).
      </p>
      {players.map((p) => (
        <details key={p.name} className="player-block">
          <summary>
            <span className="player-block-name">{p.name}</span>
            <span className="player-block-count">{p.cards.length} {p.cards.length === 1 ? 'cartón' : 'cartones'}</span>
          </summary>
          <div className="player-block-body">
            {p.cards.map((c) => <CartonBlock key={c.id} card={c} />)}
          </div>
        </details>
      ))}
    </section>
  )
}

function CartonBlock({ card }) {
  const [matches, setMatches] = useState(null)
  const [error, setError] = useState('')

  async function onToggle(e) {
    if (e.target.open && matches === null) {
      try {
        const d = await api.matches(card.id)
        setMatches((d.matches || []).filter((m) => m.stage === 'GROUP_STAGE'))
      } catch (err) {
        setError(err.message)
      }
    }
  }

  return (
    <details className="carton-block" onToggle={onToggle}>
      <summary>{card.label}</summary>
      <div className="carton-block-body">
        {error && <p className="error">{error}</p>}
        {matches === null && !error && <div className="spinner" />}
        {matches && matches.map((m) => <ViewRow key={m.id} match={m} />)}
      </div>
    </details>
  )
}

function ViewRow({ match }) {
  const pred = match.prediction
  const predText = pred && pred.home != null ? `${pred.home}–${pred.away}` : '—'
  const finished = match.status === 'FINISHED'

  return (
    <div className="fix-row">
      <span className="fix-teams">{match.homeTeamName} <span className="vs">vs</span> {match.awayTeamName}</span>
      <span className="fix-pred">Pronóstico: <strong>{predText}</strong></span>
      {finished && (
        <span className="fix-real">real {match.scoreHome}–{match.scoreAway}</span>
      )}
      {finished && match.points != null && <span className="points-badge">+{match.points}</span>}
    </div>
  )
}
