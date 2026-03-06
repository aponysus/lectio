import {
  createContext,
  type ReactNode,
  useContext,
  useEffect,
  useState,
} from 'react'
import * as api from '../api/client'

type SessionContextValue = {
  loading: boolean
  session: api.SessionState
  refresh: () => Promise<void>
  login: (password: string) => Promise<void>
  logout: () => Promise<void>
}

const defaultSession: api.SessionState = {
  authenticated: false,
  csrf_token: '',
}

const SessionContext = createContext<SessionContextValue | null>(null)

export function SessionProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<api.SessionState>(defaultSession)
  const [loading, setLoading] = useState(true)

  const refresh = async () => {
    const nextSession = await api.getSession()
    setSession(nextSession)
  }

  const login = async (password: string) => {
    const nextSession = await api.login(password)
    setSession(nextSession)
  }

  const logout = async () => {
    const nextSession = await api.logout()
    setSession(nextSession)
  }

  useEffect(() => {
    let cancelled = false

    ;(async () => {
      try {
        const nextSession = await api.getSession()
        if (!cancelled) {
          setSession(nextSession)
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    })()

    return () => {
      cancelled = true
    }
  }, [])

  return (
    <SessionContext.Provider
      value={{
        loading,
        session,
        refresh,
        login,
        logout,
      }}
    >
      {children}
    </SessionContext.Provider>
  )
}

export function useSession() {
  const value = useContext(SessionContext)
  if (!value) {
    throw new Error('useSession must be used inside SessionProvider')
  }
  return value
}
