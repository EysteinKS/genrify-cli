// Sidebar navigation component

import { NavLink } from 'react-router-dom'
import styles from './Sidebar.module.css'

export function Sidebar() {
  const links = [
    { to: '/login', label: 'Login' },
    { to: '/playlists', label: 'Playlists' },
    { to: '/tracks', label: 'Tracks' },
    { to: '/create', label: 'Create Playlist' },
    { to: '/add-tracks', label: 'Add Tracks' },
    { to: '/merge', label: 'Merge Playlists' },
  ]

  return (
    <aside className={styles.sidebar}>
      <nav className={styles.nav}>
        {links.map((link) => (
          <NavLink
            key={link.to}
            to={link.to}
            className={({ isActive }) =>
              isActive ? `${styles.link} ${styles.active}` : styles.link
            }
          >
            {link.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  )
}
