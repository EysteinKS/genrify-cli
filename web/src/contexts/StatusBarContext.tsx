// Status bar context - manages global status messages, errors, loading state, action history, and toasts

import { createContext, useContext, useState, useCallback, useEffect, ReactNode } from 'react'

export interface ActionLogEntry {
  id: string
  timestamp: number
  type: 'success' | 'error'
  title: string
  message?: string
  variables?: unknown
  data?: unknown
  error?: Error | string
}

export interface Toast {
  id: string
  type: 'success' | 'error'
  message: string
}

interface StatusBarContextValue {
  message: string
  isError: boolean
  isLoading: boolean
  setStatus: (message: string) => void
  setError: (message: string) => void
  setLoading: (message: string) => void
  clear: () => void
  // Action log
  entries: ActionLogEntry[]
  logSuccess: (title: string, opts?: { message?: string; variables?: unknown; data?: unknown }) => void
  logError: (title: string, opts?: { message?: string; variables?: unknown; error?: Error | string }) => void
  // History dialog
  isHistoryOpen: boolean
  openHistory: () => void
  closeHistory: () => void
  // Toasts
  toasts: Toast[]
  dismissToast: (id: string) => void
}

const StatusBarContext = createContext<StatusBarContextValue | null>(null)

export function StatusBarProvider({ children }: { children: ReactNode }) {
  const [message, setMessage] = useState('')
  const [isError, setIsError] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [entries, setEntries] = useState<ActionLogEntry[]>([])
  const [isHistoryOpen, setIsHistoryOpen] = useState(false)
  const [toasts, setToasts] = useState<Toast[]>([])

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

  const logSuccess = useCallback((title: string, opts?: { message?: string; variables?: unknown; data?: unknown }) => {
    const entry: ActionLogEntry = {
      id: `${Date.now()}-${Math.random()}`,
      timestamp: Date.now(),
      type: 'success',
      title,
      message: opts?.message,
      variables: opts?.variables,
      data: opts?.data,
    }
    setEntries((prev) => [entry, ...prev])
    
    // Add toast
    const toast: Toast = {
      id: entry.id,
      type: 'success',
      message: opts?.message || title,
    }
    setToasts((prev) => [...prev, toast])
    
    // Auto-dismiss toast after 5 seconds
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== toast.id))
    }, 5000)
  }, [])

  const logError = useCallback((title: string, opts?: { message?: string; variables?: unknown; error?: Error | string }) => {
    const entry: ActionLogEntry = {
      id: `${Date.now()}-${Math.random()}`,
      timestamp: Date.now(),
      type: 'error',
      title,
      message: opts?.message,
      variables: opts?.variables,
      error: opts?.error,
    }
    setEntries((prev) => [entry, ...prev])
    
    // Add toast
    const toast: Toast = {
      id: entry.id,
      type: 'error',
      message: opts?.message || title,
    }
    setToasts((prev) => [...prev, toast])
    
    // Auto-dismiss error toast after 10 seconds
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== toast.id))
    }, 10000)
  }, [])

  const openHistory = useCallback(() => {
    setIsHistoryOpen(true)
  }, [])

  const closeHistory = useCallback(() => {
    setIsHistoryOpen(false)
  }, [])

  const dismissToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
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
        entries,
        logSuccess,
        logError,
        isHistoryOpen,
        openHistory,
        closeHistory,
        toasts,
        dismissToast,
      }}
    >
      {children}
    </StatusBarContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useStatusBar() {
  const context = useContext(StatusBarContext)
  if (!context) {
    throw new Error('useStatusBar must be used within StatusBarProvider')
  }
  return context
}
