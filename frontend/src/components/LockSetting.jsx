import { useEffect, useState } from 'react'
import { api } from '../api.js'

const dateFmt = new Intl.DateTimeFormat('es', { day: 'numeric', month: 'short' })

// Admin control: choose the cutoff match. Matches up to and including it are
// "Temporalmente no editable" for players.
export default function LockSetting() {
  const [data, setData] = useState(null)
  const [sel, setSel] = useState('')
  const [status, setStatus] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    api.adminSettings()
      .then((d) => {
        setData(d)
        setSel(d.editLockUntilMatchId ? String(d.editLockUntilMatchId) : '')
      })
      .catch((e) => setError(e.message))
  }, [])

  async function save() {
    setStatus('saving')
    setError('')
    try {
      await api.adminSetSettings(sel === '' ? null : Number(sel))
      setStatus('saved')
      setTimeout(() => setStatus((s) => (s === 'saved' ? '' : s)), 1800)
    } catch (e) {
      setStatus('')
      setError(e.message)
    }
  }

  if (error && !data) return <p className="error">{error}</p>
  if (!data) return <div className="spinner" />

  return (
    <section className="admin-section">
      <h2>Bloqueo de edición</h2>
      <p className="tablas-intro">
        Los partidos hasta (incluido) el seleccionado quedan <strong>“Temporalmente no editable”</strong>
        {' '}para los jugadores. Cámbialo cuando habilites una nueva fase, o elige “Sin bloqueo”.
      </p>
      <label className="field">
        <span>Bloquear hasta (incluido)</span>
        <select value={sel} onChange={(e) => setSel(e.target.value)}>
          <option value="">— Sin bloqueo —</option>
          {data.matches.map((m) => (
            <option key={m.id} value={m.id}>
              {m.home} vs {m.away} — {dateFmt.format(new Date(m.utcDate))}
            </option>
          ))}
        </select>
      </label>
      <div className="admin-save-row">
        <button className="primary-btn admin-save-btn" onClick={save} disabled={status === 'saving'}>
          {status === 'saving' ? 'Guardando…' : 'Guardar'}
        </button>
        {status === 'saved' && <span className="save-status save-saved">Guardado ✓</span>}
      </div>
      {error && <p className="error">{error}</p>}
    </section>
  )
}
