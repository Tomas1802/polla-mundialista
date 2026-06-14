import { useState } from 'react'
import { api } from '../api.js'
import { useAuth } from '../auth.jsx'

// Forced on first login: the player replaces their assigned PIN with a new one.
export default function ChangePin() {
  const { refresh, logout, user } = useAuth()
  const [pin, setPin] = useState('')
  const [pin2, setPin2] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function submit(e) {
    e.preventDefault()
    setError('')
    if (!/^\d{4}$/.test(pin)) {
      setError('El PIN debe tener exactamente 4 dígitos.')
      return
    }
    if (pin !== pin2) {
      setError('Los PIN no coinciden.')
      return
    }
    setBusy(true)
    try {
      await api.changePin(pin)
      await refresh()
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="login">
      <div className="login-card">
        <div className="login-logo" aria-hidden="true">🔒</div>
        <h1>Crea tu PIN</h1>
        <form onSubmit={submit}>
          <p className="login-help">
            Hola {user?.playerName}. Por seguridad, cambia el PIN que te dieron por uno nuevo que
            solo tú conozcas.
          </p>
          <label className="field">
            <span>Nuevo PIN (4 dígitos)</span>
            <input type="password" inputMode="numeric" value={pin} maxLength={4}
              onChange={(e) => setPin(e.target.value)} placeholder="••••" required autoFocus />
          </label>
          <label className="field">
            <span>Repite el PIN</span>
            <input type="password" inputMode="numeric" value={pin2} maxLength={4}
              onChange={(e) => setPin2(e.target.value)} placeholder="••••" required />
          </label>
          {error && <p className="error">{error}</p>}
          <button className="primary-btn" disabled={busy}>
            {busy ? 'Guardando…' : 'Guardar PIN'}
          </button>
          <button type="button" className="link-btn" onClick={logout}>Salir</button>
        </form>
      </div>
    </div>
  )
}
