import type { User } from '@/types'
import { fetchWithRefresh } from '@/api/fetchWithRefresh'

const BASE = '/api/auth'

/**
 * Sends a generic HTTP request and handles JSON response parsing.
 */
async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetchWithRefresh(url, {
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    ...options,
  })
  const data = await response.json()
  if (!response.ok) throw new Error(data.error ?? 'Request failed')
  return data as T
}

export const register = (
  email: string,
  username: string,
  display_name: string,
  password: string,
): Promise<User> =>
  request<User>(`${BASE}/register`, {
    method: 'POST',
    body: JSON.stringify({ email, username, display_name, password }),
  })

export const login = (email: string, password: string): Promise<User> =>
  request<User>(`${BASE}/login`, {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })

export const logout = (): Promise<void> => request<void>(`${BASE}/logout`, { method: 'POST' })

export const fetchMe = (): Promise<User> => request<User>(`${BASE}/me`)

/**
 * Attempts to refresh the access token using the refresh token cookie.
 */
export const refreshToken = (): Promise<void> =>
  request<void>(`${BASE}/refresh`, { method: 'POST' })

export const updateProfile = (patch: {
  display_name?: string
  username?: string
  email?: string
  status?: string
}): Promise<User> =>
  request<User>(`${BASE}/me`, {
    method: 'PATCH',
    body: JSON.stringify(patch),
  })
