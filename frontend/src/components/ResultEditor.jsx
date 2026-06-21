import { useEffect, useState } from 'react'
import { api } from '../api.js'
import { formatDate, formatTime } from '../util/datetime.js'

// Admin tool: confirm or correct the official result of a played match. The page
// shows the result from the service; once the admin confirms (or corrects) it,
// the result is frozen and the sync no longer consults/overwrites it. Points
// recompute automatically from the shown result.
export default function ResultEditor() {
  const [matches, setMatches] = useState(null)
  const [error, setError] = useState('')

  async function load() {
    setError('')
    try {
      const d = await api.adminMatches()
      setMatches(d.matches || [])
    } catch (e) {
      setError(e.message)
      setMatches([])
    }
  }

  useEffect(() => {
    load()
  }, [])

  return (
    <section className="admin-section">
      <h2>Confirmar resultados</h2>
      <p className="tablas-intro">
        Cada partido muestra el resultado del servicio. Al <strong>confirmar</strong>, el marcador
        queda fijo y deja de consultarse. Si quedó mal, corrígelo y guárdalo (también queda confirmado).
      </p>
      {error && <p className="error">{error}</p>}
      {matches === null && <div className="spinner" />}
      {matches && matches.length === 0 && <p className="empty">Aún no hay partidos jugados.</p>}
      {matches && matches.map((m) => <ResultRow key={m.id} match={m} onChanged={load} />)}
    </section>
  )
}

function ResultRow({ match, onChanged }) {
  const [home, setHome] = useState(match.scoreHome ?? '')
  const [away, setAway] = useState(match.scoreAway ?? '')
  const [status, setStatus] = useState('')

  async function run(fn) {
    setStatus('saving')
    try {
      await fn()
      setStatus('saved')
      onChanged && onChanged()
      setTimeout(() => setStatus((s) => (s === 'saved' ? '' : s)), 1500)
    } catch (e) {
      setStatus(e.message)
    }
  }

  const save = () =>
    run(() =>
      api.adminSetResult(match.id, {
        home: home === '' ? null : Number(home),
        away: away === '' ? null : Number(away),
      }),
    )
  const confirm = () => run(() => api.adminConfirmResult(match.id))
  const revert = () => run(() => api.adminClearResult(match.id))

  return (
    <div className="result-row">
      <div className="result-row-head">
        <span className="result-row-teams">
          {match.home} <span className="vs">vs</span> {match.away}
        </span>
        {match.resultManual
          ? <span className="badge badge-confirmed">Confirmado</span>
          : <span className="badge badge-auto">Automático</span>}
      </div>
      <div className="result-row-sub">
        <span className="result-row-date">{formatDate(match.utcDate)} · {formatTime(match.utcDate)}</span>
      </div>
      <div className="result-row-edit">
        <input className="admin-score" type="number" min="0" max="99" value={home}
          onChange={(e) => setHome(e.target.value)} />
        <span className="admin-dash">–</span>
        <input className="admin-score" type="number" min="0" max="99" value={away}
          onChange={(e) => setAway(e.target.value)} />
        <button className="fix-save" onClick={save} disabled={status === 'saving'}>Guardar</button>
        {match.resultManual ? (
          <button className="link-btn result-revert" onClick={revert} disabled={status === 'saving'}>
            volver a automático
          </button>
        ) : (
          <button className="confirm-btn" onClick={confirm} disabled={status === 'saving'}>Confirmar</button>
        )}
        {status === 'saved' && <span className="save-status save-saved">✓</span>}
        {status && status !== 'saving' && status !== 'saved' && (
          <span className="save-status save-error">{status}</span>
        )}
      </div>
    </div>
  )
}
