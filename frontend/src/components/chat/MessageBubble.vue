<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useAuthStore } from '@/stores/useAuthStore'
import { useChatStore } from '@/stores/useChatStore'
import type { Message } from '@/types'
import { CheckIcon, PencilIcon, TrashIcon, XMarkIcon } from '@heroicons/vue/24/outline'
import ConfirmModal from '@/components/ui/ConfirmModal.vue'

const props = defineProps<{ message: Message; isGroup: boolean }>()

const auth = useAuthStore()
const chatStore = useChatStore()

const isOwn = computed(() => props.message.user_id === auth.user?.id)

const showActions = ref(false)
const isEditing = ref(false)
const editValue = ref('')
const editRef = ref<HTMLTextAreaElement | null>(null)

const showDeleteModal = ref(false)
const isDeleting = ref(false)

/**
 * Formats a timestamp into a short HH:MM string.
 */
function formatTime(timestamp: string): string {
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

/**
 * Enters edit mode and focuses the textarea.
 */
async function startEdit() {
  editValue.value = props.message.content
  isEditing.value = true
  showActions.value = false

  await resizeEditTextarea()
  editRef.value?.focus()
}

/**
 * Saves the edited message if the content has changed.
 */
async function saveEdit() {
  const trimmed = editValue.value.trim()
  if (!trimmed || trimmed === props.message.content) {
    cancelEdit()
    return
  }
  await chatStore.editMessage(props.message.id, trimmed)
  isEditing.value = false
}

/**
 * Cancels edit mode without saving.
 */
function cancelEdit() {
  isEditing.value = false
  editValue.value = ''
}

/**
 * Handles keyboard shortcuts inside the edit textarea.
 */
function onEditKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    saveEdit()
    return
  }

  if (e.key === 'Escape') {
    cancelEdit()
  }
}

/**
 * Deletes the message after confirmation.
 */
async function confirmDelete() {
  isDeleting.value = true

  try {
    await chatStore.removeMessage(props.message.id)
    showDeleteModal.value = false
  } finally {
    isDeleting.value = false
  }
}

async function resizeEditTextarea() {
  await nextTick()

  const el = editRef.value
  if (!el) return

  el.style.height = 'auto'
  el.style.height = `${el.scrollHeight}px`
}
</script>

<template>
  <div
    :class="['group mb-2 flex w-full', isOwn ? 'justify-end' : 'justify-start']"
    @mouseenter="showActions = true"
    @mouseleave="showActions = false"
  >
    <!-- Action buttons -->
    <Transition name="actions">
      <div
        v-if="isOwn && showActions && !isEditing"
        class="mr-2 mb-4 flex items-center gap-1 self-end"
      >
        <button
          @click="startEdit"
          class="flex size-7 items-center justify-center rounded-full text-slate-400 transition-all hover:cursor-pointer hover:bg-slate-100 hover:text-emerald-600 active:scale-95 dark:hover:bg-slate-800 dark:hover:text-emerald-400"
          title="Edit message"
        >
          <PencilIcon class="size-4" />
        </button>
        <button
          @click="showDeleteModal = true"
          class="flex size-7 items-center justify-center rounded-full text-slate-400 transition-all hover:cursor-pointer hover:bg-rose-50 hover:text-rose-500 active:scale-95 dark:hover:bg-rose-900/20 dark:hover:text-rose-400"
          title="Delete message"
        >
          <TrashIcon class="size-4" />
        </button>
      </div>
    </Transition>

    <div
      :class="[
        'flex min-w-0 flex-col',
        isOwn ? 'items-end' : 'items-start',
        isEditing ? 'w-[75%] sm:w-[70%]' : 'max-w-[75%] sm:max-w-[70%]',
      ]"
    >
      <span
        v-if="!isOwn && isGroup"
        class="mb-1 px-1 text-xs font-medium text-emerald-600 dark:text-emerald-400"
      >
        {{ message.sender_display_name }}
      </span>

      <!-- Edit mode -->
      <div v-if="isEditing" class="flex w-full flex-col gap-1.5">
        <textarea
          ref="editRef"
          v-model="editValue"
          rows="1"
          maxlength="2000"
          @input="resizeEditTextarea"
          @keydown="onEditKeydown"
          class="max-h-48 w-full resize-none overflow-y-auto rounded-2xl border border-emerald-300 bg-emerald-50 px-3.5 py-2.5 text-[15px] leading-relaxed text-slate-900 outline-none transition-colors focus:ring-2 focus:ring-emerald-500/20 dark:border-emerald-700 dark:bg-slate-800 dark:text-slate-100"
        />
        <div class="flex justify-end gap-1.5">
          <button
            @click="cancelEdit"
            class="flex size-7 shrink-0 items-center justify-center rounded-full text-slate-400 transition-all hover:cursor-pointer hover:bg-slate-200 hover:text-slate-900 dark:hover:text-slate-100 active:scale-95 dark:hover:bg-slate-700"
            title="Cancel"
          >
            <XMarkIcon class="size-4" />
          </button>
          <button
            @click="saveEdit"
            class="flex size-7 shrink-0 items-center justify-center rounded-full bg-emerald-500 text-white transition-all hover:cursor-pointer hover:bg-emerald-600 active:scale-95"
            title="Save"
          >
            <CheckIcon class="size-4" />
          </button>
        </div>
      </div>

      <!-- Normal bubble -->
      <div
        v-else
        :class="[
          'min-w-0 max-w-full whitespace-pre-wrap break-words [overflow-wrap:anywhere]',
          'rounded-3xl px-3.5 py-2.5 text-[15px] leading-relaxed',
          isOwn
            ? 'rounded-br-md border border-slate-200/70 bg-emerald-200 shadow-sm shadow-emerald-600/10 dark:border-white/10 dark:bg-emerald-800 dark:shadow-emerald-400/20'
            : 'rounded-bl-md border border-slate-200/70 bg-white text-slate-800 shadow-sm shadow-emerald-600/10 dark:border-white/5 dark:bg-slate-900 dark:text-slate-100 dark:shadow-emerald-200/10',
        ]"
      >
        {{ message.content }}
      </div>

      <div v-if="!isEditing" class="mt-1 flex items-center gap-1 px-1">
        <span class="text-xs text-slate-400 dark:text-slate-500">
          {{ formatTime(message.created_at) }}
        </span>
        <CheckIcon
          v-if="isOwn"
          :class="[
            'size-3',
            message.is_read ? 'text-emerald-500' : 'text-slate-400 dark:text-slate-500',
          ]"
        />
      </div>
    </div>
  </div>

  <Transition name="modal">
    <ConfirmModal
      v-if="showDeleteModal"
      title="Delete message?"
      message="Are you sure you want to delete this message? This action cannot be undone."
      confirm-text="Delete"
      danger
      :loading="isDeleting"
      @confirm="confirmDelete"
      @cancel="showDeleteModal = false"
    />
  </Transition>
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
