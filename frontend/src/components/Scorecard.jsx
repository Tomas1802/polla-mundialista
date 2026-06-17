// Presentational list of a card's marcadores with the points and rule each one
// earned. Receives an already-loaded scorecard payload from the API.

// After ~3h past kickoff an unfinished match is treated as "result pending"
// rather than live (its final just hasn't synced yet).
const STALE_LIVE_MS = 3.5 * 60 * 60 * 1000

function fmt(n) {
  return n == null ? '–' : n
}

export default function Scorecard({ card }) {
  const matches = card?.matches || []
  if (matches.length === 0) {
    return <p className="sc-empty">Aún no hay partidos jugados para este cartón.</p>
  }
  return (
    <ul className="sc-list">
      {matches.map((m) => (
        <li key={m.id} className="sc-match">
          <div className="sc-teams">
            <span className="sc-team">{m.homeTeamName}</span>
            <span className="sc-real">{fmt(m.scoreHome)}<span className="sc-dash">–</span>{fmt(m.scoreAway)}</span>
            <span className="sc-team sc-team-away">{m.awayTeamName}</span>
          </div>
          <div className="sc-detail">
            <span className="sc-pred">
              Pronóstico <strong>{fmt(m.predHome)}<span className="sc-dash">–</span>{fmt(m.predAway)}</strong>
            </span>
            {m.finished ? (
              <span className="sc-earned">
                <span className={'sc-pts' + (m.points > 0 ? '' : ' sc-pts-zero')}>+{m.points}</span>
                <span className="sc-rule">{m.rule}</span>
              </span>
            ) : Date.now() - new Date(m.utcDate).getTime() > STALE_LIVE_MS ? (
              <span className="sc-pending">Actualizando…</span>
            ) : (
              <span className="sc-live">En juego</span>
            )}
          </div>
        </li>
      ))}
    </ul>
  )
}
