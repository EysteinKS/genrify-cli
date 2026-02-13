// Entry point

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './App'
import './globals.css'

// Handle GitHub Pages SPA redirect (from 404.html)
const redirect = sessionStorage.getItem('redirect')
if (redirect) {
  sessionStorage.removeItem('redirect')
  history.replaceState(null, '', redirect)
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
)
