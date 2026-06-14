import { useEffect, useState } from 'react'
import { api } from '../api.js'

export default function Ranking() {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api.ranking()
      .then(setData)
      .catch((e) => setError(e.message))
  }, [])

  if (error) return <p className="error page-error">{error}</p>
  if (!data) return <div className="spinner" />

  const { ranking, mePlayerId } = data
  if (!ranking || ranking.length === 0) {
    return <p className="empty">El ranking aparecerá cuando se carguen los cartones y se jueguen partidos.</p>
  }

  return (
    <div className="ranking">
      <h2 className="ranking-title">Ranking por cartón</h2>
      <ol className="ranking-list">
        {ranking.map((e) => (
          <li
            key={e.cardId}
            className={'ranking-row' + (e.playerId === mePlayerId ? ' me' : '')}
          >
            <span className="rank-pos">{e.rank}</span>
            <span className="rank-name">
              {e.playerName}
              <span className="rank-card">{e.cardLabel}</span>
              {e.playerId === mePlayerId && <span className="you-tag">Tú</span>}
            </span>
            <span className="rank-pts">{e.points} pts</span>
          </li>
        ))}
      </ol>
    </div>
  )
}
