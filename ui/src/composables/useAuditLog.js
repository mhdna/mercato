import { API_BASE } from '@/config'
import { createResource } from './useApiResource'

const resource = createResource({ path: '/audit_logs', listPath: '/audit_logs', rootKey: 'audit_logs' })

export function useAuditLog () {
  return {
    auditLogs: resource.items,
    fetchAuditLogs: resource.fetchAll,
  }
}

export const AUDIT_LOG_API_URL = `${API_BASE}/audit_logs`
