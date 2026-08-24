import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { API_BASE } from '@/config'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const refreshToken = ref(localStorage.getItem('refreshToken') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAuthenticated = computed(() => !!token.value)

  async function login (name, password) {
    const res = await fetch(`${API_BASE}/users/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, password }),
    })

    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new Error(body.error || 'Invalid credentials')
    }

    const data = await res.json()
    token.value = data.access_token
    refreshToken.value = data.refresh_token
    user.value = data.user
    localStorage.setItem('token', token.value)
    localStorage.setItem('refreshToken', refreshToken.value)
    localStorage.setItem('user', JSON.stringify(user.value))
    return true
  }

  async function renew () {
    if (!refreshToken.value) {
      throw new Error('No refresh token')
    }

    const res = await fetch(`${API_BASE}/tokens/renew_access`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken.value }),
    })

    if (!res.ok) {
      throw new Error('Session expired')
    }

    const data = await res.json()
    token.value = data.access_token
    localStorage.setItem('token', token.value)
    return token.value
  }

  function logout () {
    token.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('user')
  }

  return { token, refreshToken, user, isAuthenticated, login, renew, logout }
})
