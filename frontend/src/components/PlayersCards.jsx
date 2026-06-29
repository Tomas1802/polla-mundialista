import { useEffect, useState } from 'react'
import { api } from '../api.js'

// Admin view: each player is a collapsible block; inside, each card is
// collapsible and shows the player's marcadores vs the official result. Editing
// is read-only by default and unlocked with a hidden chord (Ctrl+K then G).
export default function PlayersCards() {
  const [players, setPlayers] = useState(null)
  const [error, setError] = useState('')
  const [editMode, setEditMode] = useState(false)

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

  // Hidden unlock: press Ctrl+K, then G (VS Code-style chord) to toggle editing.
  useEffect(() => {
    let armed = false
    let timer = null
    function onKey(e) {
      if (e.ctrlKey && (e.key === 'k' || e.key === 'K')) {
        armed = true
        clearTimeout(timer)
        timer = setTimeout(() => { armed = false }, 1500)
        e.preventDefault()
        return
      }
      if (armed && (e.key === 'g' || e.key === 'G')) {
        armed = false
        clearTimeout(timer)
        e.preventDefault()
        setEditMode((v) => !v)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('keydown', onKey)
      clearTimeout(timer)
    }
  }, [])

  if (error) return <p className="error">{error}</p>
  if (!players) return <div className="spinner" />

  return (
    <section className="admin-section">
      <h2>Jugadores y cartones ({players.length})</h2>
      {editMode ? (
        <p className="edit-banner">
          ✏️ Edición habilitada. Corrige cualquier marcador y guarda.
          <button className="edit-banner-off" onClick={() => setEditMode(false)}>Bloquear</button>
        </p>
      ) : (
        <p className="tablas-intro">
          Abre un jugador, luego un cartón, para ver sus marcadores y el resultado real (solo lectura).
        </p>
      )}
      {players.map((p) => (
        <details key={p.name} className="player-block">
          <summary>
            <span className="player-block-name">{p.name}</span>
            <span className="player-block-count">{p.cards.length} {p.cards.length === 1 ? 'cartón' : 'cartones'}</span>
          </summary>
          <div className="player-block-body">
            {p.cards.map((c) => <CartonBlock key={c.id} card={c} editMode={editMode} />)}
          </div>
        </details>
      ))}
    </section>
  )
}

function CartonBlock({ card, editMode }) {
  const [matches, setMatches] = useState(null)
  const [error, setError] = useState('')

  async function onToggle(e) {
    if (e.target.open && matches === null) {
      try {
        const d = await api.matches(card.id)
        setMatches(d.matches || [])
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
        {matches && matches.map((m) =>
          editMode
            ? <FixRow key={m.id} cardId={card.id} match={m} />
            : <ViewRow key={m.id} match={m} />,
        )}
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

function FixRow({ cardId, match }) {
  const [home, setHome] = useState(match.prediction?.home ?? '')
  const [away, setAway] = useState(match.prediction?.away ?? '')
  const [pen, setPen] = useState(match.prediction?.penaltyWinner ?? '')
  const [status, setStatus] = useState('')

  // Knockout draws need a penalty winner to decide who advances.
  const isDrawPred = home !== '' && away !== '' && Number(home) === Number(away)
  const showPenalty = match.knockout && isDrawPred

  async function save() {
    setStatus('saving')
    try {
      await api.adminEditPrediction(cardId, match.id, {
        home: home === '' ? null : Number(home),
        away: away === '' ? null : Number(away),
        penaltyWinner: showPenalty ? pen : '',
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
      {showPenalty && (
        <select className="admin-pen" value={pen} onChange={(e) => setPen(e.target.value)}>
          <option value="">Penales: —</option>
          <option value="HOME">Gana {match.homeTeamName}</option>
          <option value="AWAY">Gana {match.awayTeamName}</option>
        </select>
      )}
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
