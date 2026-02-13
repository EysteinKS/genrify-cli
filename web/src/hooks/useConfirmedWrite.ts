import { useCallback } from 'react'
import { useConfirm, type ConfirmActionPlan } from '@/contexts/ConfirmContext'
import { useStatusBar } from '@/contexts/StatusBarContext'
import { withWriteAccess } from '@/lib/spotify-client'
import { CancelledError } from '@/lib/cancelled'

type Message<T> = string | ((value: T) => string)

function resolveMessage<T>(msg: Message<T> | undefined, value: T): string | undefined {
  if (!msg) return undefined
  return typeof msg === 'function' ? msg(value) : msg
}

export function useConfirmedWrite() {
  const { confirm } = useConfirm()
  const { setLoading, setStatus, setError } = useStatusBar()

  const confirmAndRun = useCallback(
    async <T>(opts: {
      plan: ConfirmActionPlan
      startingMessage: string
      successMessage?: Message<T>
      errorPrefix?: string
      action: () => Promise<T>
    }): Promise<T> => {
      const ok = await confirm(opts.plan)
      if (!ok) {
        setStatus('Cancelled')
        throw new CancelledError()
      }

      setLoading(opts.startingMessage)

      try {
        const res = await withWriteAccess(opts.action)
        const successMsg = resolveMessage(opts.successMessage, res)
        if (successMsg) {
          setStatus(successMsg)
        }
        return res
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err)
        setError(opts.errorPrefix ? `${opts.errorPrefix}: ${msg}` : msg)
        throw err
      }
    },
    [confirm, setError, setLoading, setStatus]
  )

  return { confirmAndRun }
}
