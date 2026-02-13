// Confirm context - provides a global confirmation dialog for write actions

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'
import { ConfirmDialog } from '@/components/ConfirmDialog/ConfirmDialog'

export type ConfirmHttpMethod = 'POST' | 'PUT' | 'PATCH' | 'DELETE'

export interface ConfirmRequest {
  method: ConfirmHttpMethod
  path: string
  description?: string
  body?: unknown
}

export interface ConfirmActionPlan {
  title: string
  intro?: string
  summary?: string[]
  requests: ConfirmRequest[]
  confirmLabel?: string
}

interface ConfirmContextValue {
  confirm: (plan: ConfirmActionPlan) => Promise<boolean>
}

const ConfirmContext = createContext<ConfirmContextValue | null>(null)

interface PendingConfirmation {
  plan: ConfirmActionPlan
  resolve: (value: boolean) => void
}

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [pending, setPending] = useState<PendingConfirmation | null>(null)

  const confirm = useCallback((plan: ConfirmActionPlan): Promise<boolean> => {
    return new Promise<boolean>((resolve) => {
      setPending({ plan, resolve })
    })
  }, [])

  const handleClose = useCallback(
    (confirmed: boolean) => {
      if (!pending) return
      pending.resolve(confirmed)
      setPending(null)
    },
    [pending]
  )

  const value = useMemo(() => ({ confirm }), [confirm])

  return (
    <ConfirmContext.Provider value={value}>
      {children}
      {pending && <ConfirmDialog plan={pending.plan} onClose={handleClose} />}
    </ConfirmContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useConfirm() {
  const ctx = useContext(ConfirmContext)
  if (!ctx) {
    throw new Error('useConfirm must be used within a ConfirmProvider')
  }
  return ctx
}
