// Status bar component at bottom of layout

import { useStatusBar } from '@/contexts/StatusBarContext'
import styles from './StatusBar.module.css'

export function StatusBar() {
  const { message, isError, isLoading } = useStatusBar()

  if (!message) {
    return <div className={styles.statusBar} />
  }

  return (
    <div className={`${styles.statusBar} ${isError ? styles.error : ''}`}>
      {isLoading && <span className={styles.spinner} />}
      <span className={styles.message}>{message}</span>
    </div>
  )
}
