// Redirect URI helpers for the browser OAuth flow.

const LOOPBACK_HOSTS = new Set(['localhost', '127.0.0.1', '::1'])

export function isLoopbackUrl(value: string): boolean {
  try {
    const url = new URL(value)
    return LOOPBACK_HOSTS.has(url.hostname)
  } catch {
    return false
  }
}

function normalizeBaseUrl(baseUrl: string | undefined): string {
  const raw = (baseUrl || '/').trim()
  if (raw === '' || raw === '/') return ''
  // Ensure leading slash, remove trailing slashes.
  const withLeading = raw.startsWith('/') ? raw : `/${raw}`
  return withLeading.replace(/\/+$/, '')
}

export function getAppCallbackRedirectUri(): string {
  if (typeof window === 'undefined') return ''

  const base = normalizeBaseUrl(import.meta.env.BASE_URL)
  const path = `${base}/callback`

  return new URL(path, window.location.origin).toString()
}

/**
 * For the web app we generally want the redirect URI to point back to this
 * running origin + (base) + /callback. If the configured redirect URI is a
 * loopback URL but with the wrong port/path (common when copied from the CLI),
 * we override it to the app callback URL to avoid dead callbacks.
 */
export function getEffectiveRedirectUri(configuredRedirectUri: string): string {
  const appRedirectUri = getAppCallbackRedirectUri()
  const configured = configuredRedirectUri.trim()

  if (!configured) {
    return appRedirectUri || configuredRedirectUri
  }

  let configuredUrl: URL
  try {
    configuredUrl = new URL(configured)
  } catch {
    return appRedirectUri || configuredRedirectUri
  }

  if (!appRedirectUri) {
    return configuredUrl.toString()
  }

  const appUrl = new URL(appRedirectUri)

  const configuredIsLoopback = LOOPBACK_HOSTS.has(configuredUrl.hostname)
  if (configuredIsLoopback) {
    // Common case: user copied a loopback redirect URI from the CLI config.
    // In the browser we always want Spotify to redirect back to *this* app.
    return appUrl.toString()
  }

  // If a non-loopback redirect URI was configured but doesn't match the current app,
  // redirecting there won't hit this SPA's /callback handler. Prefer the app callback.
  if (configuredUrl.origin !== appUrl.origin || configuredUrl.pathname !== appUrl.pathname) {
    return appUrl.toString()
  }

  return configuredUrl.toString()
}
