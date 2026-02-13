// Main layout component with grid layout

import { Outlet } from 'react-router-dom'
import { Header } from './Header'
import { Sidebar } from './Sidebar'
import { StatusBar } from './StatusBar'
import { ToastHost } from '../ToastHost/ToastHost'
import { ActionHistoryDialog } from '../ActionHistoryDialog/ActionHistoryDialog'
import { useStatusBar } from '@/contexts/StatusBarContext'
import styles from './Layout.module.css'

export function Layout() {
  const { entries, isHistoryOpen, closeHistory } = useStatusBar()

  return (
    <div className={styles.layout}>
      <Header />
      <Sidebar />
      <main className={styles.content}>
        <Outlet />
      </main>
      <StatusBar />
      <ToastHost />
      {isHistoryOpen && <ActionHistoryDialog entries={entries} onClose={closeHistory} />}
    </div>
  )
}
