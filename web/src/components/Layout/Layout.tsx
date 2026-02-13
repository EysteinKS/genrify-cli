// Main layout component with grid layout

import { Outlet } from 'react-router-dom'
import { Header } from './Header'
import { Sidebar } from './Sidebar'
import { StatusBar } from './StatusBar'
import styles from './Layout.module.css'

export function Layout() {
  return (
    <div className={styles.layout}>
      <Header />
      <Sidebar />
      <main className={styles.content}>
        <Outlet />
      </main>
      <StatusBar />
    </div>
  )
}
