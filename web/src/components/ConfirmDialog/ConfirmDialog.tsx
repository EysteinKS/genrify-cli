// Confirmation dialog for write actions

import type { ConfirmActionPlan, ConfirmRequest } from '@/contexts/ConfirmContext'
import styles from './ConfirmDialog.module.css'

function formatJSON(value: unknown): string {
  if (value === undefined) return ''
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

function RequestDetails({ req }: { req: ConfirmRequest }) {
  const body = formatJSON(req.body)

  return (
    <div className={styles.request}>
      <div className={styles.requestHeader}>
        <span className={styles.method}>{req.method}</span>
        <span className={styles.path}>{req.path}</span>
      </div>

      {req.description && <div className={styles.requestDescription}>{req.description}</div>}

      {req.body !== undefined && (
        <details className={styles.details}>
          <summary className={styles.summary}>View exact JSON payload</summary>
          <pre className={styles.code}>{body}</pre>
        </details>
      )}
    </div>
  )
}

export function ConfirmDialog({
  plan,
  onClose,
}: {
  plan: ConfirmActionPlan
  onClose: (confirmed: boolean) => void
}) {
  return (
    <div className={styles.overlay} role="dialog" aria-modal="true">
      <div className={styles.dialog}>
        <div className={styles.header}>
          <h2 className={styles.title}>{plan.title}</h2>
          <button
            className={styles.closeButton}
            onClick={() => onClose(false)}
            aria-label="Close"
            title="Close"
          >
            ×
          </button>
        </div>

        <div className={styles.content}>
          <p className={styles.intro}>{plan.intro || 'This action will send the following data to Spotify:'}</p>

          {plan.summary && plan.summary.length > 0 && (
            <ul className={styles.summaryList}>
              {plan.summary.map((line, idx) => (
                <li key={idx}>{line}</li>
              ))}
            </ul>
          )}

          <div className={styles.requests}>
            {plan.requests.map((req, idx) => (
              <RequestDetails key={idx} req={req} />
            ))}
          </div>
        </div>

        <div className={styles.actions}>
          <button className={styles.cancelButton} onClick={() => onClose(false)}>
            Cancel
          </button>
          <button className={styles.confirmButton} onClick={() => onClose(true)}>
            {plan.confirmLabel || 'Confirm'}
          </button>
        </div>
      </div>
    </div>
  )
}
