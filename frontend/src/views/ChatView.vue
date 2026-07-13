<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted, onMounted } from 'vue'
import {
  MoonIcon,
  SunIcon,
  UsersIcon,
  Bars3Icon,
  PencilSquareIcon,
  MagnifyingGlassIcon,
  XMarkIcon,
  TrashIcon,
} from '@heroicons/vue/24/outline'
import { useDark, useToggle, useDebounceFn } from '@vueuse/core'
import { useAuthStore } from '@/stores/useAuthStore'
import { useChatStore } from '@/stores/useChatStore'
import ProfilePanel from '@/components/chat/ProfilePanel.vue'
import SidebarItem from '@/components/chat/SidebarItem.vue'
import MessageBubble from '@/components/chat/MessageBubble.vue'
import MessageInput from '@/components/chat/MessageInput.vue'
import NewChatModal from '@/components/chat/NewChatModal.vue'
import ConfirmModal from '@/components/ui/ConfirmModal.vue'

const auth = useAuthStore()
const chatStore = useChatStore()

const isDark = useDark()
const toggleDark = useToggle(isDark)

const showNewChat = ref(false)
const messagesEl = ref<HTMLElement | null>(null)
const loadingMore = ref(false)

const sidebarWidth = ref(340)
const MIN_SIDEBAR_WIDTH = 280
const MAX_SIDEBAR_WIDTH = 520
const COLLAPSED_SIDEBAR_WIDTH = 64

const isSidebarOpen = ref(window.innerWidth >= 768)

const sidebarSearch = ref('')
const isSearchOpen = ref(false)
const messageSearch = ref('')

const showDeleteChatModal = ref(false)
const isDeletingChat = ref(false)

const touchStartX = ref(0)
const touchStartY = ref(0)

chatStore.loadChats()

/**
 * Returns the display name for the active chat header.
 */
const activeChatName = computed(() => {
  const chat = chatStore.activeChat
  if (!chat) return ''
  if (chat.is_group) return chat.name ?? 'Group Chat'
  const other = chat.members.find((m) => m.id !== auth.user?.id)
  return other?.display_name ?? 'Chat'
})

/**
 * Returns the subtitle shown under the active chat name.
 */
const activeChatSub = computed(() => {
  const chat = chatStore.activeChat
  if (!chat) return ''
  if (chat.is_group) return `${chat.members.length} members`
  const other = chat.members.find((m) => m.id !== auth.user?.id)
  return other ? `@${other.username} · ${other.status}` : ''
})

/**
 * Filters the sidebar chat list by the sidebar search query.
 */
const filteredChats = computed(() => {
  const query = sidebarSearch.value.trim().toLowerCase()
  if (!query) return chatStore.chats
  return chatStore.chats.filter((chat) => {
    if (chat.is_group) return (chat.name ?? '').toLowerCase().includes(query)
    const other = chat.members.find((m) => m.id !== auth.user?.id)
    return (
      other?.display_name.toLowerCase().includes(query) ||
      other?.username.toLowerCase().includes(query)
    )
  })
})

const runMessageSearch = useDebounceFn(async (query: string) => {
  await chatStore.searchMessages(query)
}, 350)

/**
 * Opens the message search bar and resets its state.
 */
function openSearch() {
  isSearchOpen.value = true
  messageSearch.value = ''
  chatStore.searchMessages('')
}

/**
 * Closes the message search bar and clears results.
 */
function closeSearch() {
  isSearchOpen.value = false
  messageSearch.value = ''
  chatStore.searchMessages('')
}

watch(messageSearch, (val) => runMessageSearch(val))

/**
 * Scrolls the message container to the bottom.
 */
async function scrollToBottom(smooth = false) {
  await nextTick()
  if (!messagesEl.value) return
  messagesEl.value.scrollTo({
    top: messagesEl.value.scrollHeight,
    behavior: smooth ? 'smooth' : 'instant',
  })
}

/**
 * Handles scroll-to-top to load older messages with position preservation.
 */
async function onScroll() {
  if (!messagesEl.value || loadingMore.value || isSearchOpen.value) return
  if (messagesEl.value.scrollTop > 80) return

  loadingMore.value = true
  const prevHeight = messagesEl.value.scrollHeight

  const hadMore = await chatStore.loadMoreMessages()
  if (hadMore) {
    await nextTick()
    messagesEl.value.scrollTop = messagesEl.value.scrollHeight - prevHeight
  }
  loadingMore.value = false
}

/**
 * Toggles the sidebar open/closed via Ctrl+B / Cmd+B keyboard shortcut.
 */
function onToggleSidebar(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'b') {
    e.preventDefault()
    isSidebarOpen.value = !isSidebarOpen.value
  }
}

/**
 * Starts a pointer-based drag to resize the sidebar width.
 */
function startSidebarResize(e: PointerEvent) {
  e.preventDefault()
  const startX = e.clientX
  const startWidth = sidebarWidth.value

  function onPointerMove(e: PointerEvent) {
    sidebarWidth.value = Math.min(
      MAX_SIDEBAR_WIDTH,
      Math.max(MIN_SIDEBAR_WIDTH, startWidth + e.clientX - startX),
    )
  }

  function onPointerUp() {
    window.removeEventListener('pointermove', onPointerMove)
    window.removeEventListener('pointerup', onPointerUp)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }

  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
  window.addEventListener('pointermove', onPointerMove)
  window.addEventListener('pointerup', onPointerUp)
}

watch(
  () => chatStore.activeChatId,
  () => {
    closeSearch()
    scrollToBottom()
  },
)

watch(
  () => chatStore.messages.length,
  (newLen, oldLen) => {
    if (!messagesEl.value || isSearchOpen.value) return
    const el = messagesEl.value
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 120
    if (newLen > oldLen && nearBottom) scrollToBottom(true)
  },
)

onMounted(() => window.addEventListener('keydown', onToggleSidebar))

onUnmounted(() => {
  window.removeEventListener('keydown', onToggleSidebar)
  chatStore.disconnectWs()
})

async function confirmDeleteChat() {
  const chatId = chatStore.activeChatId
  if (!chatId) return

  isDeletingChat.value = true

  try {
    await chatStore.deleteChat(chatId)
    showDeleteChatModal.value = false
  } finally {
    isDeletingChat.value = false
  }
}

function handleChatOpen() {
  if (window.innerWidth < 768) {
    isSidebarOpen.value = false
  }
}

function onTouchStart(e: TouchEvent) {
  const touch = e.changedTouches[0]
  if (!touch) return

  touchStartX.value = touch.clientX
  touchStartY.value = touch.clientY
}

function onTouchEnd(e: TouchEvent) {
  const touch = e.changedTouches[0]
  if (!touch) return

  const deltaX = touch.clientX - touchStartX.value
  const deltaY = touch.clientY - touchStartY.value

  if (Math.abs(deltaX) < 60 || Math.abs(deltaX) <= Math.abs(deltaY)) {
    return
  }

  if (deltaX > 0) {
    isSidebarOpen.value = true
  }
  if (deltaX < 0) {
    isSidebarOpen.value = false
  }
}
</script>

<template>
  <div
    @touchstart="onTouchStart"
    @touchend="onTouchEnd"
    class="flex h-screen overflow-hidden bg-slate-50 text-slate-900 transition-colors dark:bg-slate-950 dark:text-slate-100"
  >
    <!-- Sidebar -->
    <aside
      :style="{
        '--sidebar-width': `${sidebarWidth}px`,
      }"
      :class="[
        'flex flex-shrink-0 flex-col overflow-hidden border-r',
        'border-slate-200/80 bg-white/90',
        'transition-[width] duration-200 ease-out',
        'dark:border-slate-800/80 dark:bg-slate-900/80',

        isSidebarOpen ? 'w-full md:w-[var(--sidebar-width)]' : 'w-16',
      ]"
    >
      <!-- Full sidebar -->
      <div
        v-if="isSidebarOpen"
        class="flex h-full w-full flex-col pb-3 md:w-[var(--sidebar-width)]"
      >
        <ProfilePanel v-model:isSidebarOpen="isSidebarOpen" />

        <div class="flex items-center gap-2 p-3">
          <input
            v-model="sidebarSearch"
            type="text"
            placeholder="Search chats..."
            class="min-w-0 flex-1 rounded-full border border-slate-200 bg-slate-100/70 px-3 py-2 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-500 focus:bg-white focus:ring-2 focus:ring-emerald-500/10 dark:border-slate-700/80 dark:bg-slate-800/70 dark:text-slate-100 dark:placeholder:text-slate-500 dark:focus:border-emerald-500 dark:focus:bg-slate-800"
          />
          <button
            @click="showNewChat = true"
            class="flex size-10 items-center justify-center rounded-full text-slate-900 transition-all hover:cursor-pointer hover:bg-slate-100 active:scale-95 dark:text-slate-100 dark:hover:bg-slate-800"
            title="New chat"
          >
            <PencilSquareIcon class="size-5" />
          </button>
        </div>

        <ul class="no-scrollbar flex-1 space-y-1 overflow-y-auto px-2 pb-3">
          <SidebarItem
            v-for="chat in filteredChats"
            :key="chat.id"
            :chat="chat"
            @open="handleChatOpen"
          />
          <li
            v-if="filteredChats.length === 0"
            class="select-none py-12 text-center text-xs text-slate-400 dark:text-slate-500"
          >
            {{ sidebarSearch.trim() ? 'No chats found' : 'No chats yet. Start one!' }}
          </li>
        </ul>

        <button
          @click="toggleDark()"
          class="ml-3 flex size-10 items-center justify-center rounded-full text-slate-900 transition-all hover:cursor-pointer hover:bg-slate-100 active:scale-95 dark:text-slate-100 dark:hover:bg-slate-800"
          aria-label="Toggle dark mode"
        >
          <SunIcon v-if="isDark" class="size-5" />
          <MoonIcon v-else class="size-5" />
        </button>
      </div>

      <!-- Compact sidebar -->
      <div v-else class="flex h-full w-full flex-col items-center py-3">
        <button
          @click="isSidebarOpen = true"
          class="flex size-10 items-center justify-center rounded-full text-slate-900 transition-all hover:cursor-pointer hover:bg-slate-100 active:scale-95 dark:text-slate-100 dark:hover:bg-slate-800"
          title="Expand sidebar"
        >
          <Bars3Icon class="size-5" />
        </button>

        <button
          @click="showNewChat = true"
          class="mt-1 flex size-10 items-center justify-center rounded-full text-slate-900 transition-all hover:cursor-pointer hover:bg-slate-100 active:scale-95 dark:text-slate-100 dark:hover:bg-slate-800"
          title="New chat"
        >
          <PencilSquareIcon class="size-5" />
        </button>

        <div
          class="no-scrollbar mt-3 flex w-full flex-1 flex-col items-center gap-1 overflow-y-auto"
        >
          <SidebarItem
            v-for="chat in chatStore.chats"
            :key="chat.id"
            :chat="chat"
            compact
            @open="handleChatOpen"
          />
        </div>

        <button
          @click="toggleDark()"
          class="flex size-10 items-center justify-center rounded-full text-slate-900 transition-all hover:cursor-pointer hover:bg-slate-100 active:scale-95 dark:text-slate-100 dark:hover:bg-slate-800"
          aria-label="Toggle dark mode"
        >
          <SunIcon v-if="isDark" class="size-5" />
          <MoonIcon v-else class="size-5" />
        </button>
      </div>
    </aside>

    <!-- Resizer -->
    <div
      v-if="isSidebarOpen"
      class="relative z-10 hidden md:block w-1 flex-shrink-0 cursor-col-resize bg-slate-200/70 hover:bg-emerald-500 dark:bg-slate-800 dark:hover:bg-emerald-500"
      @pointerdown="startSidebarResize"
    >
      <div class="absolute inset-y-0 -left-1 -right-1"></div>
    </div>

    <!-- Main panel -->
    <main
      :class="[
        'relative min-w-0 flex-1 flex-col bg-slate-50/10 transition-colors dark:bg-slate-900/30',
        isSidebarOpen ? 'hidden md:flex' : 'flex',
      ]"
    >
      <template v-if="chatStore.activeChat">
        <!-- Chat header -->
        <header
          class="flex flex-shrink-0 items-center gap-3 border-b border-slate-200/80 bg-white/80 px-5 py-3 backdrop-blur-xl transition-colors dark:border-slate-800/80 dark:bg-slate-900/80"
        >
          <div
            class="flex size-10 flex-shrink-0 select-none items-center justify-center rounded-full bg-emerald-200/80 text-sm font-semibold text-emerald-700 ring-1 ring-emerald-500/10 dark:bg-emerald-400/20 dark:text-emerald-300 dark:ring-emerald-400/10"
          >
            {{ activeChatName.charAt(0).toUpperCase() }}
          </div>

          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-semibold text-slate-900 dark:text-slate-100">
              {{ activeChatName }}
            </p>
            <p class="truncate text-xs text-slate-500 dark:text-slate-400">
              {{ activeChatSub }}
            </p>
          </div>

          <div class="flex items-center gap-1">
            <!-- Search bar -->
            <Transition name="search">
              <input
                v-if="isSearchOpen"
                v-model="messageSearch"
                type="text"
                placeholder="Search in chat..."
                autofocus
                class="w-57 rounded-full border border-slate-200 bg-slate-100/70 px-4 py-1.5 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/10 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:placeholder:text-slate-500"
              />
            </Transition>

            <button
              @click="isSearchOpen ? closeSearch() : openSearch()"
              class="flex size-9 items-center justify-center rounded-full text-slate-500 transition-all hover:cursor-pointer hover:bg-slate-100 active:scale-95 dark:text-slate-400 dark:hover:bg-slate-800"
              :aria-label="isSearchOpen ? 'Close search' : 'Search messages'"
            >
              <XMarkIcon v-if="isSearchOpen" class="size-5" />
              <MagnifyingGlassIcon v-else class="size-5" />
            </button>

            <div
              v-if="chatStore.activeChat.is_group"
              class="flex size-9 items-center justify-center rounded-full text-slate-400"
            >
              <UsersIcon class="size-5" />
            </div>

            <button
              @click="showDeleteChatModal = true"
              class="flex size-9 items-center justify-center rounded-full text-slate-500 transition-all hover:cursor-pointer hover:bg-rose-50 hover:text-rose-500 active:scale-95 dark:text-slate-400 dark:hover:bg-rose-900/20 dark:hover:text-rose-400"
              title="Delete chat"
            >
              <TrashIcon class="size-5" />
            </button>
          </div>
        </header>

        <!-- Search mode banner -->
        <Transition name="search-banner">
          <div
            v-if="isSearchOpen"
            class="flex items-center justify-between border-b border-slate-200/60 bg-emerald-50/60 px-5 py-2 text-xs dark:border-slate-800/60 dark:bg-emerald-900/10"
          >
            <span class="text-slate-500 dark:text-slate-400">
              <template v-if="chatStore.searchLoading">Searching...</template>
              <template v-else-if="messageSearch.trim()">
                {{ chatStore.searchResults.length }} result{{
                  chatStore.searchResults.length !== 1 ? 's' : ''
                }}
                for
                <span class="font-medium text-slate-700 dark:text-slate-200">
                  "{{ messageSearch }}"
                </span>
              </template>
              <template v-else>Type to search messages in this chat</template>
            </span>

            <button
              @click="closeSearch"
              class="text-slate-400 transition-colors hover:cursor-pointer hover:text-slate-700 dark:hover:text-slate-200"
            >
              Clear
            </button>
          </div>
        </Transition>

        <!-- Messages -->
        <div
          ref="messagesEl"
          @scroll="onScroll"
          class="no-scrollbar flex-1 space-y-0.5 overflow-y-auto px-5 py-5 pb-24"
        >
          <div
            v-if="loadingMore && !isSearchOpen"
            class="py-2 text-center text-xs text-slate-400 dark:text-slate-500"
          >
            Loading older messages...
          </div>

          <div v-if="chatStore.loadingMessages" class="flex h-full items-center justify-center">
            <span class="text-sm text-slate-400 dark:text-slate-500">Loading messages...</span>
          </div>

          <template v-else-if="isSearchOpen && messageSearch.trim() && !chatStore.searchLoading">
            <div
              v-if="chatStore.searchResults.length === 0"
              class="flex h-full items-center justify-center"
            >
              <span class="text-sm text-slate-400 dark:text-slate-500">No messages found</span>
            </div>
            <MessageBubble
              v-else
              v-for="msg in chatStore.searchResults"
              :key="msg.id"
              :message="msg"
              :is-group="chatStore.activeChat.is_group"
            />
          </template>

          <template v-else-if="!isSearchOpen || !messageSearch.trim()">
            <MessageBubble
              v-for="msg in chatStore.messages"
              :key="msg.id"
              :message="msg"
              :is-group="chatStore.activeChat.is_group"
            />
          </template>
        </div>

        <div class="absolute inset-x-0 bottom-0 flex justify-center p-2">
          <div
            class="w-full transition-all duration-300 ease-in-out"
            :class="isSidebarOpen ? 'max-w-full' : 'max-w-4xl'"
          >
            <MessageInput />
          </div>
        </div>
      </template>

      <!-- Empty state -->
      <div
        v-else
        class="relative flex flex-1 select-none flex-col items-center justify-center px-6"
      >
        <div
          class="pointer-events-none absolute size-[28rem] rounded-full bg-emerald-300/10 blur-3xl dark:bg-emerald-500/5"
        ></div>

        <div class="flex flex-col items-center text-center">
          <div class="mb-5 flex size-20 items-center justify-center">
            <img
              src="/icon.svg"
              alt="LightChat logo"
              class="size-15 drop-shadow-[0_0_10px_rgba(16,185,129,0.8)]"
            />
          </div>

          <h1 class="text-center text-2xl font-semibold text-slate-900 dark:text-white">
            Welcome to
            <span
              class="bg-gradient-to-r from-emerald-600 to-cyan-600 bg-clip-text text-transparent dark:from-emerald-500 dark:to-cyan-500"
            >
              LightChat
            </span>
          </h1>

          <p
            class="mt-2 max-w-xs text-center text-sm leading-relaxed text-slate-500 dark:text-slate-400"
          >
            Select a conversation from the sidebar or start a new chat.
          </p>

          <button
            @click="showNewChat = true"
            class="group mt-6 flex items-center gap-2 rounded-2xl bg-gradient-to-r from-emerald-600 to-teal-600 px-5 py-3 text-sm font-semibold text-white shadow-lg shadow-emerald-600/20 transition-all duration-200 hover:-translate-y-0.5 hover:from-emerald-500 hover:to-teal-500 hover:shadow-xl hover:shadow-emerald-600/25 active:translate-y-0 active:scale-[0.97] hover:cursor-pointer"
          >
            <PencilSquareIcon
              class="size-4.5 transition-transform duration-200 group-hover:rotate-5"
            />
            New Chat
          </button>
        </div>
      </div>
    </main>
  </div>

  <Transition name="modal">
    <NewChatModal v-if="showNewChat" @close="showNewChat = false" />
  </Transition>

  <Transition name="modal">
    <ConfirmModal
      v-if="showDeleteChatModal"
      title="Delete chat?"
      :message="`Are you sure you want to delete this chat (${activeChatName})? This action cannot be undone.`"
      confirm-text="Delete"
      danger
      :loading="isDeletingChat"
      @confirm="confirmDeleteChat"
      @cancel="showDeleteChatModal = false"
    />
  </Transition>
</template>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  scrollbar-width: none;
}

.search-enter-active,
.search-leave-active {
  transition: all 200ms ease;
}
.search-enter-from,
.search-leave-to {
  opacity: 0;
  width: 0;
  padding-left: 0;
  padding-right: 0;
}

.search-banner-enter-active,
.search-banner-leave-active {
  transition: all 250ms ease-out;
  overflow: hidden;
}

.search-banner-enter-from,
.search-banner-leave-to {
  opacity: 0;
  transform: translateY(-10px);
  max-height: 0;
}

.search-banner-enter-to,
.search-banner-leave-from {
  opacity: 1;
  transform: translateY(0);
  max-height: 80px;
}
</style>
