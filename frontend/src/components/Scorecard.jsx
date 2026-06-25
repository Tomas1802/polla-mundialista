// Compact, transparent history of every point a card has earned: one tight row
// per match (with date, result, the player's marcador and the points/rule it
// gave) interleaved with the group-position bonuses awarded when each group
// closed, ending with the grand total — so anyone can audit a player's score.
import { formatShortDate } from '../util/datetime.js'

// After ~3h past kickoff an unfinished match is treated as "result pending"
// rather than live (its final just hasn't synced yet).
const STALE_LIVE_MS = 3.5 * 60 * 60 * 1000

function fmt(n) {
  return n == null ? '–' : n
}

export default function Scorecard({ card }) {
  const matches = card?.matches || []
  const groups = card?.groups || []

  if (matches.length === 0 && groups.length === 0) {
    return <p className="sc-empty">Aún no hay puntos registrados para este cartón.</p>
  }

  // Merge matches and group bonuses into a single timeline. On ties (a group's
  // last match shares its closing date) the match comes first, then its bonus.
  const items = [
    ...matches.map((m) => ({ kind: 'match', date: m.utcDate, data: m })),
    ...groups.map((g) => ({ kind: 'group', date: g.closedAt, data: g })),
  ].sort((a, b) => {
    const d = new Date(a.date) - new Date(b.date)
    if (d !== 0) return d
    return a.kind === b.kind ? 0 : a.kind === 'match' ? -1 : 1
  })

  return (
    <div className="sc">
      <ul className="sc-list">
        {items.map((it) =>
          it.kind === 'match'
            ? <MatchRow key={`m${it.data.id}`} m={it.data} />
            : <GroupRow key={`g${it.data.group}`} g={it.data} />,
        )}
      </ul>
      <div className="sc-total">
        <span>Total</span>
        <strong>{card?.totalPoints ?? 0} pts</strong>
      </div>
    </div>
  )
}

function MatchRow({ m }) {
  const stale = !m.finished && Date.now() - new Date(m.utcDate).getTime() > STALE_LIVE_MS
  const hasPred = m.predHome != null && m.predAway != null
  return (
    <li className="sc-row">
      <span className="sc-day">{formatShortDate(m.utcDate)}</span>
      <div className="sc-body">
        <div className="sc-line1">
          <span className="sc-mt">
            {m.homeTeamName} <b>{fmt(m.scoreHome)}<span className="sc-dash">–</span>{fmt(m.scoreAway)}</b> {m.awayTeamName}
          </span>
          {m.finished ? (
            <span className={'sc-pts' + (m.points > 0 ? '' : ' sc-pts-zero')}>+{m.points}</span>
          ) : stale ? (
            <span className="sc-pending">Actualizando…</span>
          ) : (
            <span className="sc-live">En juego</span>
          )}
        </div>
        <div className="sc-line2">
          <span>{hasPred
            ? <>Pronóstico <b>{m.predHome}<span className="sc-dash">–</span>{m.predAway}</b></>
            : 'Sin pronóstico'}</span>
          {m.finished && m.rule && <span className="sc-rule">{m.rule}</span>}
        </div>
      </div>
    </li>
  )
}

function GroupRow({ g }) {
  return (
    <li className="sc-row sc-row-group">
      <span className="sc-day">{g.closedAt ? formatShortDate(g.closedAt) : ''}</span>
      <div className="sc-body">
        <div className="sc-line1">
          <span className="sc-mt sc-group-title">🏁 Grupo {g.group} cerrado</span>
          <span className={'sc-pts' + (g.points > 0 ? '' : ' sc-pts-zero')}>+{g.points}</span>
        </div>
        <div className="sc-line2">
          <span className="sc-rule sc-rule-group">Posiciones de la tabla</span>
        </div>
      </div>
    </li>
  )
}
