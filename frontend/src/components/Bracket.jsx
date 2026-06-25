import { formatDate } from '../util/datetime.js'

// Rounds drawn from the outside of the bracket inward toward the final.
const ROUNDS = [
  { stage: 'LAST_32', label: '16avos' },
  { stage: 'LAST_16', label: 'Octavos' },
  { stage: 'QUARTER_FINALS', label: 'Cuartos' },
  { stage: 'SEMI_FINALS', label: 'Semis' },
]

const TBD = 'Por definir'

function byDate(a, b) {
  return new Date(a.utcDate).getTime() - new Date(b.utcDate).getTime()
}

// Bracket renders the knockout phase as a symmetric tree: the two halves fan in
// from the sides toward the final in the center. Teams and scores come straight
// from the synced fixtures, so it updates as teams advance. Horizontally
// scrollable on small screens.
export default function Bracket({ matches }) {
  const knockout = (matches || []).filter((m) => m.stage !== 'GROUP_STAGE')

  if (knockout.length === 0) {
    return (
      <p className="empty">
        La fase eliminatoria comenzará cuando terminen los grupos. Aquí verás el cuadro con los
        equipos que vayan clasificando.
      </p>
    )
  }

  const byStage = {}
  for (const m of knockout) {
    ;(byStage[m.stage] ||= []).push(m)
  }
  for (const k of Object.keys(byStage)) byStage[k].sort(byDate)

  const leftColumns = ROUNDS.map((r) => {
    const arr = byStage[r.stage] || []
    return { ...r, matches: arr.slice(0, Math.ceil(arr.length / 2)) }
  })
  const rightColumns = [...ROUNDS].reverse().map((r) => {
    const arr = byStage[r.stage] || []
    return { ...r, matches: arr.slice(Math.ceil(arr.length / 2)) }
  })
  const final = byStage.FINAL || []
  const third = byStage.THIRD_PLACE || []

  return (
    <div className="bracket-wrap">
      <p className="bracket-hint">Desliza horizontalmente para ver todo el cuadro →</p>
      <div className="bracket">
        {leftColumns.map((col) => (
          <BracketColumn key={`L-${col.stage}`} label={col.label} matches={col.matches} />
        ))}

        <div className="bk-col bk-center">
          <div className="bk-round-label bk-final-label">🏆 Final</div>
          {final.length ? (
            final.map((m) => <BracketMatch key={m.id} m={m} big />)
          ) : (
            <BracketMatch placeholder big />
          )}
          {third.length > 0 && (
            <>
              <div className="bk-round-label bk-third-label">3er puesto</div>
              {third.map((m) => <BracketMatch key={m.id} m={m} />)}
            </>
          )}
        </div>

        {rightColumns.map((col) => (
          <BracketColumn key={`R-${col.stage}`} label={col.label} matches={col.matches} />
        ))}
      </div>
    </div>
  )
}

function BracketColumn({ label, matches }) {
  return (
    <div className="bk-col">
      <div className="bk-round-label">{label}</div>
      {matches.length === 0
        ? <BracketMatch placeholder />
        : matches.map((m) => <BracketMatch key={m.id} m={m} />)}
    </div>
  )
}

function BracketMatch({ m, big, placeholder }) {
  if (placeholder || !m) {
    return (
      <div className={'bk-match bk-match-empty' + (big ? ' bk-final' : '')}>
        <BkTeam name={TBD} tbd />
        <BkTeam name={TBD} tbd />
      </div>
    )
  }

  const hasScore = m.scoreHome != null && m.scoreAway != null
  const homeWin = m.winner === 'HOME_WIN' || (hasScore && m.scoreHome > m.scoreAway)
  const awayWin = m.winner === 'AWAY_WIN' || (hasScore && m.scoreAway > m.scoreHome)
  const pens = m.duration === 'PENALTY_SHOOTOUT'
  const pred = m.prediction
  const hasPred = pred && pred.home != null && pred.away != null

  return (
    <div className={'bk-match' + (big ? ' bk-final' : '')}>
      <BkTeam
        name={m.homeTeamName}
        crest={m.homeCrest}
        score={hasScore ? m.scoreHome : null}
        win={homeWin}
        pen={pens && homeWin}
        tbd={m.homeTeamName === TBD}
      />
      <BkTeam
        name={m.awayTeamName}
        crest={m.awayCrest}
        score={hasScore ? m.scoreAway : null}
        win={awayWin}
        pen={pens && awayWin}
        tbd={m.awayTeamName === TBD}
      />
      {!hasScore && m.homeTeamName !== TBD && (
        <div className="bk-date">{formatDate(m.utcDate)}</div>
      )}
      {hasPred && (
        <div className="bk-pred">
          <span>Tú: {pred.home}–{pred.away}</span>
          {m.points != null && (
            <span className={m.provisional ? 'bk-pts bk-pts-live' : 'bk-pts'}>+{m.points}</span>
          )}
        </div>
      )}
    </div>
  )
}

function BkTeam({ name, crest, score, win, pen, tbd }) {
  return (
    <div className={'bk-team' + (win ? ' bk-team-win' : '') + (tbd ? ' bk-team-tbd' : '')}>
      {crest
        ? <img className="bk-crest" src={crest} alt="" />
        : <span className="bk-crest bk-crest-empty" aria-hidden="true" />}
      <span className="bk-name">{name}</span>
      {pen && <span className="bk-pen" title="Ganó en penales">pen</span>}
      <span className="bk-score">{score != null ? score : ''}</span>
    </div>
  )
}
