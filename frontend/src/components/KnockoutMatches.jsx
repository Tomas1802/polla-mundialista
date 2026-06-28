import { useEffect, useRef } from 'react'
import MatchCard from './MatchCard.jsx'
import Reglas from './Reglas.jsx'

// KnockoutMatches lists the knockout fixtures (Eliminatoria) with the same
// editable scoreboards as the group phase. MatchCard already shows the "¿Quién
// gana en penales?" winner picker when the predicted score is a draw, which is
// the only case where the advancing team is not implied by the marcador.
export default function KnockoutMatches({ data, cardId }) {
  const activeRef = useRef(null)
  const scrolledRef = useRef(false)

  const matches = (data.matches || [])
    .filter((m) => m.stage !== 'GROUP_STAGE')
    .sort((a, b) => new Date(a.utcDate).getTime() - new Date(b.utcDate).getTime())

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
        <p className="empty">La fase eliminatoria comenzará cuando terminen los grupos.</p>
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
