import { refreshToken } from '@/api/auth'

type PendingRequest = {
  resolve: () => void
  reject: (error?: unknown) => void
}

let isRefreshing = false
let pendingQueue: PendingRequest[] = []

/** 
 * Drains the pending queue after a refresh attempt.
 */
function drainQueue(error?: unknown) {
  pendingQueue.forEach(({ resolve, reject }) => {
    if (error !== undefined) reject(error)
    else resolve()
  })
  pendingQueue = []
}

/**
 * Wraps fetch with automatic access token refresh on 401 responses.
 */
export async function fetchWithRefresh(input: RequestInfo, init?: RequestInit): Promise<Response> {
  const fetchOptions: RequestInit = { credentials: 'include', ...init }
  const response = await fetch(input, fetchOptions)

  if (response.status !== 401) return response

  const url = typeof input === 'string' ? input : input.url
  if (url.includes('/api/auth/refresh')) {
    throw new Error('Refresh token expired or invalid')
  }

  if (isRefreshing) {
    await new Promise<void>((resolve, reject) => {
      pendingQueue.push({ resolve, reject })
    })
    return fetch(input, fetchOptions)
  }

  isRefreshing = true
  try {
    await refreshToken()
    drainQueue()
    return fetch(input, fetchOptions)
  } catch (error) {
    drainQueue(error)
    throw error
  } finally {
    isRefreshing = false
  }
}
