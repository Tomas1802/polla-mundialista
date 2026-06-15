import { useEffect, useState } from 'react'
import { api } from '../api.js'

export default function Equipos() {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api.teams()
      .then(setData)
      .catch((e) => setError(e.message))
  }, [])

  if (error) return <p className="error page-error">{error}</p>
  if (!data) return <div className="spinner" />

  const { teams, sinEquipo, mePlayerId } = data
  if (!teams || teams.length === 0) {
    return <p className="empty">Los equipos aparecerán cuando el administrador los cargue.</p>
  }

  const mine = (members) => members.some((m) => m.playerId === mePlayerId)

  return (
    <div className="equipos">
      <h2 className="ranking-title">Resultado por equipos</h2>
      <p className="tablas-intro">
        El puntaje del equipo es la suma del <strong>mejor cartón</strong> de cada integrante.
      </p>

      {teams.map((t) => (
        <section key={t.name} className={'team-block' + (mine(t.members) ? ' me' : '')}>
          <header className="team-head">
            <span className="team-rank">{t.rank}</span>
            <span className="team-name">{t.name}</span>
            <span className="team-points">{t.points} pts</span>
          </header>
          <ul className="team-members">
            {t.members.map((m) => (
              <li key={m.playerId}>
                <span>{m.name}{m.playerId === mePlayerId && <span className="you-tag">Tú</span>}</span>
                <span className="member-points">{m.points}</span>
              </li>
            ))}
          </ul>
        </section>
      ))}

      {sinEquipo && sinEquipo.length > 0 && (
        <section className="team-block">
          <header className="team-head">
            <span className="team-name">Sin equipo</span>
          </header>
          <ul className="team-members">
            {sinEquipo.map((m) => (
              <li key={m.playerId}>
                <span>{m.name}{m.playerId === mePlayerId && <span className="you-tag">Tú</span>}</span>
                <span className="member-points">{m.points}</span>
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  )
}
