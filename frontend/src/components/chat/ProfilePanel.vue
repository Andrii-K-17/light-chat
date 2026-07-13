<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  PencilIcon,
  CheckIcon,
  XMarkIcon,
  ArrowRightStartOnRectangleIcon,
  ChevronLeftIcon,
} from '@heroicons/vue/24/outline'
import { useAuthStore } from '@/stores/useAuthStore'
import { useChatStore } from '@/stores/useChatStore'
import ConfirmModal from '@/components/ui/ConfirmModal.vue'

const auth = useAuthStore()
const chatStore = useChatStore()
const router = useRouter()

const editing = ref(false)
const displayName = ref('')
const status = ref('')
const error = ref('')
const showLogoutModal = ref(false)
const isLoggingOut = ref(false)

const emit = defineEmits<{ 'update:isSidebarOpen': [value: boolean] }>()

defineProps<{ isSidebarOpen: boolean }>()

/**
 * Enters edit mode and populates fields with current profile values.
 */
function startEdit() {
  displayName.value = auth.user?.display_name ?? ''
  status.value = auth.user?.status ?? ''
  editing.value = true
  error.value = ''
}

/**
 * Saves the updated profile fields and exits edit mode.
 */
async function saveEdit() {
  try {
    await auth.updateProfile({
      display_name: displayName.value.trim() || undefined,
      status: status.value.trim() || undefined,
    })
    editing.value = false
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to save'
  }
}

async function confirmLogout() {
  isLoggingOut.value = true

  try {
    chatStore.reset()
    await auth.logout()
    await router.push('/')
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to log out'
  } finally {
    isLoggingOut.value = false
    showLogoutModal.value = false
  }
}
</script>

<template>
  <div
    class="flex items-center gap-3 px-3 py-3 border-b border-slate-200/80 dark:border-slate-800 transition-colors"
  >
    <div
      class="w-10 h-10 rounded-full bg-emerald-200 ring-1 ring-emerald-500/10 dark:ring-emerald-400/10 dark:bg-emerald-800 flex items-center justify-center flex-shrink-0 text-emerald-800 dark:text-emerald-200 font-semibold text-sm select-none"
    >
      {{ auth.user?.display_name.charAt(0).toUpperCase() }}
    </div>

    <div class="flex-1 min-w-0">
      <template v-if="editing">
        <input
          v-model="displayName"
          type="text"
          placeholder="Display name"
          class="w-full text-sm bg-emerald-50 border border-emerald-200 rounded-lg px-2 py-0.5 text-slate-900 focus:outline-none dark:bg-slate-800 dark:border-slate-600 dark:text-slate-100 mb-1"
        />
        <input
          v-model="status"
          type="text"
          placeholder="Status"
          class="w-full text-sm bg-emerald-50 border border-emerald-200 rounded-lg px-2 py-0.5 text-gray-500 focus:outline-none dark:bg-slate-800 dark:border-slate-600 dark:text-slate-400"
        />
        <p v-if="error" class="text-sm text-rose-500 mt-0.5">{{ error }}</p>
      </template>
      <template v-else>
        <p class="text-sm font-medium text-slate-900 dark:text-slate-100 truncate">
          {{ auth.user?.display_name }}
        </p>
        <p class="text-xs text-gray-500 dark:text-slate-400 truncate">
          {{ auth.user?.status }}
        </p>
      </template>
    </div>

    <div class="flex items-center gap-1 flex-shrink-0">
      <template v-if="editing">
        <button
          @click="saveEdit"
          class="p-1.5 hover:cursor-pointer text-emerald-600 hover:bg-emerald-100 rounded-lg transition-colors dark:hover:bg-slate-700"
        >
          <CheckIcon class="w-4 h-4" />
        </button>
        <button
          @click="editing = false"
          class="p-1.5 hover:cursor-pointer text-gray-500 hover:bg-gray-100 rounded-lg transition-colors dark:hover:bg-slate-700"
        >
          <XMarkIcon class="w-4 h-4" />
        </button>
      </template>
      <template v-else>
        <button
          @click="startEdit"
          class="p-2 hover:cursor-pointer text-gray-900 hover:text-emerald-600 hover:bg-slate-100 rounded-full transition-colors dark:text-gray-100 dark:hover:bg-slate-700"
        >
          <PencilIcon class="size-4" />
        </button>
        <button
          @click="showLogoutModal = true"
          class="p-2 hover:cursor-pointer text-gray-900 hover:text-rose-600 hover:bg-slate-100 rounded-full transition-colors dark:text-gray-100 dark:hover:bg-slate-800"
          title="Sign out"
        >
          <ArrowRightStartOnRectangleIcon class="size-4" />
        </button>
        <div class="ml-1.5 h-5 w-px bg-slate-300/80 dark:bg-slate-700/80"></div>
        <button
          @click="emit('update:isSidebarOpen', false)"
          class="flex size-10 items-center justify-center rounded-full text-slate-900 transition-all hover:cursor-pointer hover:bg-slate-100 active:scale-95 dark:text-slate-100 dark:hover:bg-slate-800"
          title="Collapse sidebar"
        >
          <ChevronLeftIcon class="size-5" />
        </button>
      </template>
    </div>
  </div>

  <Transition name="modal">
    <ConfirmModal
      v-if="showLogoutModal"
      title="Log out?"
      message="Are you sure you want to log out of your account?"
      confirm-text="Log out"
      danger
      :loading="isLoggingOut"
      @confirm="confirmLogout"
      @cancel="showLogoutModal = false"
    />
  </Transition>
</template>
