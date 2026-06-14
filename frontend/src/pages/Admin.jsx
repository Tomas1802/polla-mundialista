import { useEffect, useState, useCallback } from 'react'
import { api } from '../api.js'

export default function Admin({ onImported }) {
  const [players, setPlayers] = useState(null)
  const [error, setError] = useState('')
  const [importing, setImporting] = useState(false)
  const [result, setResult] = useState(null)

  const loadPlayers = useCallback(async () => {
    try {
      const d = await api.adminPlayers()
      setPlayers(d.players || [])
    } catch (e) {
      setError(e.message)
    }
  }, [])

  useEffect(() => {
    loadPlayers()
  }, [loadPlayers])

  async function doImport() {
    setImporting(true)
    setError('')
    setResult(null)
    try {
      const res = await api.adminImport()
      setResult(res)
      await loadPlayers()
      if (onImported) onImported()
    } catch (e) {
      setError(e.message)
    } finally {
      setImporting(false)
    }
  }

  if (error && !players) return <p className="error page-error">{error}</p>
  if (!players) return <div className="spinner" />

  return (
    <div className="admin">
      <section className="admin-section">
        <h2>Cartones (importar)</h2>
        <p className="tablas-intro">
          Carga los cartones desde la carpeta <code>scores</code>: crea jugadores, cartones,
          marcadores y asigna un PIN a cada jugador nuevo.
        </p>
        <button className="primary-btn admin-save-btn" onClick={doImport} disabled={importing}>
          {importing ? 'Importando…' : 'Importar cartones'}
        </button>
        {result && (
          <p className="save-status save-saved">
            Listo: {result.players} jugadores · {result.cards} cartones · {result.predictions} marcadores.
          </p>
        )}
        {error && <p className="error">{error}</p>}
      </section>

      <section className="admin-section">
        <h2>Jugadores ({players.length})</h2>
        <p className="tablas-intro">
          Los PINs se generan aparte (archivo CSV) y se entregan a cada jugador. Aquí ves el estado:
          “PIN sin asignar”, “PIN entregado” (aún el asignado) o “PIN propio ✓” (ya lo cambió).
        </p>
        <ul className="admin-users">
          {players.map((p) => (
            <li key={p.id}>
              <span className="admin-user-name">
                {p.name}
                {p.cardCount > 1 && <span className="rank-card">{p.cardCount} cartones</span>}
              </span>
              <span className="admin-user-email">
                {!p.hasPin
                  ? <span className="pin-changed">PIN sin asignar</span>
                  : p.mustChangePin
                    ? <span className="pin-changed">PIN entregado</span>
                    : <span className="pin-changed">PIN propio ✓</span>}
              </span>
            </li>
          ))}
        </ul>
      </section>
    </div>
  )
}
