import { useEffect, useRef } from 'react'
import MatchCard from './MatchCard.jsx'
import Reglas from './Reglas.jsx'

// GroupMatches lists the group-stage fixtures (Fase 1 → "Marcadores"). Knockout
// fixtures are intentionally excluded here; they live in the Eliminatoria phase.
export default function GroupMatches({ data, cardId }) {
  const activeRef = useRef(null)
  const scrolledRef = useRef(false)

  const matches = (data.matches || []).filter((m) => m.stage === 'GROUP_STAGE')

  // Auto-scroll to the active match once per mount.
  useEffect(() => {
    if (activeRef.current && !scrolledRef.current) {
      scrolledRef.current = true
      activeRef.current.scrollIntoView({ block: 'center' })
    }
  }, [])

  return (
    <div className="grupos-marcadores">
      <Reglas />

      {matches.length === 0 && (
        <p className="empty">Aún no hay partidos cargados. Vuelve pronto.</p>
      )}

      <div className="match-list">
        {matches.map((m) => (
          <div key={m.id} ref={m.id === data.activeMatchId ? activeRef : null}>
            <MatchCard match={m} cardId={cardId} />
          </div>
        ))}
      </div>
    </div>
  )
}
