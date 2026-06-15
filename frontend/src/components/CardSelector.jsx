// Tabs to switch between a player's cards (cartones) on the Marcadores and
// Tablas tabs. Rendered only when the player has more than one card.
export default function CardSelector({ cards, cardId, onChange }) {
  return (
    <div className="card-tabs" role="tablist" aria-label="Tus cartones">
      {cards.map((c) => (
        <button
          key={c.id}
          role="tab"
          aria-selected={c.id === cardId}
          className={'card-tab' + (c.id === cardId ? ' active' : '')}
          onClick={() => onChange(c.id)}
        >
          <span className="card-tab-label">{c.label}</span>
          {c.rank ? <span className="card-tab-sub">#{c.rank} · {c.points} pts</span> : null}
        </button>
      ))}
    </div>
  )
}
