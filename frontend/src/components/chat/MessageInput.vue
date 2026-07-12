<script setup lang="ts">
import { ref } from 'vue'
import { PaperAirplaneIcon } from '@heroicons/vue/24/solid'
import { useChatStore } from '@/stores/useChatStore'
import { useTextareaAutosize } from '@vueuse/core'

const chatStore = useChatStore()

const text = ref('')

const { textarea, triggerResize } = useTextareaAutosize({
  input: text,
})

/**
 * Sends the message if the text is non-empty and resets the input.
 */
function send() {
  const trimmed = text.value.trim()
  if (!trimmed) return

  chatStore.sendMessage(trimmed)
  text.value = ''
  triggerResize()
}

/**
 * Intercepts Enter without Shift to trigger send instead of a newline.
 */
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}
</script>

<template>
  <div
    class="flex items-end gap-2 bg-white shadow-sm dark:bg-gray-900 px-2 m-3 py-2 rounded-4xl border border-slate-300/30"
  >
    <textarea
      ref="textarea"
      v-model="text"
      rows="1"
      maxlength="2000"
      placeholder="Write a message..."
      @keydown="onKeydown"
      class="max-h-38 flex-1 resize-none overflow-y-auto rounded-3xl bg-white dark:bg-gray-900 px-4 py-2.5 text-md text-slate-900 outline-none transition-colors placeholder:text-slate-400 dark:text-slate-100 dark:placeholder:text-slate-500 [&::-webkit-scrollbar]:hidden [scrollbar-width:none]"
    ></textarea>

    <button
      @click="send"
      :disabled="!text.trim()"
      class="flex-shrink-0 rounded-3xl bg-emerald-600 p-2.5 mr-2 mb-0.5 text-white transition-all hover:cursor-pointer hover:bg-emerald-700 active:scale-95 disabled:cursor-not-allowed disabled:opacity-40 dark:hover:bg-emerald-500"
      aria-label="Send message"
    >
      <PaperAirplaneIcon class="size-5" />
    </button>
  </div>
</template>
