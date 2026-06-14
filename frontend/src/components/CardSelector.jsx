// Lets a player switch between their cards (cartones) on the Marcadores and
// Tablas tabs. Hidden when the player has a single card.
export default function CardSelector({ cards, cardId, onChange }) {
  return (
    <div className="card-selector">
      <span className="card-selector-label">Tu cartón:</span>
      <div className="card-chips">
        {cards.map((c) => (
          <button
            key={c.id}
            className={'card-chip' + (c.id === cardId ? ' active' : '')}
            onClick={() => onChange(c.id)}
          >
            {c.label}{c.rank ? ` · #${c.rank}` : ''}
          </button>
        ))}
      </div>
    </div>
  )
}
