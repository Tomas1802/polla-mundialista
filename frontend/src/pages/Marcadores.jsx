import { useEffect, useState, useCallback } from 'react'
import { api } from '../api.js'
import CardSelector from '../components/CardSelector.jsx'
import GroupMatches from '../components/GroupMatches.jsx'
import GroupTables from '../components/GroupTables.jsx'
import Bracket from '../components/Bracket.jsx'

// Marcadores is the unified play view. A compact sticky header keeps the key
// controls in view even after the list auto-scrolls to the current match:
//   • the player's rank for the selected card
//   • the card switcher (when the player has more than one cartón)
//   • the phase switch (Fase de grupos / Eliminatoria)
//   • the group sub-switch (Marcadores / Tablas)
// Group fixtures and the bracket are both derived from a single /matches load.
export default function Marcadores({ cardId, cards = [], onCardChange }) {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [phase, setPhase] = useState('grupos') // 'grupos' | 'eliminatoria'
  const [groupView, setGroupView] = useState('marcadores') // 'marcadores' | 'tablas'

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
    load()
  }, [load])

  return (
    <div className="marcadores">
      <header className="mk-header">
        <div className="mk-bar">
          <div className="mk-rank">
            <span className="mk-rank-pos">#{data?.myRank || '—'}</span>
            <span className="mk-rank-meta">
              de {data?.totalCards ?? '—'}
              <span className="mk-rank-pts">{data?.myPoints ?? 0} pts</span>
            </span>
          </div>
          {cards.length > 1 && (
            <CardSelector cards={cards} cardId={cardId} onChange={onCardChange} />
          )}
        </div>

        <div className="phase-toggle" role="tablist" aria-label="Fase">
          <button
            role="tab"
            aria-selected={phase === 'grupos'}
            className={'phase-btn' + (phase === 'grupos' ? ' phase-active' : '')}
            onClick={() => setPhase('grupos')}
          >
            Fase de grupos
          </button>
          <button
            role="tab"
            aria-selected={phase === 'eliminatoria'}
            className={'phase-btn' + (phase === 'eliminatoria' ? ' phase-active' : '')}
            onClick={() => setPhase('eliminatoria')}
          >
            Eliminatoria
          </button>
        </div>

        {phase === 'grupos' && (
          <div className="subtoggle" role="tablist" aria-label="Vista de grupos">
            <button
              role="tab"
              aria-selected={groupView === 'marcadores'}
              className={'subtoggle-btn' + (groupView === 'marcadores' ? ' subtoggle-active' : '')}
              onClick={() => setGroupView('marcadores')}
            >
              ⚽ Marcadores
            </button>
            <button
              role="tab"
              aria-selected={groupView === 'tablas'}
              className={'subtoggle-btn' + (groupView === 'tablas' ? ' subtoggle-active' : '')}
              onClick={() => setGroupView('tablas')}
            >
              📊 Tablas
            </button>
          </div>
        )}
      </header>

      {error && <p className="error page-error">{error}</p>}
      {!error && !data && <div className="spinner" />}

      {!error && data && phase === 'grupos' && (
        groupView === 'marcadores'
          ? <GroupMatches data={data} cardId={cardId} />
          : <GroupTables cardId={cardId} />
      )}

      {!error && data && phase === 'eliminatoria' && <Bracket matches={data.matches} />}
    </div>
  )
}
