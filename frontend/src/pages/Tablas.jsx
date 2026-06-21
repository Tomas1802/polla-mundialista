import { useEffect, useState } from 'react'
import { api } from '../api.js'

export default function Tablas({ cardId }) {
  const [groups, setGroups] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    setGroups(null)
    setError('')
    api.tables(cardId)
      .then((d) => setGroups(d.groups || []))
      .catch((e) => setError(e.message))
  }, [cardId])

  if (error) return <p className="error page-error">{error}</p>
  if (!groups) return <div className="spinner" />
  if (groups.length === 0) return <p className="empty">Las tablas aparecerán cuando empiece la fase de grupos.</p>

  return (
    <div className="tablas">
      <p className="tablas-intro">
        Tu tabla (según tus marcadores) comparada con la tabla real. Los puntos de cada grupo se
        cuentan cuando el grupo termina.
      </p>
      <p className="tablas-legend">
        <span className="legend-item"><span className="legend-swatch legend-qual" /> clasifican</span>
        <span className="legend-item"><span className="pos-ok">✓</span> posición acertada</span>
      </p>
      {groups.map((g) => {
        const hasReal = (g.real || []).some((r) => r.played > 0)
        const correct = hasReal
          ? g.predicted.map((r, i) => g.real[i] && g.real[i].teamId === r.teamId)
          : null
        return (
          <section key={g.group} className="group-block">
            <header className="group-head">
              <h3>Grupo {g.group}</h3>
              {g.finished
                ? <span className="points-badge">+{g.userPoints} pts</span>
                : <span className="group-state">en curso</span>}
            </header>
            <div className="group-tables">
              <MiniTable title="Tu tabla" rows={g.predicted} correct={correct} />
              <MiniTable title="Tabla real" rows={g.real} />
            </div>
          </section>
        )
      })}
    </div>
  )
}

function MiniTable({ title, rows, correct }) {
  return (
    <div className="mini-table">
      <div className="mini-title">{title}</div>
      <table>
        <thead>
          <tr>
            <th className="pos">#</th>
            <th className="team">Equipo</th>
            <th>PJ</th>
            <th>DG</th>
            <th>Pts</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r, i) => (
            <tr key={r.teamId} className={(i < 2 ? 'qualifies' : '') + (i === 1 ? ' qual-line' : '')}>
              <td className="pos"><span className="pos-badge">{i + 1}</span></td>
              <td className="team">
                <span className="team-label">{r.teamName}</span>
                {correct && correct[i] && <span className="pos-ok" title="Posición acertada">✓</span>}
              </td>
              <td>{r.played}</td>
              <td>{r.goalDiff > 0 ? `+${r.goalDiff}` : r.goalDiff}</td>
              <td><strong>{r.points}</strong></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
