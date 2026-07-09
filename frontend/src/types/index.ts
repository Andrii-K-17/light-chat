export interface User {
  id: number
  email: string
  username: string
  display_name: string
  status: string
  created_at: string
}

export interface ChatMember {
  id: number
  username: string
  display_name: string
  status: string
}

export interface Message {
  id: number
  chat_id: number
  user_id: number
  content: string
  is_read: boolean
  created_at: string
  sender_username: string
  sender_display_name: string
}

export interface Chat {
  id: number
  name: string | null
  is_group: boolean
  created_by: number | null
  created_at: string
  members: ChatMember[]
  last_message: Message | null
  unread_count: number
}

export interface WsEvent {
  type: 'new_message' | 'read_receipt'
  payload: unknown
}

export interface WsReadReceipt {
  chat_id: number
  reader_id: number
}
