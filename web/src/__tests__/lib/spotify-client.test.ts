import { beforeEach, describe, expect, it, vi } from 'vitest'
import { SpotifyClient, withWriteAccess } from '@/lib/spotify-client'

function mockFetchOnceJson(payload: unknown) {
  const resp: Pick<Response, 'ok' | 'status' | 'headers' | 'text'> = {
    ok: true,
    status: 200,
    headers: new Headers(),
    text: async () => JSON.stringify(payload),
  }

  return vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => resp as unknown as Response)
}

describe('SpotifyClient write guard', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
  })

  it('allows GET requests without write scope', async () => {
    const fetchMock = mockFetchOnceJson({ id: 'me', display_name: 'Me' })
    vi.stubGlobal('fetch', fetchMock)

    const client = new SpotifyClient(
      async () => 'token',
      async () => 'token'
    )

    const me = await client.getMe()

    expect(me.id).toBe('me')
    expect(fetchMock).toHaveBeenCalledOnce()
    expect(fetchMock).toHaveBeenCalledWith(
      'https://api.spotify.com/v1/me',
      expect.objectContaining({ method: 'GET' })
    )
  })

  it('blocks non-GET requests outside write scope', async () => {
    const fetchMock = mockFetchOnceJson({})
    vi.stubGlobal('fetch', fetchMock)

    const client = new SpotifyClient(
      async () => 'token',
      async () => 'token'
    )

    await expect(client.createPlaylist('My Playlist', '', false)).rejects.toThrow(
      /Blocked Spotify API write \(POST\)/
    )

    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('allows non-GET requests inside write scope', async () => {
    const fetchMock = mockFetchOnceJson({ id: 'pl1', name: 'My Playlist' })
    vi.stubGlobal('fetch', fetchMock)

    const client = new SpotifyClient(
      async () => 'token',
      async () => 'token'
    )

    const pl = await withWriteAccess(() => client.createPlaylist('My Playlist', '', false))

    expect(pl.id).toBe('pl1')
    expect(fetchMock).toHaveBeenCalledOnce()
    expect(fetchMock).toHaveBeenCalledWith(
      'https://api.spotify.com/v1/me/playlists',
      expect.objectContaining({ method: 'POST' })
    )
  })
})
