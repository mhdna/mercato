import { API_BASE } from '@/config'
import { authFetch } from './useApi'

async function requestJSON (url, options) {
  const res = await authFetch(url, options)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  return res.json().catch(() => null)
}

// Employees are a company-wide HR roster with a single current payroll
// snapshot each. The API recomputes deduction_cents (rate x late units) and
// net_salary_cents (base + commission - deduction) on every read -- the UI
// never has to trust its own arithmetic once a save round-trips.
export function useEmployees () {
  async function listEmployees (params = {}) {
    const url = new URL(`${API_BASE}/employees`)
    url.searchParams.set('page_size', params.pageSize ?? 100)
    url.searchParams.set('page_id', params.pageId ?? 0)
    if (params.search) {url.searchParams.set('search', params.search)}
    if (params.branchId) {url.searchParams.set('branch_id', params.branchId)}
    const data = await requestJSON(url.toString())
    return { employees: data?.employees ?? [], total: data?.total ?? 0 }
  }

  async function createEmployee (payload) {
    const data = await requestJSON(`${API_BASE}/employees`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return data?.employee ?? null
  }

  async function updateEmployee (id, payload) {
    const data = await requestJSON(`${API_BASE}/employees/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return data?.employee ?? null
  }

  async function deleteEmployee (id) {
    await requestJSON(`${API_BASE}/employees/${id}`, { method: 'DELETE' })
  }

  // items: [{ id, base_salary_cents, commission_cents,
  //           late_deduction_rate_cents, late_deduction_unit, late_units_centi }]
  async function batchSetSalaries (items) {
    const data = await requestJSON(`${API_BASE}/employees/salaries`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items }),
    })
    return data?.employees ?? []
  }

  return { listEmployees, createEmployee, updateEmployee, deleteEmployee, batchSetSalaries }
}
