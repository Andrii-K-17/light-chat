import type { User } from '@/types'
import { fetchWithRefresh } from '@/api/fetchWithRefresh'

export const searchUser = async (username: string): Promise<User> => {
  const response = await fetchWithRefresh(
    `/api/users/search?username=${encodeURIComponent(username)}`,
    {
      credentials: 'include',
    },
  )
  const data = await response.json()
  if (!response.ok) throw new Error(data.error ?? 'User not found')
  return data as User
}
