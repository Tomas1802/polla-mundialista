import { useEffect, useState, useCallback } from 'react'
import { useAuth } from './auth.jsx'
import { api } from './api.js'
import Login from './components/Login.jsx'
import ChangePin from './components/ChangePin.jsx'
import CardSelector from './components/CardSelector.jsx'
import Marcadores from './pages/Marcadores.jsx'
import Tablas from './pages/Tablas.jsx'
import Ranking from './pages/Ranking.jsx'
import Admin from './pages/Admin.jsx'

const BASE_TABS = [
  { id: 'marcadores', label: 'Marcadores', icon: '⚽' },
  { id: 'tablas', label: 'Tablas', icon: '📊' },
  { id: 'ranking', label: 'Ranking', icon: '🏆' },
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
      setCardId((prev) => prev || (list.length ? list[0].id : null))
    } catch {
      setCards([])
    }
  }, [])

  useEffect(() => {
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

  const tabs = user.isAdmin ? [...BASE_TABS, ADMIN_TAB] : BASE_TABS
  const showCardSelector = (tab === 'marcadores' || tab === 'tablas') && cards && cards.length > 1
  const noCard = (tab === 'marcadores' || tab === 'tablas') && !cardId

  return (
    <div className="app">
      <header className="app-header">
        <span className="app-title">Polla Mundial 2026</span>
        <button className="logout-btn" onClick={logout}>Salir</button>
      </header>

      <main className="app-main">
        {showCardSelector && <CardSelector cards={cards} cardId={cardId} onChange={setCardId} />}

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
        {tab === 'admin' && user.isAdmin && <Admin onImported={loadCards} />}
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
