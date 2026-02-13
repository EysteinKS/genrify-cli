// Main app component with routing

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { ConfigProvider } from './contexts/ConfigContext'
import { AuthProvider } from './contexts/AuthContext'
import { StatusBarProvider } from './contexts/StatusBarContext'
import { Layout } from './components/Layout/Layout'
import { LoginPage } from './pages/LoginPage/LoginPage'
import { CallbackPage } from './pages/CallbackPage/CallbackPage'
import { PlaylistsPage } from './pages/PlaylistsPage/PlaylistsPage'
import { TracksPage } from './pages/TracksPage/TracksPage'
import { CreatePage } from './pages/CreatePage/CreatePage'
import { AddTracksPage } from './pages/AddTracksPage/AddTracksPage'
import { MergePage } from './pages/MergePage/MergePage'

// Configure TanStack Query
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false, // Client handles its own retries
      refetchOnWindowFocus: false,
    },
  },
})

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider>
        <AuthProvider>
          <StatusBarProvider>
            <BrowserRouter>
              <Routes>
                {/* Callback sits outside Layout (no sidebar during OAuth) */}
                <Route path="/callback" element={<CallbackPage />} />

                {/* All other routes use Layout */}
                <Route element={<Layout />}>
                  <Route path="/login" element={<LoginPage />} />
                  <Route path="/playlists" element={<PlaylistsPage />} />
                  <Route path="/tracks" element={<TracksPage />} />
                  <Route path="/create" element={<CreatePage />} />
                  <Route path="/add-tracks" element={<AddTracksPage />} />
                  <Route path="/merge" element={<MergePage />} />
                  <Route path="/" element={<Navigate to="/login" replace />} />
                </Route>
              </Routes>
            </BrowserRouter>
          </StatusBarProvider>
        </AuthProvider>
      </ConfigProvider>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  )
}
