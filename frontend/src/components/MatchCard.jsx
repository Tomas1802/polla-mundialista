import { useState } from 'react'
import { api } from '../api.js'
import { formatDate, formatTime } from '../util/datetime.js'

function clamp(n) {
  return Math.max(0, Math.min(99, n))
}

// After this long past kickoff, an "IN_PLAY/PAUSED" status is treated as a
// not-yet-synced result rather than a genuinely live match.
const STALE_LIVE_MS = 3.5 * 60 * 60 * 1000

const STAGE_LABELS = {
  GROUP_STAGE: 'Fase de grupos',
  LAST_32: 'Dieciseisavos',
  LAST_16: 'Octavos',
  QUARTER_FINALS: 'Cuartos de final',
  SEMI_FINALS: 'Semifinal',
  THIRD_PLACE: 'Tercer puesto',
  FINAL: 'Final',
}

function stageLabel(match) {
  if (match.stage === 'GROUP_STAGE' && match.group) return `Grupo ${match.group}`
  return STAGE_LABELS[match.stage] || match.stage
}

export default function MatchCard({ match, cardId }) {
  const editable = match.editable
  const initial = match.prediction || {}
  const [home, setHome] = useState(initial.home ?? null)
  const [away, setAway] = useState(initial.away ?? null)
  const [pen, setPen] = useState(initial.penaltyWinner ?? '')
  const [status, setStatus] = useState('')

  const finished = match.status === 'FINISHED'
  const rawLive = match.status === 'IN_PLAY' || match.status === 'PAUSED'
  // A match lasts ~3h at most (extra time + penalties). If it still reads "live"
  // long after kickoff, the result just hasn't synced yet — don't call it live.
  const stale = rawLive && Date.now() - new Date(match.utcDate).getTime() > STALE_LIVE_MS
  const live = rawLive && !stale
  const hasReal = match.scoreHome != null && match.scoreAway != null
  const realLabel = finished
    ? 'Resultado real'
    : stale
      ? 'Actualizando resultado…'
      : live
        ? 'En vivo'
        : 'Resultado'
  const isDrawPred = home != null && away != null && home === away
  const showPenalty = editable && match.knockout && isDrawPred

  async function save(nextHome, nextAway, nextPen) {
    setStatus('saving')
    try {
      await api.putPrediction(cardId, match.id, { home: nextHome, away: nextAway, penaltyWinner: nextPen })
      setStatus('saved')
      setTimeout(() => setStatus((s) => (s === 'saved' ? '' : s)), 1600)
    } catch {
      setStatus('error')
    }
  }

  function step(side, delta) {
    if (!editable) return
    if (side === 'home') {
      const v = clamp((home ?? 0) + delta)
      setHome(v)
      save(v, away ?? 0, pen)
    } else {
      const v = clamp((away ?? 0) + delta)
      setAway(v)
      save(home ?? 0, v, pen)
    }
  }

  function choosePenalty(side) {
    const next = pen === side ? '' : side
    setPen(next)
    save(home, away, next)
  }

  return (
    <article className={'match-card' + (match.active ? ' match-active' : '')}>
      <div className="match-meta">
        <span className="match-stage">{stageLabel(match)}</span>
        <span className="match-date">{formatDate(match.utcDate)} · {formatTime(match.utcDate)}</span>
      </div>

      <div className="match-badge-row">
        {match.active && <span className="badge badge-active">Partido actual</span>}
        {!finished && match.tempLocked && <span className="badge badge-locked">Temporalmente no editable</span>}
        {!finished && !match.tempLocked && (editable
          ? <span className="badge badge-open">Puedes editar</span>
          : <span className="badge badge-closed">Cerrado</span>)}
        {finished && <span className="badge badge-final">Finalizado</span>}
      </div>

      <TeamRow
        name={match.homeTeamName}
        crest={match.homeCrest}
        value={home}
        editable={editable}
        onStep={(d) => step('home', d)}
      />
      <TeamRow
        name={match.awayTeamName}
        crest={match.awayCrest}
        value={away}
        editable={editable}
        onStep={(d) => step('away', d)}
      />

      {showPenalty && (
        <div className="penalty">
          <span>¿Quién gana en penales?</span>
          <div className="penalty-options">
            <button
              className={'penalty-btn' + (pen === 'HOME' ? ' selected' : '')}
              onClick={() => choosePenalty('HOME')}
            >
              {match.homeTeamName}
            </button>
            <button
              className={'penalty-btn' + (pen === 'AWAY' ? ' selected' : '')}
              onClick={() => choosePenalty('AWAY')}
            >
              {match.awayTeamName}
            </button>
          </div>
        </div>
      )}

      {editable && status && (
        <p className={'save-status save-' + status}>
          {status === 'saving' && 'Guardando…'}
          {status === 'saved' && 'Guardado ✓'}
          {status === 'error' && 'No se pudo guardar'}
        </p>
      )}

      {hasReal && (
        <div className="match-result">
          <div className="result-compare">
            <span className="result-line">
              <span className="result-tag">Tu pronóstico</span>
              <strong>{(home ?? '–')}–{(away ?? '–')}</strong>
            </span>
            <span className={'result-line' + (live ? ' result-live' : stale ? ' result-pending' : '')}>
              <span className="result-tag">{realLabel}</span>
              <strong>{match.scoreHome}–{match.scoreAway}</strong>
            </span>
          </div>
          {match.points != null && <span className="points-badge">+{match.points} pts</span>}
        </div>
      )}
    </article>
  )
}

function TeamRow({ name, crest, value, editable, onStep }) {
  return (
    <div className="team-row">
      <div className="team-id">
        {crest ? <img className="crest" src={crest} alt="" /> : <span className="crest crest-empty" aria-hidden="true" />}
        <span className="team-name">{name}</span>
      </div>
      {editable ? (
        <div className="stepper">
          <button className="step-btn" onClick={() => onStep(-1)} aria-label={`Quitar gol a ${name}`}>−</button>
          <span className="step-value">{value ?? '–'}</span>
          <button className="step-btn" onClick={() => onStep(1)} aria-label={`Agregar gol a ${name}`}>+</button>
        </div>
      ) : (
        <span className="score-readonly">{value ?? '–'}</span>
      )}
    </div>
  )
}
