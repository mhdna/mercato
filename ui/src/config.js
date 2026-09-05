export const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8090'
export const WS_BASE = API_BASE.replace(/^http/, 'ws')

// Temporary device gate: blocks mobile/tablet user agents (see App.vue),
// not narrow viewports. Set VITE_BLOCK_NON_DESKTOP=false to restore access
// on those devices without removing the unsupported view.
export const BLOCK_NON_DESKTOP = import.meta.env.VITE_BLOCK_NON_DESKTOP !== 'false'
