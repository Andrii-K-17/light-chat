import type { Chat, ChatMember, Message } from '@/types'
import { fetchWithRefresh } from '@/api/fetchWithRefresh'

/**
 * Sends a generic authenticated request and handles JSON response parsing.
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

/**
 * Fetches all chats the current user is a member of.
 */
export const getChats = (): Promise<Chat[]> => request<Chat[]>('/api/chats')

/**
 * Creates a new direct or group chat.
 */
export const createChat = (payload: {
  name?: string | null
  is_group: boolean
  member_ids: number[]
}): Promise<Chat> =>
  request<Chat>('/api/chats', {
    method: 'POST',
    body: JSON.stringify(payload),
  })

/**
 * Fetches paginated message history for a chat.
 */
export const getMessages = (chatID: number, limit = 50, offset = 0): Promise<Message[]> =>
  request<Message[]>(`/api/chats/${chatID}/messages?limit=${limit}&offset=${offset}`)

/**
 * Searches messages in a chat by content substring.
 */
export const searchMessages = (chatID: number, q: string, limit = 50): Promise<Message[]> =>
  request<Message[]>(
    `/api/chats/${chatID}/messages/search?q=${encodeURIComponent(q)}&limit=${limit}`,
  )

/**
 * Deletes a chat by ID.
 */
export const deleteChat = (chatID: number): Promise<{ deleted: boolean }> =>
  request<{ deleted: boolean }>(`/api/chats/${chatID}`, { method: 'DELETE' })

/**
 * Edits the content of a message.
 */
export const updateMessage = (
  chatID: number,
  messageID: number,
  content: string,
): Promise<Message> =>
  request<Message>(`/api/chats/${chatID}/messages/${messageID}`, {
    method: 'PATCH',
    body: JSON.stringify({ content }),
  })

/**
 * Deletes a message by ID.
 */
export const deleteMessage = (chatID: number, messageID: number): Promise<{ deleted: boolean }> =>
  request<{ deleted: boolean }>(`/api/chats/${chatID}/messages/${messageID}`, {
    method: 'DELETE',
  })

/**
 * Fetches all members of a chat.
 */
export const getChatMembers = (chatID: number): Promise<ChatMember[]> =>
  request<ChatMember[]>(`/api/chats/${chatID}/members`)

/**
 * Adds a user to a group chat by username.
 */
export const addChatMember = (chatID: number, username: string): Promise<ChatMember> =>
  request<ChatMember>(`/api/chats/${chatID}/members`, {
    method: 'POST',
    body: JSON.stringify({ username }),
  })

/**
 * Removes a user from a group chat by their ID.
 */
export const removeChatMember = (chatID: number, memberID: number): Promise<{ removed: boolean }> =>
  request<{ removed: boolean }>(`/api/chats/${chatID}/members/${memberID}`, {
    method: 'DELETE',
  })
