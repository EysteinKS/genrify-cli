// Action history dialog - shows all logged actions with expandable details

import { useState } from 'react'
import type { ActionLogEntry } from '@/contexts/StatusBarContext'
import styles from './ActionHistoryDialog.module.css'

function formatJSON(value: unknown): string {
  if (value === undefined || value === null) return ''
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

function formatTimestamp(timestamp: number): string {
  const date = new Date(timestamp)
  return date.toLocaleString()
}

function EntryDetails({ entry }: { entry: ActionLogEntry }) {
  const [expanded, setExpanded] = useState(false)
  
  const hasDetails = entry.variables !== undefined || entry.data !== undefined || entry.error !== undefined

  return (
    <div className={`${styles.entry} ${styles[entry.type]}`}>
      <div className={styles.entryHeader}>
        <div className={styles.entryInfo}>
          <span className={styles.entryIcon}>{entry.type === 'success' ? '✓' : '✕'}</span>
          <div className={styles.entryMain}>
            <div className={styles.entryTitle}>{entry.title}</div>
            {entry.message && <div className={styles.entryMessage}>{entry.message}</div>}
          </div>
        </div>
        <div className={styles.entryMeta}>
          <span className={styles.entryTime}>{formatTimestamp(entry.timestamp)}</span>
          {hasDetails && (
            <button
              className={styles.expandButton}
              onClick={() => setExpanded(!expanded)}
              title={expanded ? 'Collapse details' : 'Expand details'}
            >
              {expanded ? '▼' : '▶'}
            </button>
          )}
        </div>
      </div>

      {expanded && hasDetails && (
        <div className={styles.entryDetails}>
          {entry.variables !== undefined && (
            <details className={styles.detailsSection} open>
              <summary className={styles.detailsSummary}>Variables</summary>
              <pre className={styles.detailsCode}>{formatJSON(entry.variables)}</pre>
            </details>
          )}

          {entry.data !== undefined && (
            <details className={styles.detailsSection} open>
              <summary className={styles.detailsSummary}>Result Data</summary>
              <pre className={styles.detailsCode}>{formatJSON(entry.data)}</pre>
            </details>
          )}

          {entry.error !== undefined && (
            <details className={styles.detailsSection} open>
              <summary className={styles.detailsSummary}>Error Details</summary>
              <pre className={styles.detailsCode}>
                {entry.error instanceof Error 
                  ? `${entry.error.message}\n\n${entry.error.stack || ''}` 
                  : String(entry.error)}
              </pre>
            </details>
          )}
        </div>
      )}
    </div>
  )
}

export function ActionHistoryDialog({
  entries,
  onClose,
}: {
  entries: ActionLogEntry[]
  onClose: () => void
}) {
  return (
    <div className={styles.overlay} role="dialog" aria-modal="true" onClick={onClose}>
      <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>Action History</h2>
          <button
            className={styles.closeButton}
            onClick={onClose}
            aria-label="Close"
            title="Close"
          >
            ×
          </button>
        </div>

        <div className={styles.content}>
          {entries.length === 0 ? (
            <div className={styles.empty}>No actions logged yet</div>
          ) : (
            <div className={styles.entries}>
              {entries.map((entry) => (
                <EntryDetails key={entry.id} entry={entry} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
