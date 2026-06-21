import { useEffect, useState, useCallback } from 'react'
import { api } from '../api.js'

// Read-state is tracked client-side in localStorage, keyed by the day returned
// by the backend — so it resets automatically each day and never touches the DB.
function readKey(date) {
  return `polla_notif_read_${date}`
}
function getReadIds(date) {
  try {
    return new Set(JSON.parse(localStorage.getItem(readKey(date)) || '[]'))
  } catch {
    return new Set()
  }
}
function saveReadIds(date, ids) {
  try {
    localStorage.setItem(readKey(date), JSON.stringify([...ids]))
  } catch {
    // ignore storage errors (private mode, etc.)
  }
}

export default function NotificationsBell() {
  const [data, setData] = useState(null) // { date, notifications }
  const [readIds, setReadIds] = useState(new Set())
  const [open, setOpen] = useState(false)

  const load = useCallback(async () => {
    try {
      const d = await api.notifications()
      setData(d)
      setReadIds(getReadIds(d.date))
    } catch {
      // silent: notifications are non-critical
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const notifs = data?.notifications || []
  const unread = notifs.reduce((n, x) => (readIds.has(x.id) ? n : n + 1), 0)

  function openPanel() {
    setOpen(true)
    if (data) {
      const all = new Set(notifs.map((x) => x.id))
      saveReadIds(data.date, all)
      setReadIds(all)
    }
  }

  return (
    <>
      <button className="bell-btn" onClick={openPanel} aria-label="Notificaciones">
        <span aria-hidden="true">🔔</span>
        {unread > 0 && <span className="bell-badge">{unread > 9 ? '9+' : unread}</span>}
      </button>

      {open && (
        <div className="notif-overlay" onClick={() => setOpen(false)}>
          <div className="notif-panel" onClick={(e) => e.stopPropagation()}>
            <div className="notif-head">
              <h3>Novedades de hoy</h3>
              <button className="notif-close" onClick={() => setOpen(false)} aria-label="Cerrar">✕</button>
            </div>
            {notifs.length === 0 ? (
              <p className="notif-empty">Aún no hay novedades hoy. ¡Que ruede el balón! ⚽</p>
            ) : (
              <ul className="notif-list">
                {notifs.map((n) => (
                  <li key={n.id} className="notif-item">
                    <span className="notif-icon" aria-hidden="true">{n.icon}</span>
                    <div className="notif-text">
                      <div className="notif-title">{n.title}</div>
                      {n.body && <div className="notif-body">{n.body}</div>}
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      )}
    </>
  )
}
