// Status bar context - manages global status messages, errors, and loading state

import { createContext, useContext, useState, useCallback, useEffect, ReactNode } from 'react'

interface StatusBarContextValue {
  message: string
  isError: boolean
  isLoading: boolean
  setStatus: (message: string) => void
  setError: (message: string) => void
  setLoading: (message: string) => void
  clear: () => void
}

const StatusBarContext = createContext<StatusBarContextValue | null>(null)

export function StatusBarProvider({ children }: { children: ReactNode }) {
  const [message, setMessage] = useState('')
  const [isError, setIsError] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const clear = useCallback(() => {
    setMessage('')
    setIsError(false)
    setIsLoading(false)
  }, [])

  const setStatus = useCallback((msg: string) => {
    setMessage(msg)
    setIsError(false)
    setIsLoading(false)
  }, [])

  const setError = useCallback((msg: string) => {
    setMessage(msg)
    setIsError(true)
    setIsLoading(false)
  }, [])

  const setLoading = useCallback((msg: string) => {
    setMessage(msg)
    setIsError(false)
    setIsLoading(true)
  }, [])

  // Auto-clear errors after 10 seconds
  useEffect(() => {
    if (isError && message) {
      const timer = setTimeout(() => {
        clear()
      }, 10000)
      return () => clearTimeout(timer)
    }
  }, [isError, message, clear])

  return (
    <StatusBarContext.Provider
      value={{
        message,
        isError,
        isLoading,
        setStatus,
        setError,
        setLoading,
        clear,
      }}
    >
      {children}
    </StatusBarContext.Provider>
  )
}

export function useStatusBar() {
  const context = useContext(StatusBarContext)
  if (!context) {
    throw new Error('useStatusBar must be used within StatusBarProvider')
  }
  return context
}
