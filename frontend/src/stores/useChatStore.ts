import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { Chat, Message, WsEvent, WsReadReceipt, WsMessageDeleted } from '@/types'
import * as chatsApi from '@/api/chats'
import { useAuthStore } from '@/stores/useAuthStore'

/**
 * Global store for chats, messages, and WebSocket connection management.
 */
export const useChatStore = defineStore('chat', () => {
  const authStore = useAuthStore()

  /** Full list of chats the user belongs to. */
  const chats = ref<Chat[]>([])

  /** ID of the currently open chat. */
  const activeChatId = ref<number | null>(null)

  /** Messages for the currently open chat. */
  const messages = ref<Message[]>([])

  /** Whether older messages are being loaded. */
  const loadingMessages = ref(false)

  /** Search results for in-chat message search. */
  const searchResults = ref<Message[]>([])

  /** Whether a message search is in progress. */
  const searchLoading = ref(false)

  /** Current WebSocket instance. */
  let socket: WebSocket | null = null

  /** Reconnect timer handle. */
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null

  /** Currently active chat object. */
  const activeChat = computed(() => chats.value.find((c) => c.id === activeChatId.value) ?? null)

  /**
   * Loads all chats for the current user.
   */
  async function loadChats(): Promise<void> {
    chats.value = await chatsApi.getChats()
  }

  /**
   * Opens a chat, loads its messages, and connects the WebSocket.
   */
  async function openChat(chatId: number): Promise<void> {
    if (activeChatId.value === chatId) return

    disconnectWs()
    activeChatId.value = chatId
    messages.value = []
    loadingMessages.value = true

    try {
      const fetched = await chatsApi.getMessages(chatId)
      messages.value = [...fetched].reverse()
    } finally {
      loadingMessages.value = false
    }

    connectWs(chatId)

    const chat = chats.value.find((c) => c.id === chatId)
    if (chat) chat.unread_count = 0
  }

  /**
   * Loads older messages and prepends them to the message list.
   */
  async function loadMoreMessages(): Promise<boolean> {
    if (!activeChatId.value || loadingMessages.value) return false
    loadingMessages.value = true
    try {
      const older = await chatsApi.getMessages(activeChatId.value, 50, messages.value.length)
      if (older.length === 0) return false
      messages.value = [...older.reverse(), ...messages.value]
      return true
    } finally {
      loadingMessages.value = false
    }
  }

  /**
   * Creates a new chat and prepends it to the chat list.
   */
  async function createChat(payload: {
    name?: string | null
    is_group: boolean
    member_ids: number[]
  }): Promise<Chat> {
    const chat = await chatsApi.createChat(payload)
    const exists = chats.value.find((c) => c.id === chat.id)
    if (!exists) chats.value.unshift(chat)
    return chat
  }

  /**
   * Deletes a chat and removes it from the store.
   */
  async function deleteChat(chatId: number): Promise<void> {
    await chatsApi.deleteChat(chatId)
    chats.value = chats.value.filter((c) => c.id !== chatId)
    if (activeChatId.value === chatId) {
      disconnectWs()
      activeChatId.value = null
      messages.value = []
    }
  }

  /**
   * Edits a message via the API and updates it in the local list.
   */
  async function editMessage(messageId: number, content: string): Promise<void> {
    if (!activeChatId.value) return
    const updated = await chatsApi.updateMessage(activeChatId.value, messageId, content)
    messages.value = messages.value.map((m) => (m.id === messageId ? { ...m, ...updated } : m))
  }

  /**
   * Deletes a message via the API and removes it from the local list.
   */
  async function removeMessage(messageId: number): Promise<void> {
    if (!activeChatId.value) return
    await chatsApi.deleteMessage(activeChatId.value, messageId)
    messages.value = messages.value.filter((m) => m.id !== messageId)
  }

  /**
   * Searches messages in the active chat by query string.
   */
  async function searchMessages(query: string): Promise<void> {
    if (!activeChatId.value) return
    if (!query.trim()) {
      searchResults.value = []
      return
    }
    searchLoading.value = true
    try {
      searchResults.value = await chatsApi.searchMessages(activeChatId.value, query)
    } catch {
      searchResults.value = []
    } finally {
      searchLoading.value = false
    }
  }

  /**
   * Connects a WebSocket for the given chat and sets up event handling.
   */
  function connectWs(chatId: number): void {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${protocol}://${location.host}/ws?chat_id=${chatId}`

    socket = new WebSocket(url)

    socket.onmessage = (event) => {
      try {
        const wsEvent: WsEvent = JSON.parse(event.data as string)
        handleWsEvent(wsEvent)
      } catch (error) {
        console.warn(error)
      }
    }

    socket.onclose = () => {
      if (activeChatId.value === chatId) {
        reconnectTimer = setTimeout(() => connectWs(chatId), 3000)
      }
    }

    socket.onerror = () => socket?.close()
  }

  /**
   * Disconnects the active WebSocket connection.
   */
  function disconnectWs(): void {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (socket) {
      socket.onclose = null
      socket.close()
      socket = null
    }
  }

  /**
   * Routes incoming WebSocket events to their respective handlers.
   */
  function handleWsEvent(event: WsEvent): void {
    if (event.type === 'new_message') {
      const msg = event.payload as Message
      if (msg.chat_id === activeChatId.value) {
        messages.value.push(msg)
      }
      const chat = chats.value.find((c) => c.id === msg.chat_id)
      if (chat) {
        chat.last_message = msg
        if (msg.chat_id !== activeChatId.value && msg.user_id !== authStore.user?.id) {
          chat.unread_count++
        }
        chats.value = [chat, ...chats.value.filter((c) => c.id !== chat.id)]
      }
    }

    if (event.type === 'message_updated') {
      const updated = event.payload as Message
      if (updated.chat_id === activeChatId.value) {
        messages.value = messages.value.map((m) => (m.id === updated.id ? { ...m, ...updated } : m))
      }
    }

    if (event.type === 'message_deleted') {
      const { message_id, chat_id } = event.payload as WsMessageDeleted
      if (chat_id === activeChatId.value) {
        messages.value = messages.value.filter((m) => m.id !== message_id)
      }
    }

    if (event.type === 'read_receipt') {
      const receipt = event.payload as WsReadReceipt
      if (receipt.chat_id === activeChatId.value) {
        messages.value = messages.value.map((m) => ({ ...m, is_read: true }))
      }
    }
  }

  /**
   * Sends a message through the active WebSocket connection.
   */
  function sendMessage(content: string): void {
    if (!socket || socket.readyState !== WebSocket.OPEN || !activeChatId.value) return
    socket.send(
      JSON.stringify({
        type: 'send_message',
        payload: { chat_id: activeChatId.value, content },
      }),
    )
  }

  /**
   * Sends a read receipt through the active WebSocket connection.
   */
  function sendReadReceipt(): void {
    if (!socket || socket.readyState !== WebSocket.OPEN) return
    socket.send(JSON.stringify({ type: 'read_receipt', payload: {} }))
  }

  /**
   * Resets the store state to its default values.
   */
  function reset(): void {
    disconnectWs()
    chats.value = []
    activeChatId.value = null
    messages.value = []
    searchResults.value = []
  }

  return {
    chats,
    activeChatId,
    activeChat,
    messages,
    loadingMessages,
    searchResults,
    searchLoading,
    loadChats,
    openChat,
    loadMoreMessages,
    createChat,
    deleteChat,
    editMessage,
    removeMessage,
    searchMessages,
    sendMessage,
    sendReadReceipt,
    disconnectWs,
    reset,
  }
})
