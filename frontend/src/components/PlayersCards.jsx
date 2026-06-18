import { useEffect, useState } from 'react'
import { api } from '../api.js'

// Admin view: each player is a collapsible block; inside, each of their cards
// is collapsible; opening a card lazily loads its matches and lets the admin
// edit any marcador (including already-played ones).
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
        Abre un jugador, luego un cartón, y corrige cualquier marcador (incluidos los ya jugados).
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
        {matches && matches.map((m) => <FixRow key={m.id} cardId={card.id} match={m} />)}
      </div>
    </details>
  )
}

function FixRow({ cardId, match }) {
  const [home, setHome] = useState(match.prediction?.home ?? '')
  const [away, setAway] = useState(match.prediction?.away ?? '')
  const [status, setStatus] = useState('')

  async function save() {
    setStatus('saving')
    try {
      await api.adminEditPrediction(cardId, match.id, {
        home: home === '' ? null : Number(home),
        away: away === '' ? null : Number(away),
        penaltyWinner: '',
      })
      setStatus('saved')
      setTimeout(() => setStatus((s) => (s === 'saved' ? '' : s)), 1500)
    } catch (e) {
      setStatus(e.message)
    }
  }

  return (
    <div className="fix-row">
      <span className="fix-teams">{match.homeTeamName} <span className="vs">vs</span> {match.awayTeamName}</span>
      <input className="admin-score" type="number" min="0" max="99" value={home}
        onChange={(e) => setHome(e.target.value)} />
      <span className="admin-dash">–</span>
      <input className="admin-score" type="number" min="0" max="99" value={away}
        onChange={(e) => setAway(e.target.value)} />
      <button className="fix-save" onClick={save} disabled={status === 'saving'}>Guardar</button>
      {match.status === 'FINISHED' && (
        <span className="fix-real">real {match.scoreHome}–{match.scoreAway}</span>
      )}
      {status === 'saved' && <span className="save-status save-saved">✓</span>}
      {status && status !== 'saving' && status !== 'saved' && (
        <span className="save-status save-error">{status}</span>
      )}
    </div>
  )
}
