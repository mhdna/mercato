import { useAuthStore } from '@/stores/auth'

// Thrown by authFetch when the server can't be reached at all (powered off,
// wrong host, no network). ErrorBoundary matches on this exact message to swap
// a blank screen for a friendly "server is down" state, so keep them in sync.
export const SERVER_DOWN_MESSAGE = 'Server is down, we\'ll be back soon.'

let refreshPromise = null

function renewOnce (authStore) {
  if (!refreshPromise) {
    refreshPromise = authStore.renew().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

export async function authFetch (url, options = {}) {
  const authStore = useAuthStore()

  const request = token => {
    const headers = { ...options.headers }
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }
    return fetch(url, { ...options, headers })
  }

  let res
  try {
    res = await request(authStore.token)
  } catch {
    // fetch() rejects (rather than resolving with a bad status) when the
    // server can't be reached at all -- e.g. it's powered off. Surface a
    // clear, distinct message instead of the browser's raw "Failed to fetch".
    throw new Error(SERVER_DOWN_MESSAGE)
  }

  if (res.status === 401 && authStore.refreshToken) {
    try {
      const newToken = await renewOnce(authStore)
      res = await request(newToken)
    } catch {
      // refresh failed, fall through to the logout/redirect below
    }
  }

  if (res.status === 401) {
    authStore.logout()
    if (window.location.pathname !== '/login') {
      window.location.assign('/login')
    }
  }

  return res
}
