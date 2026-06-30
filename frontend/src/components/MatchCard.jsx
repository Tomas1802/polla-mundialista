import { useState, useRef, useEffect } from 'react'
import { api } from '../api.js'
import { getPending, setPending, clearPending } from '../util/predictions.js'
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

const eq = (a, b) =>
  a.home === b.home && a.away === b.away && (a.pen || '') === (b.pen || '')

export default function MatchCard({ match, cardId }) {
  const editable = match.editable
  const server = match.prediction || {}
  // A leftover localStorage draft (edited but never synced) wins over the
  // server value so the player never loses unsynced work across reloads.
  const pending = getPending(cardId, match.id)
  const initial = pending || server
  const [home, setHome] = useState(initial.home ?? null)
  const [away, setAway] = useState(initial.away ?? null)
  const [pen, setPen] = useState(initial.penaltyWinner ?? '')
  const [status, setStatus] = useState('')
  const [dirty, setDirty] = useState(!!pending)

  // syncedRef = last value the server confirmed; draftRef = latest local edit.
  // Comparing them tells us whether there's anything to push, and guards the
  // success path against an out-of-order response clearing a newer edit.
  const syncedRef = useRef({ home: server.home ?? null, away: server.away ?? null, pen: server.penaltyWinner ?? '' })
  const draftRef = useRef({ home: home, away: away, pen: pen })
  const timerRef = useRef(null)

  async function sync() {
    clearTimeout(timerRef.current)
    const snap = draftRef.current
    if (eq(snap, syncedRef.current)) {
      setDirty(false)
      return
    }
    setStatus('saving')
    try {
      await api.putPrediction(cardId, match.id, { home: snap.home, away: snap.away, penaltyWinner: snap.pen })
      syncedRef.current = snap
      // Clear the local backup only if the player hasn't edited again while
      // this request was in flight; otherwise the newer draft stays pending.
      if (eq(draftRef.current, snap)) {
        clearPending(cardId, match.id)
        setDirty(false)
        setStatus('saved')
        setTimeout(() => setStatus((s) => (s === 'saved' ? '' : s)), 1600)
      }
    } catch {
      setStatus('error') // local backup kept; confirm button offers a retry
    }
  }

  // Flush any leftover unsynced draft when the card mounts (e.g. the player
  // reopened the app after a save failed), and cancel pending timers on unmount.
  useEffect(() => {
    if (dirty) sync()
    return () => clearTimeout(timerRef.current)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Apply an edit: update the display, persist locally at once (never lost),
  // then optimistically auto-save after a short debounce that collapses rapid
  // taps into one request and keeps writes in order.
  function commit(nh, na, np) {
    setHome(nh)
    setAway(na)
    setPen(np)
    const draft = { home: nh, away: na, pen: np }
    draftRef.current = draft
    if (eq(draft, syncedRef.current)) {
      clearPending(cardId, match.id)
      setDirty(false)
      setStatus('')
      clearTimeout(timerRef.current)
      return
    }
    setPending(cardId, match.id, { home: nh, away: na, penaltyWinner: np })
    setDirty(true)
    clearTimeout(timerRef.current)
    timerRef.current = setTimeout(sync, 800)
  }

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

  function step(side, delta) {
    if (!editable) return
    if (side === 'home') commit(clamp((home ?? 0) + delta), away ?? 0, pen)
    else commit(home ?? 0, clamp((away ?? 0) + delta), pen)
  }

  function choosePenalty(side) {
    commit(home, away, pen === side ? '' : side)
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

      {editable && (dirty || status === 'saving' || status === 'error') && (
        <button
          className={'confirm-btn' + (status === 'error' ? ' confirm-error' : '')}
          onClick={sync}
          disabled={status === 'saving'}
        >
          {status === 'saving'
            ? 'Guardando…'
            : status === 'error'
              ? 'No se guardó — reintentar'
              : 'Confirmar marcador'}
        </button>
      )}
      {editable && status === 'saved' && <p className="save-status save-saved">Guardado ✓</p>}

      {hasReal && (
        <div className="match-result">
          <div className="result-compare">
            <span className={'result-line result-real' + (live ? ' result-live' : stale ? ' result-pending' : '')}>
              <span className="result-tag">{realLabel}</span>
              <strong>{match.scoreHome}–{match.scoreAway}</strong>
            </span>
            <span className="result-line result-pred">
              <span className="result-tag">Tu pronóstico</span>
              <strong>{(home ?? '–')}–{(away ?? '–')}</strong>
            </span>
          </div>
          {match.points != null && (
            match.provisional ? (
              <span className="points-live" title="Puntos si el marcador se mantiene así">
                <span className="points-live-num">+{match.points}</span>
                <span className="points-live-tag">provisional</span>
              </span>
            ) : (
              <span className="points-badge">+{match.points} pts</span>
            )
          )}
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
