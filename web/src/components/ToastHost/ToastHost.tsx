// Toast notifications component - displays success/error toasts below header

import { useStatusBar } from '@/contexts/StatusBarContext'
import styles from './ToastHost.module.css'

export function ToastHost() {
  const { toasts, dismissToast, openHistory } = useStatusBar()

  const handleToastClick = (toastId: string) => {
    dismissToast(toastId)
    openHistory()
  }

  return (
    <div className={styles.toastContainer}>
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`${styles.toast} ${styles[toast.type]}`}
          onClick={() => handleToastClick(toast.id)}
          role="alert"
        >
          <span className={styles.toastIcon}>
            {toast.type === 'success' ? '✓' : '✕'}
          </span>
          <span className={styles.toastMessage}>{toast.message}</span>
          <button
            className={styles.toastDismiss}
            onClick={(e) => {
              e.stopPropagation()
              dismissToast(toast.id)
            }}
            aria-label="Dismiss"
          >
            ×
          </button>
        </div>
      ))}
    </div>
  )
}
