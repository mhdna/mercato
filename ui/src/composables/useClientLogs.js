import { computed, readonly, ref } from 'vue'

const MAX_ENTRIES = 250
const MAX_MESSAGE_LENGTH = 4000
const entries = ref([])
let nextId = 1
let installed = false

function stringify (value) {
  if (value instanceof Error) {return value.stack || value.message}
  if (typeof value === 'string') {return value}

  try {
    const seen = new WeakSet()
    return JSON.stringify(value, (key, item) => {
      if (typeof item === 'object' && item !== null) {
        if (seen.has(item)) {return '[Circular]'}
        seen.add(item)
      }
      return item
    })
  } catch {
    return String(value)
  }
}

function addLog (level, args, source = 'console') {
  const message = args.map(value => stringify(value)).join(' ').slice(0, MAX_MESSAGE_LENGTH)
  const entry = {
    id: nextId++,
    group: 'browser',
    level,
    message: message || '(empty message)',
    source,
    timestamp: new Date().toISOString(),
  }

  // Replace the array in one operation and keep a hard cap. This prevents a
  // noisy session from retaining an ever-growing log history.
  entries.value = [...entries.value.slice(-(MAX_ENTRIES - 1)), entry]
}

export function installClientLogger () {
  if (installed) {return}
  installed = true

  for (const level of ['log', 'info', 'warn', 'error']) {
    const original = console[level].bind(console)
    console[level] = (...args) => {
      original(...args)
      addLog(level === 'log' ? 'info' : level, args)
    }
  }

  window.addEventListener('error', event => {
    addLog('error', [event.error || event.message], 'window')
  })
  window.addEventListener('unhandledrejection', event => {
    addLog('error', [event.reason || 'Unhandled promise rejection'], 'promise')
  })

  addLog('info', ['Application started'], 'system')
}

export function useClientLogs () {
  return {
    entries: readonly(entries),
    count: computed(() => entries.value.length),
    capacity: MAX_ENTRIES,
    clear: () => { entries.value = [] },
  }
}
