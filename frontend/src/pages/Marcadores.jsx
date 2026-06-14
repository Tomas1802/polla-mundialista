import { useEffect, useRef, useState, useCallback } from 'react'
import { api } from '../api.js'
import MatchCard from '../components/MatchCard.jsx'
import Reglas from '../components/Reglas.jsx'

export default function Marcadores({ cardId }) {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const activeRef = useRef(null)
  const scrolledRef = useRef(false)

  const load = useCallback(async () => {
    try {
      setData(await api.matches(cardId))
    } catch (e) {
      setError(e.message)
    }
  }, [cardId])

  useEffect(() => {
    setData(null)
    setError('')
    scrolledRef.current = false
    load()
  }, [load])

  // Auto-scroll to the active match once per card load.
  useEffect(() => {
    if (data && activeRef.current && !scrolledRef.current) {
      scrolledRef.current = true
      activeRef.current.scrollIntoView({ block: 'center' })
    }
  }, [data])

  if (error) return <p className="error page-error">{error}</p>
  if (!data) return <div className="spinner" />

  return (
    <div className="marcadores">
      <div className="rank-chip">
        <span>Tu puesto</span>
        <strong>#{data.myRank || '—'}</strong>
        <span>de {data.totalCards}</span>
        <span className="rank-points">{data.myPoints} pts</span>
      </div>

      <Reglas />

      {data.matches.length === 0 && (
        <p className="empty">Aún no hay partidos cargados. Vuelve pronto.</p>
      )}

      <div className="match-list">
        {data.matches.map((m) => (
          <div key={m.id} ref={m.id === data.activeMatchId ? activeRef : null}>
            <MatchCard match={m} cardId={cardId} />
          </div>
        ))}
      </div>
    </div>
  )
}
