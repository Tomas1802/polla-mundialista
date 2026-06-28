import { useEffect, useState } from 'react'
import { api } from '../api.js'
import Scorecard from '../components/Scorecard.jsx'

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

  // The podium shows the top 3 distinct people (a player's best card only), so
  // the same person never appears twice — even though the list below keeps every
  // board in its true per-card position.
  const podiumTop = []
  const seenPlayers = new Set()
  for (const e of ranking) {
    if (seenPlayers.has(e.playerId)) continue
    seenPlayers.add(e.playerId)
    podiumTop.push(e)
    if (podiumTop.length === 3) break
  }

  return (
    <div className="ranking">
      <h2 className="ranking-title">Ranking por cartón</h2>
      <Podium top={podiumTop} mePlayerId={mePlayerId} />
      <p className="tablas-intro">Toca un cartón para ver sus marcadores y los puntos de cada partido.</p>
      <ol className="ranking-list">
        {ranking.map((e) => (
          <RankingRow key={e.cardId} entry={e} isMe={e.playerId === mePlayerId} />
        ))}
      </ol>
    </div>
  )
}

const MEDALS = { 1: '🥇', 2: '🥈', 3: '🥉' }

function Podium({ top, mePlayerId }) {
  if (!top || top.length === 0) return null
  const [first, second, third] = top
  // Visual order: 2nd, 1st, 3rd (center stage for the leader).
  return (
    <div className="podium" aria-label="Podio de líderes">
      <PodiumSlot entry={second} place={2} mePlayerId={mePlayerId} />
      <PodiumSlot entry={first} place={1} mePlayerId={mePlayerId} />
      <PodiumSlot entry={third} place={3} mePlayerId={mePlayerId} />
    </div>
  )
}

function PodiumSlot({ entry, place, mePlayerId }) {
  if (!entry) return <div className="podium-slot podium-empty" aria-hidden="true" />
  const isMe = entry.playerId === mePlayerId
  return (
    <div className={`podium-slot podium-${place}` + (isMe ? ' me' : '')}>
      <div className="podium-medal" aria-hidden="true">{MEDALS[place]}</div>
      <div className="podium-name">{entry.playerName}</div>
      <div className="podium-bar">
        <span className="podium-pts">{entry.points}</span>
        <span className="podium-place">{place}°</span>
      </div>
    </div>
  )
}

function RankingRow({ entry, isMe }) {
  const [card, setCard] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function onToggle(ev) {
    if (!ev.target.open || card || loading) return
    setLoading(true)
    setError('')
    try {
      setCard(await api.scorecard(entry.cardId))
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <li>
      <details className={'rank-details' + (isMe ? ' me' : '')} onToggle={onToggle}>
        <summary className="ranking-row">
          <span className="rank-pos">{entry.rank}</span>
          <span className="rank-name">
            {entry.playerName}
            <span className="rank-card">{entry.cardLabel}</span>
            {isMe && <span className="you-tag">Tú</span>}
          </span>
          <span className="rank-pts">{entry.points} pts</span>
          <span className="rank-caret" aria-hidden="true">▾</span>
        </summary>
        <div className="rank-panel">
          {loading && <div className="spinner" />}
          {error && <p className="error">{error}</p>}
          {card && <Scorecard card={card} />}
        </div>
      </details>
    </li>
  )
}
