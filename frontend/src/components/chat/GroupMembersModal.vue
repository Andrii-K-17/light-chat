<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { XMarkIcon, UserPlusIcon, TrashIcon } from '@heroicons/vue/24/outline'
import { useAuthStore } from '@/stores/useAuthStore'
import type { ChatMember } from '@/types'
import { getChatMembers, addChatMember, removeChatMember } from '@/api/chats'

const props = defineProps<{ chatId: number; createdBy: number | null }>()
const emit = defineEmits<{ close: [] }>()

const auth = useAuthStore()

const members = ref<ChatMember[]>([])
const loading = ref(true)
const addUsername = ref('')
const addError = ref('')
const adding = ref(false)

const isCreator = computed(() => auth.user?.id === props.createdBy)

onMounted(async () => {
  try {
    members.value = await getChatMembers(props.chatId)
  } finally {
    loading.value = false
  }
})

/**
 * Adds a new member by username and appends them to the list.
 */
async function handleAdd() {
  const username = addUsername.value.trim()
  if (!username) return

  adding.value = true
  addError.value = ''
  try {
    const member = await addChatMember(props.chatId, username)
    const exists = members.value.find((m) => m.id === member.id)
    if (!exists) members.value.push(member)
    addUsername.value = ''
  } catch (e: unknown) {
    addError.value = e instanceof Error ? e.message : 'Failed to add user'
  } finally {
    adding.value = false
  }
}

/**
 * Removes a member from the chat and updates the local list.
 */
async function handleRemove(memberID: number) {
  try {
    await removeChatMember(props.chatId, memberID)
    members.value = members.value.filter((m) => m.id !== memberID)
  } catch (e: unknown) {
    addError.value = e instanceof Error ? e.message : 'Failed to remove user'
  }
}
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 p-4 backdrop-blur-sm"
    @click.self="emit('close')"
  >
    <div
      class="flex w-full max-w-sm flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-2xl dark:border-slate-800 dark:bg-slate-900"
    >
      <div
        class="flex items-center justify-between border-b border-slate-200 px-5 py-4 dark:border-slate-800"
      >
        <h2 class="font-semibold text-slate-900 dark:text-slate-100">Group Members</h2>
        <button
          @click="emit('close')"
          class="flex size-8 items-center justify-center rounded-full text-slate-400 transition-all hover:cursor-pointer hover:bg-slate-100 hover:text-rose-500 active:scale-95 dark:hover:bg-slate-800 dark:hover:text-rose-400"
        >
          <XMarkIcon class="size-5" />
        </button>
      </div>

      <div v-if="isCreator" class="border-b border-slate-200 px-5 py-3 dark:border-slate-800">
        <div class="flex gap-2">
          <input
            v-model="addUsername"
            @keydown.enter="handleAdd"
            type="text"
            placeholder="Add by username..."
            class="min-w-0 flex-1 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/10 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:placeholder:text-slate-500"
          />
          <button
            @click="handleAdd"
            :disabled="adding || !addUsername.trim()"
            class="flex size-9 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-600 transition-all hover:cursor-pointer hover:bg-emerald-500/20 active:scale-95 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-emerald-400/10 dark:text-emerald-400 dark:hover:bg-emerald-400/20"
            title="Add member"
          >
            <UserPlusIcon class="size-4" />
          </button>
        </div>
        <p v-if="addError" class="mt-1.5 text-xs text-rose-500 dark:text-rose-400">
          {{ addError }}
        </p>
      </div>

      <div class="no-scrollbar max-h-72 overflow-y-auto">
        <div
          v-if="loading"
          class="flex items-center justify-center py-10 text-sm text-slate-400 dark:text-slate-500"
        >
          Loading...
        </div>

        <ul v-else class="divide-y divide-slate-100 dark:divide-slate-800">
          <li v-for="member in members" :key="member.id" class="flex items-center gap-3 px-5 py-3">
            <div
              class="flex size-9 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 text-sm font-semibold text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300"
            >
              {{ member.display_name.charAt(0).toUpperCase() }}
            </div>

            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium text-slate-900 dark:text-slate-100">
                {{ member.display_name }}
                <span
                  v-if="member.id === props.createdBy"
                  class="ml-1.5 rounded-full bg-emerald-100 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400"
                >
                  creator
                </span>
                <span
                  v-if="member.id === auth.user?.id"
                  class="ml-1.5 rounded-full bg-slate-100 px-1.5 py-0.5 text-[10px] font-semibold text-slate-500 dark:bg-slate-800 dark:text-slate-400"
                >
                  you
                </span>
              </p>
              <p class="truncate text-xs text-slate-400 dark:text-slate-500">
                @{{ member.username }}
              </p>
            </div>

            <button
              v-if="isCreator && member.id !== auth.user?.id && member.id !== props.createdBy"
              @click="handleRemove(member.id)"
              class="flex size-7 flex-shrink-0 items-center justify-center rounded-full text-slate-300 transition-all hover:cursor-pointer hover:bg-rose-50 hover:text-rose-500 active:scale-95 dark:text-slate-600 dark:hover:bg-rose-900/20 dark:hover:text-rose-400"
              title="Remove member"
            >
              <TrashIcon class="size-3.5" />
            </button>
          </li>

          <li
            v-if="members.length === 0 && !loading"
            class="px-5 py-8 text-center text-sm text-slate-400 dark:text-slate-500"
          >
            No members yet
          </li>
        </ul>
      </div>

      <div class="border-t border-slate-200 px-5 py-3 dark:border-slate-800">
        <p class="text-xs text-slate-400 dark:text-slate-500">
          {{ members.length }} member{{ members.length !== 1 ? 's' : '' }}
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  scrollbar-width: none;
}
</style>
