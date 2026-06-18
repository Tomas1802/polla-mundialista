import { useEffect, useState, useCallback } from 'react'
import { useAuth } from './auth.jsx'
import { api } from './api.js'
import Login from './components/Login.jsx'
import ChangePin from './components/ChangePin.jsx'
import CardSelector from './components/CardSelector.jsx'
import Marcadores from './pages/Marcadores.jsx'
import Tablas from './pages/Tablas.jsx'
import Ranking from './pages/Ranking.jsx'
import Equipos from './pages/Equipos.jsx'
import Admin from './pages/Admin.jsx'

const BASE_TABS = [
  { id: 'marcadores', label: 'Marcadores', icon: '⚽' },
  { id: 'tablas', label: 'Tablas', icon: '📊' },
  { id: 'ranking', label: 'Ranking', icon: '🏆' },
  { id: 'equipos', label: 'Equipos', icon: '👥' },
]
const ADMIN_TAB = { id: 'admin', label: 'Admin', icon: '🔧' }

export default function App() {
  const { user, loading, logout } = useAuth()
  const [tab, setTab] = useState('marcadores')
  const [cards, setCards] = useState(null)
  const [cardId, setCardId] = useState(null)

  const loadCards = useCallback(async () => {
    try {
      const d = await api.cards()
      const list = d.cards || []
      setCards(list)
      // Keep the current selection only if it belongs to this list; otherwise
      // fall back to the first card. Prevents reusing another user's cardId.
      setCardId((prev) =>
        prev && list.some((c) => c.id === prev) ? prev : list.length ? list[0].id : null,
      )
    } catch {
      setCards([])
    }
  }, [])

  // Reset card state whenever the logged-in user changes (e.g. logout + login as
  // someone else without a page reload), then load the new user's cards.
  useEffect(() => {
    setCards(null)
    setCardId(null)
    if (user && user.playerId) loadCards()
  }, [user, loadCards])

  // Admins (no player) land on the Admin tab.
  useEffect(() => {
    if (user && user.isAdmin && !user.playerId) setTab('admin')
  }, [user])

  if (loading) {
    return <div className="center-screen"><div className="spinner" aria-label="Cargando" /></div>
  }
  if (!user) return <Login />
  if (user.playerId && user.mustChangePin) return <ChangePin />

  // The admin account does not play: it only sees Ranking, Equipos and Admin.
  const tabs = user.isAdmin
    ? [...BASE_TABS.filter((t) => t.id === 'ranking' || t.id === 'equipos'), ADMIN_TAB]
    : BASE_TABS
  const onCardTab = tab === 'marcadores' || tab === 'tablas'
  const showCardSelector = onCardTab && cards && cards.length > 1
  const cardsLoading = onCardTab && user.playerId && cards === null
  const noCard = onCardTab && cards !== null && !cardId

  return (
    <div className="app">
      <header className="app-header">
        <span className="app-title">Polla Mundial 2026</span>
        <button className="logout-btn" onClick={logout}>Salir</button>
      </header>

      <main className="app-main">
        {showCardSelector && <CardSelector cards={cards} cardId={cardId} onChange={setCardId} />}

        {cardsLoading && <div className="spinner" aria-label="Cargando cartones" />}

        {noCard && (
          <p className="empty">
            {user.isAdmin
              ? 'Eres administrador. Usa la pestaña Admin para importar y repartir PINs.'
              : 'Aún no tienes cartones cargados. El administrador los cargará pronto.'}
          </p>
        )}

        {tab === 'marcadores' && cardId && <Marcadores cardId={cardId} />}
        {tab === 'tablas' && cardId && <Tablas cardId={cardId} />}
        {tab === 'ranking' && <Ranking />}
        {tab === 'equipos' && <Equipos />}
        {tab === 'admin' && user.isAdmin && <Admin />}
      </main>

      <nav className="tabbar" aria-label="Secciones">
        {tabs.map((t) => (
          <button
            key={t.id}
            className={'tab' + (tab === t.id ? ' tab-active' : '')}
            onClick={() => setTab(t.id)}
            aria-current={tab === t.id ? 'page' : undefined}
          >
            <span className="tab-icon" aria-hidden="true">{t.icon}</span>
            <span className="tab-label">{t.label}</span>
          </button>
        ))}
      </nav>
    </div>
  )
}
