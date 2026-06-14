import { useEffect, useState } from 'react'
import { api } from '../api.js'
import { useAuth } from '../auth.jsx'

// Login by player name + 4-digit PIN, with a separate master-admin entry.
export default function Login() {
  const { setUser } = useAuth()
  const [mode, setMode] = useState('player') // 'player' | 'admin'
  const [players, setPlayers] = useState(null)
  const [playerId, setPlayerId] = useState('')
  const [pin, setPin] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    api.players()
      .then((d) => setPlayers(d.players || []))
      .catch((e) => setError(e.message))
  }, [])

  async function loginPlayer(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      const data = await api.login(Number(playerId), pin.trim())
      setUser(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  async function loginAdmin(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      const data = await api.adminLogin(pin.trim())
      setUser(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="login">
      <div className="login-card">
        <div className="login-logo" aria-hidden="true">⚽</div>
        <h1>Polla Mundial 2026</h1>

        {mode === 'player' ? (
          <form onSubmit={loginPlayer}>
            <p className="login-help">Elige tu nombre y escribe tu PIN de 4 dígitos.</p>
            <label className="field">
              <span>Tu nombre</span>
              <select value={playerId} onChange={(e) => setPlayerId(e.target.value)} required>
                <option value="">— Elige tu nombre —</option>
                {(players || []).map((p) => (
                  <option key={p.id} value={p.id}>{p.name}</option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>Tu PIN</span>
              <input
                className="pin-input"
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                autoComplete="off"
                autoCorrect="off"
                value={pin}
                onChange={(e) => setPin(e.target.value.replace(/\D/g, '').slice(0, 4))}
                placeholder="••••"
                maxLength={4}
                required
              />
            </label>
            {error && <p className="error">{error}</p>}
            <button className="primary-btn" disabled={busy || !playerId}>
              {busy ? 'Entrando…' : 'Entrar'}
            </button>
            <button type="button" className="link-btn" onClick={() => { setMode('admin'); setError(''); setPin('') }}>
              Soy administrador
            </button>
          </form>
        ) : (
          <form onSubmit={loginAdmin}>
            <p className="login-help">Escribe el PIN de administrador.</p>
            <label className="field">
              <span>PIN de administrador</span>
              <input
                type="password"
                value={pin}
                onChange={(e) => setPin(e.target.value)}
                placeholder="••••••"
                required
                autoFocus
              />
            </label>
            {error && <p className="error">{error}</p>}
            <button className="primary-btn" disabled={busy}>
              {busy ? 'Entrando…' : 'Entrar como admin'}
            </button>
            <button type="button" className="link-btn" onClick={() => { setMode('player'); setError(''); setPin('') }}>
              Volver
            </button>
          </form>
        )}
      </div>
    </div>
  )
}
