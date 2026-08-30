export const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8090'
export const WS_BASE = API_BASE.replace(/^http/, 'ws')

// Temporary responsive gate. Set VITE_BLOCK_NON_DESKTOP=false to restore
// access on mobile and tablet layouts without removing the unsupported view.
export const BLOCK_NON_DESKTOP = import.meta.env.VITE_BLOCK_NON_DESKTOP !== 'false'
