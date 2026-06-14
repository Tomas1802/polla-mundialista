import { createContext, useContext, useEffect, useState, useCallback } from 'react'
import { api } from './api.js'

const AuthContext = createContext(null)

export function useAuth() {
  return useContext(AuthContext)
}

// The session object is { isAdmin, playerId, playerName, mustChangePin }.
export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  const refresh = useCallback(async () => {
    try {
      const data = await api.me()
      setUser(data)
    } catch {
      setUser(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    refresh()
  }, [refresh])

  const logout = useCallback(async () => {
    try {
      await api.logout()
    } catch {
      // ignore; clear local session regardless
    }
    setUser(null)
  }, [])

  return (
    <AuthContext.Provider value={{ user, setUser, loading, refresh, logout }}>
      {children}
    </AuthContext.Provider>
  )
}
