<script setup lang="ts">
import { ref } from 'vue'
import { XMarkIcon, UserPlusIcon } from '@heroicons/vue/24/outline'
import { useChatStore } from '@/stores/useChatStore'
import { searchUser } from '@/api/users'
import type { User } from '@/types'

const emit = defineEmits<{ close: [] }>()

const chatStore = useChatStore()

const mode = ref<'direct' | 'group'>('direct')
const searchQuery = ref('')
const groupName = ref('')
const foundUser = ref<User | null>(null)
const selectedUsers = ref<User[]>([])
const searchError = ref('')
const saving = ref(false)
const error = ref('')

/**
 * Searches for a user by exact username and sets the result.
 */
async function search() {
  searchError.value = ''
  foundUser.value = null
  if (!searchQuery.value.trim()) return
  try {
    foundUser.value = await searchUser(searchQuery.value.trim())
  } catch {
    searchError.value = 'User not found'
  }
}

/**
 * Adds the found user to the selected list if not already present.
 */
function addUser() {
  if (!foundUser.value) return
  const exists = selectedUsers.value.find((u) => u.id === foundUser.value!.id)
  if (!exists) selectedUsers.value.push(foundUser.value)
  foundUser.value = null
  searchQuery.value = ''
}

/**
 * Removes a user from the selected list by ID.
 */
function removeUser(id: number) {
  selectedUsers.value = selectedUsers.value.filter((u) => u.id !== id)
}

/**
 * Creates the chat with all selected users and closes the modal.
 */
async function save() {
  if (selectedUsers.value.length === 0) {
    error.value = 'Add at least one user'
    return
  }
  saving.value = true
  error.value = ''
  try {
    const chat = await chatStore.createChat({
      name: mode.value === 'group' ? groupName.value.trim() || null : null,
      is_group: mode.value === 'group',
      member_ids: selectedUsers.value.map((u) => u.id),
    })
    await chatStore.openChat(chat.id)
    emit('close')
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to create the chat'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div
    class="fixed inset-0 bg-slate-800/30 backdrop-blur-sm z-50 flex items-center justify-center p-4 dark:bg-black/50"
    @click.self="emit('close')"
  >
    <div
      class="bg-white border border-emerald-200 rounded-2xl w-full max-w-sm shadow-2xl dark:bg-slate-900 dark:border-slate-800 transition-colors overflow-hidden"
    >
      <div
        class="flex items-center justify-between px-5 py-4 border-b border-emerald-200 dark:border-slate-800"
      >
        <h2 class="font-semibold text-slate-900 dark:text-slate-100">New Chat</h2>
        <button
          @click="emit('close')"
          class="text-gray-500 hover:text-rose-600 hover:cursor-pointer transition-colors dark:text-slate-400 dark:hover:text-rose-400"
        >
          <XMarkIcon class="w-5 h-5" />
        </button>
      </div>

      <div class="px-5 py-4 space-y-4">
        <div
          class="flex bg-emerald-50 rounded-xl p-1 border border-emerald-200 dark:bg-slate-800 dark:border-slate-700"
        >
          <button
            v-for="m in ['direct', 'group']"
            :key="m"
            @click="mode = m"
            :class="[
              'flex-1 py-1 rounded-lg text-xs font-medium hover:cursor-pointer transition-all capitalize',
              mode === m ? 'bg-emerald-600 text-white' : 'text-gray-600 dark:text-slate-400',
            ]"
          >
            {{ m === 'direct' ? 'Direct' : 'Group' }}
          </button>
        </div>

        <div v-if="mode === 'group'">
          <input
            v-model="groupName"
            type="text"
            placeholder="Group name (optional)"
            class="w-full bg-emerald-50/30 border border-emerald-200 rounded-xl px-3 py-2 text-sm text-slate-900 placeholder-gray-400 focus:outline-none focus:border-emerald-400 dark:bg-slate-800 dark:border-slate-700 dark:text-slate-100 dark:placeholder-slate-500"
          />
        </div>

        <div class="flex gap-2">
          <input
            v-model="searchQuery"
            @keydown.enter="search"
            type="text"
            placeholder="Search by username..."
            class="flex-1 bg-emerald-50/30 border border-emerald-200 rounded-xl px-3 py-2 text-sm text-slate-900 placeholder-gray-400 focus:outline-none focus:border-emerald-400 dark:bg-slate-800 dark:border-slate-700 dark:text-slate-100 dark:placeholder-slate-500"
          />
          <button
            @click="search"
            class="px-3 py-2 bg-emerald-100 hover:bg-emerald-200 text-emerald-700 rounded-xl text-sm hover:cursor-pointer transition-colors dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-emerald-400"
          >
            Find
          </button>
        </div>

        <p v-if="searchError" class="text-xs text-rose-500 dark:text-rose-400">
          {{ searchError }}
        </p>

        <div
          v-if="foundUser"
          class="flex items-center justify-between px-3 py-2 rounded-xl border border-emerald-200 dark:border-slate-700"
        >
          <div>
            <p class="text-sm font-medium text-slate-900 dark:text-slate-100">
              {{ foundUser.display_name }}
            </p>
            <p class="text-xs text-gray-500 dark:text-slate-400">@{{ foundUser.username }}</p>
          </div>
          <button
            @click="addUser"
            class="p-1.5 bg-emerald-100 hover:bg-emerald-200 text-emerald-700 rounded-lg hover:cursor-pointer transition-colors dark:bg-slate-700 dark:hover:bg-slate-600 dark:text-emerald-400"
          >
            <UserPlusIcon class="w-4 h-4" />
          </button>
        </div>

        <ul v-if="selectedUsers.length > 0" class="space-y-1.5">
          <li
            v-for="u in selectedUsers"
            :key="u.id"
            class="flex items-center justify-between px-3 py-1.5 rounded-xl bg-emerald-50 dark:bg-slate-800"
          >
            <span class="text-sm text-slate-900 dark:text-slate-100">{{ u.display_name }}</span>
            <button
              @click="removeUser(u.id)"
              class="text-gray-400 hover:text-rose-500 hover:cursor-pointer transition-colors"
            >
              <XMarkIcon class="w-4 h-4" />
            </button>
          </li>
        </ul>

        <p v-if="error" class="text-xs text-rose-500 dark:text-rose-400">{{ error }}</p>

        <button
          @click="save"
          :disabled="saving"
          class="w-full py-2 bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed hover:cursor-pointer text-white text-sm font-medium rounded-xl transition-colors dark:bg-emerald-600 dark:hover:bg-emerald-500"
        >
          {{ saving ? 'Creating…' : 'Start Chat' }}
        </button>
      </div>
    </div>
  </div>
</template>
