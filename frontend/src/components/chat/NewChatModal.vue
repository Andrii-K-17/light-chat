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
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 p-4 backdrop-blur-sm"
    @click.self="emit('close')"
  >
    <div
      class="flex w-full max-w-sm flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-2xl dark:border-slate-800 dark:bg-slate-900"
    >
      <div
        class="flex items-center justify-between border-b border-slate-200 px-5 py-4 dark:border-slate-800"
      >
        <h2 class="font-semibold text-slate-900 dark:text-slate-100">New Chat</h2>

        <button
          @click="emit('close')"
          class="flex size-8 items-center justify-center rounded-full text-slate-400 transition-all hover:cursor-pointer hover:bg-slate-100 hover:text-rose-500 active:scale-95 dark:hover:bg-slate-800 dark:hover:text-rose-400"
          aria-label="Close"
        >
          <XMarkIcon class="size-5" />
        </button>
      </div>

      <div class="space-y-4 px-5 py-4">
        <div
          class="flex rounded-xl border border-slate-200 bg-slate-100/70 p-1 dark:border-slate-700 dark:bg-slate-800"
        >
          <button
            @click="mode = 'direct'"
            :class="[
              'flex-1 rounded-lg py-1.5 text-xs font-medium transition-all hover:cursor-pointer',
              mode === 'direct'
                ? 'bg-white text-slate-900 shadow-sm dark:bg-slate-700 dark:text-slate-100'
                : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200',
            ]"
          >
            Direct
          </button>

          <button
            @click="mode = 'group'"
            :class="[
              'flex-1 rounded-lg py-1.5 text-xs font-medium transition-all hover:cursor-pointer',
              mode === 'group'
                ? 'bg-white text-slate-900 shadow-sm dark:bg-slate-700 dark:text-slate-100'
                : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200',
            ]"
          >
            Group
          </button>
        </div>

        <div v-if="mode === 'group'">
          <input
            v-model="groupName"
            type="text"
            placeholder="Group name (optional)"
            class="w-full rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/10 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:placeholder:text-slate-500"
          />
        </div>

        <div>
          <div class="flex gap-2">
            <input
              v-model="searchQuery"
              @keydown.enter="search"
              type="text"
              placeholder="Search by username..."
              class="min-w-0 flex-1 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/10 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:placeholder:text-slate-500"
            />

            <button
              @click="search"
              :disabled="!searchQuery.trim()"
              class="rounded-xl bg-emerald-500/10 px-3 text-sm font-medium text-emerald-600 transition-all hover:cursor-pointer hover:bg-emerald-500/20 active:scale-95 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-emerald-400/10 dark:text-emerald-400 dark:hover:bg-emerald-400/20"
            >
              Find
            </button>
          </div>

          <p v-if="searchError" class="mt-1.5 text-xs text-rose-500 dark:text-rose-400">
            {{ searchError }}
          </p>
        </div>

        <div
          v-if="foundUser"
          class="flex items-center gap-3 rounded-xl border border-slate-200 px-3 py-2.5 dark:border-slate-700"
        >
          <div
            class="flex size-9 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 text-sm font-semibold text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300"
          >
            {{ foundUser.display_name.charAt(0).toUpperCase() }}
          </div>

          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium text-slate-900 dark:text-slate-100">
              {{ foundUser.display_name }}
            </p>
            <p class="truncate text-xs text-slate-400 dark:text-slate-500">
              @{{ foundUser.username }}
            </p>
          </div>

          <button
            @click="addUser"
            class="flex size-8 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-600 transition-all hover:cursor-pointer hover:bg-emerald-500/20 active:scale-95 dark:bg-emerald-400/10 dark:text-emerald-400 dark:hover:bg-emerald-400/20"
            title="Add user"
          >
            <UserPlusIcon class="size-4" />
          </button>
        </div>

        <ul
          v-if="selectedUsers.length > 0"
          class="overflow-hidden rounded-xl border border-slate-200 dark:border-slate-700"
        >
          <li
            v-for="(user, index) in selectedUsers"
            :key="user.id"
            :class="[
              'flex items-center gap-3 px-3 py-2.5',
              index !== selectedUsers.length - 1
                ? 'border-b border-slate-100 dark:border-slate-800'
                : '',
            ]"
          >
            <div
              class="flex size-8 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 text-xs font-semibold text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300"
            >
              {{ user.display_name.charAt(0).toUpperCase() }}
            </div>

            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium text-slate-900 dark:text-slate-100">
                {{ user.display_name }}
              </p>
              <p class="truncate text-xs text-slate-400 dark:text-slate-500">
                @{{ user.username }}
              </p>
            </div>

            <button
              @click="removeUser(user.id)"
              class="flex size-7 flex-shrink-0 items-center justify-center rounded-full text-slate-300 transition-all hover:cursor-pointer hover:bg-rose-50 hover:text-rose-500 active:scale-95 dark:text-slate-600 dark:hover:bg-rose-900/20 dark:hover:text-rose-400"
              title="Remove user"
            >
              <XMarkIcon class="size-4" />
            </button>
          </li>
        </ul>

        <p v-if="error" class="text-xs text-rose-500 dark:text-rose-400">
          {{ error }}
        </p>
      </div>

      <div class="border-t border-slate-200 px-5 py-4 dark:border-slate-800">
        <button
          @click="save"
          :disabled="saving || selectedUsers.length === 0"
          class="w-full rounded-xl bg-emerald-600 py-2.5 text-sm font-medium text-white transition-all hover:cursor-pointer hover:bg-emerald-700 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-emerald-500"
        >
          {{ saving ? 'Creating…' : mode === 'group' ? 'Create Group' : 'Start Chat' }}
        </button>
      </div>
    </div>
  </div>
</template>
