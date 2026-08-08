import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as authApi from '../api/auth'
import type { User } from '../api/auth'
import { ApiError } from '../api/client'
import { getToken, setToken } from '../api/client'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const user = ref<User | null>(null)
  const bootstrapped = ref(false)
  const loading = ref(false)

  const isAuthenticated = computed(() => Boolean(token.value))

  function applySession(nextToken: string, nextUser: User) {
    token.value = nextToken
    user.value = nextUser
    setToken(nextToken)
  }

  function clearSession() {
    token.value = null
    user.value = null
    setToken(null)
  }

  async function bootstrap() {
    if (bootstrapped.value) {
      return
    }
    bootstrapped.value = true

    if (!token.value) {
      return
    }

    try {
      user.value = await authApi.fetchMe()
    } catch (err) {
      if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
        console.info('Stored session expired; clearing token')
        clearSession()
      } else {
        console.warn('Failed to restore session', err)
      }
    }
  }

  async function register(email: string, password: string) {
    loading.value = true
    try {
      const result = await authApi.register(email, password)
      applySession(result.token, result.user)
      return result
    } finally {
      loading.value = false
    }
  }

  async function login(email: string, password: string) {
    loading.value = true
    try {
      const result = await authApi.login(email, password)
      applySession(result.token, result.user)
      return result
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    loading.value = true
    try {
      if (token.value) {
        try {
          await authApi.logout()
        } catch (err) {
          console.warn('Logout request failed; clearing local session anyway', err)
        }
      }
    } finally {
      clearSession()
      loading.value = false
    }
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    loading.value = true
    try {
      const result = await authApi.changePassword(oldPassword, newPassword)
      if (result && typeof result === 'object' && 'token' in result) {
        const authResult = result as authApi.AuthResponse
        if (typeof authResult.token === 'string' && authResult.user) {
          applySession(authResult.token, authResult.user)
        } else if (typeof authResult.token === 'string') {
          token.value = authResult.token
          setToken(authResult.token)
          user.value = await authApi.fetchMe()
        }
      }
      return result
    } finally {
      loading.value = false
    }
  }

  return {
    token,
    user,
    bootstrapped,
    loading,
    isAuthenticated,
    bootstrap,
    register,
    login,
    logout,
    changePassword,
    clearSession,
  }
})
