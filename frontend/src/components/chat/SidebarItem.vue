<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/useAuthStore'
import { useChatStore } from '@/stores/useChatStore'
import type { Chat } from '@/types'
import { TrashIcon } from '@heroicons/vue/24/outline'
import ConfirmModal from '@/components/ui/ConfirmModal.vue'

const props = defineProps<{ chat: Chat; compact?: boolean }>()

const auth = useAuthStore()
const chatStore = useChatStore()

const showActions = ref(false)
const deleting = ref(false)
const showDeleteModal = ref(false)

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
  if (isToday) return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

/**
 * Opens the delete confirmation modal.
 */
function openDeleteModal(e: MouseEvent) {
  e.stopPropagation()
  showDeleteModal.value = true
}

/**
 * Deletes the chat after confirmation.
 */
async function confirmDelete() {
  deleting.value = true

  try {
    await chatStore.deleteChat(props.chat.id)
    showDeleteModal.value = false
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <!-- Compact mode -->
  <button
    v-if="compact"
    @click="chatStore.openChat(chat.id)"
    :class="[
      'relative m-1 flex size-10 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold ring-1 transition-all hover:cursor-pointer active:scale-95',
      isActive
        ? 'bg-emerald-200/90 text-emerald-700 ring-emerald-500/30 dark:bg-emerald-400/20 dark:text-emerald-300 dark:ring-emerald-400/20'
        : 'bg-emerald-100/70 text-slate-600 ring-slate-900/5 hover:bg-emerald-100 hover:text-emerald-600 dark:bg-slate-800 dark:text-slate-300 dark:ring-white/5 dark:hover:bg-emerald-300/10 dark:hover:text-emerald-400',
    ]"
    :title="chatName"
  >
    {{ chatName.charAt(0).toUpperCase() }}
    <span
      v-if="chat.unread_count > 0"
      class="absolute -right-0.5 -top-0.5 flex size-4 items-center justify-center rounded-full bg-emerald-500 text-[10px] font-bold text-white"
    >
      {{ chat.unread_count > 9 ? '9+' : chat.unread_count }}
    </span>
  </button>

  <!-- Full mode -->
  <li
    v-else
    @click="chatStore.openChat(chat.id)"
    @mouseenter="showActions = true"
    @mouseleave="showActions = false"
    :class="[
      'group relative flex cursor-pointer items-center gap-3 rounded-3xl px-3 py-2.5 transition-colors',
      isActive
        ? 'bg-emerald-500/10 dark:bg-emerald-400/10'
        : 'hover:bg-slate-100/80 dark:hover:bg-slate-800/60',
    ]"
  >
    <div
      :class="[
        'flex size-10 flex-shrink-0 select-none items-center justify-center rounded-full text-sm font-semibold',
        isActive
          ? 'bg-emerald-200/90 text-emerald-700 dark:bg-emerald-400/20 dark:text-emerald-300'
          : 'bg-emerald-100/70 text-slate-600 dark:bg-slate-800 dark:text-slate-300',
      ]"
    >
      {{ chatName.charAt(0).toUpperCase() }}
    </div>

    <div class="min-w-0 flex-1">
      <div class="flex items-center justify-between gap-1">
        <span class="truncate text-sm font-medium text-slate-900 dark:text-slate-100">
          {{ chatName }}
        </span>
        <span
          v-if="chat.last_message && !showActions"
          class="flex-shrink-0 text-xs text-slate-400 dark:text-slate-500"
        >
          {{ formatTime(chat.last_message.created_at) }}
        </span>

        <!-- Delete button -->
        <Transition name="actions">
          <button
            v-if="showActions"
            @click="openDeleteModal"
            :disabled="deleting"
            class="flex-shrink-0 rounded-full text-slate-400 transition-all hover:cursor-pointer hover:bg-rose-50 hover:text-rose-500 active:scale-95 disabled:opacity-50 dark:hover:bg-rose-900/20 dark:hover:text-rose-400"
            title="Delete chat"
          >
            <TrashIcon class="size-4" />
          </button>
        </Transition>
      </div>

      <div class="mt-0.5 flex items-center justify-between gap-1">
        <span class="truncate text-xs text-slate-400 dark:text-slate-500">
          {{ chat.last_message?.content ?? chatUsername ?? '' }}
        </span>
        <span
          v-if="chat.unread_count > 0"
          class="flex-shrink-0 rounded-full bg-emerald-500 px-1.5 py-0.5 text-center text-[10px] font-bold text-white"
        >
          {{ chat.unread_count > 9 ? '9+' : chat.unread_count }}
        </span>
      </div>
    </div>
  </li>

  <ConfirmModal
    v-if="showDeleteModal"
    title="Delete chat?"
    :message="`Are you sure you want to delete this chat (${chatName})? This action cannot be undone.`"
    confirm-text="Delete"
    danger
    :loading="deleting"
    @confirm="confirmDelete"
    @cancel="showDeleteModal = false"
  />
</template>

<style scoped>
.actions-enter-active,
.actions-leave-active {
  transition:
    opacity 150ms ease,
    transform 150ms ease;
}
.actions-enter-from,
.actions-leave-to {
  opacity: 0;
  transform: scale(0.8);
}
</style>
