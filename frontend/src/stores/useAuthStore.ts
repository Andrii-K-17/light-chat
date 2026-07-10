import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { User } from '@/types'
import * as authApi from '@/api/auth'

/**
 * Global store for user authentication and session management.
 */
export const useAuthStore = defineStore('auth', () => {
  /** Current authenticated user state. */
  const user = ref<User | null>(null)

  const isLoggedIn = computed(() => user.value !== null)

  /** Cache for the initialization request to prevent parallel duplicate calls. */
  let initPromise: Promise<void> | null = null

  /**
   * Initializes the session by fetching the current user profile.
   */
  async function init(): Promise<void> {
    if (initPromise) return initPromise
    initPromise = (async () => {
      try {
        user.value = await authApi.fetchMe()
      } catch {
        user.value = null
      } finally {
        initPromise = null
      }
    })()
    return initPromise
  }

  async function register(
    email: string,
    username: string,
    display_name: string,
    password: string,
  ): Promise<void> {
    user.value = await authApi.register(email, username, display_name, password)
  }

  async function login(email: string, password: string): Promise<void> {
    user.value = await authApi.login(email, password)
  }

  async function logout(): Promise<void> {
    await authApi.logout()
    user.value = null
    initPromise = null
  }

  async function updateProfile(patch: {
    display_name?: string
    username?: string
    email?: string
    status?: string
  }): Promise<void> {
    user.value = await authApi.updateProfile(patch)
  }

  return {
    user,
    isLoggedIn,
    init,
    register,
    login,
    logout,
    updateProfile,
  }
})
