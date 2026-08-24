import { useAuthStore } from '@/stores/auth'

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
    const headers = { ...(options.headers || {}) }
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }
    return fetch(url, { ...options, headers })
  }

  let res = await request(authStore.token)

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
