import { authFetch } from './useApi'

const inflight = new Map()

export function dedupedFetch (url) {
  if (inflight.has(url)) {
    return inflight.get(url)
  }

  const promise = authFetch(url)
    .then(async r => {
      if (!r.ok) {
        const body = await r.json().catch(() => ({}))
        throw new Error(body.error || `Request failed with status ${r.status}`)
      }
      return r.json()
    })
    .finally(() => {
      setTimeout(() => inflight.delete(url), 50)
    })

  inflight.set(url, promise)
  return promise
}
