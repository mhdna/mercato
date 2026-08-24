import { ref } from 'vue'
import { WS_BASE } from '@/config'
import { useAuthStore } from '@/stores/auth'

// Module-scope (not inside useAdminSocket) so every component that calls
// this composable shares one connection instead of each opening its own --
// SyncCard and the branch-activity toast queue both need the same stream of
// admin pushes, and there's no reason for a browser tab to hold two sockets
// to the same endpoint.
const status = ref('connecting')
const listeners = new Set()
let ws = null
let reconnectTimer = null

function handleMessage (event) {
  let message
  try {
    message = JSON.parse(event.data)
  } catch {
    return // Ignore malformed pushes -- consumers have their own poll fallback.
  }
  for (const listener of listeners) {
    listener(message)
  }
}

function handleClose () {
  status.value = 'closed'
  clearTimeout(reconnectTimer)
  reconnectTimer = setTimeout(connect, 5000)
}

function connect () {
  const authStore = useAuthStore()

  status.value = 'connecting'
  ws = new WebSocket(`${WS_BASE}/admin/ws?token=${authStore.token}`)

  ws.addEventListener('open', () => {
    status.value = 'open'
  })
  ws.addEventListener('message', handleMessage)
  ws.addEventListener('close', handleClose)
  ws.addEventListener('error', () => {
    ws.close()
  })
}

export function useAdminSocket () {
  function ensureConnected () {
    if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
      return
    }
    connect()
  }

  function onMessage (callback) {
    listeners.add(callback)
    return () => listeners.delete(callback)
  }

  return { status, ensureConnected, onMessage }
}
