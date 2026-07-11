<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/useAuthStore'
import { useChatStore } from '@/stores/useChatStore'
import type { Chat } from '@/types'

const props = defineProps<{ chat: Chat }>()

const auth = useAuthStore()
const chatStore = useChatStore()

const isActive = computed(() => chatStore.activeChatId === props.chat.id)

/**
 * Returns the display name for a chat.
 */
const chatName = computed(() => {
  if (props.chat.is_group) return props.chat.name ?? 'Group chat'
  const other = props.chat.members.find((m) => m.id !== auth.user?.id)
  return other?.display_name ?? 'Unknown'
})

/**
 * Returns the username label shown below the chat name for direct chats.
 */
const chatUsername = computed(() => {
  if (props.chat.is_group) return null
  const other = props.chat.members.find((m) => m.id !== auth.user?.id)
  return other ? `@${other.username}` : null
})

/**
 * Formats a timestamp into a short time or date string.
 */
function formatTime(iso: string): string {
  const date = new Date(iso)
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  if (isToday) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
}
</script>

<template>
  <li
    @click="chatStore.openChat(chat.id)"
    :class="[
      'flex items-center gap-3 px-3 py-2.5 rounded-xl cursor-pointer transition-colors',
      isActive
        ? 'bg-emerald-100 dark:bg-emerald-900/40'
        : 'hover:bg-emerald-50 dark:hover:bg-slate-800/60',
    ]"
  >
    <div
      class="w-10 h-10 rounded-full bg-teal-200 dark:bg-emerald-800 flex items-center justify-center flex-shrink-0 text-emerald-800 dark:text-emerald-200 font-semibold text-sm select-none"
    >
      {{ chatName.charAt(0).toUpperCase() }}
    </div>

    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between gap-1">
        <span class="text-sm font-medium truncate text-slate-900 dark:text-slate-100">
          {{ chatName }}
        </span>
        <span
          v-if="chat.last_message"
          class="text-xs text-gray-500 dark:text-slate-500 flex-shrink-0"
        >
          {{ formatTime(chat.last_message.created_at) }}
        </span>
      </div>

      <div class="flex items-center justify-between gap-1 mt-0.5">
        <span class="text-xs text-gray-500 dark:text-slate-400 truncate">
          {{ chat.last_message?.content ?? chatUsername ?? '&nbsp;' }}
        </span>
        <span
          v-if="chat.unread_count > 0"
          class="flex-shrink-0 bg-emerald-500 text-white text-xs font-semibold rounded-full px-1.5 py-0.5 min-w-[1.25rem] text-center"
        >
          {{ chat.unread_count }}
        </span>
      </div>
    </div>
  </li>
</template>
