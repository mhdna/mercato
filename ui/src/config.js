export const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8090'
export const WS_BASE = API_BASE.replace(/^http/, 'ws')
