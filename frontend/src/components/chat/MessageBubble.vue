<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/useAuthStore'
import type { Message } from '@/types'
import { CheckIcon } from '@heroicons/vue/24/outline'

const props = defineProps<{ message: Message; isGroup: boolean }>()

const auth = useAuthStore()

const isOwn = computed(() => props.message.user_id === auth.user?.id)

/**
 * Formats a timestamp into a short HH:MM string.
 */
function formatTime(timestamp: string): string {
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div :class="['mb-2 flex w-full', isOwn ? 'justify-end' : 'justify-start']">
    <div
      :class="[
        'flex min-w-0 max-w-[75%] flex-col sm:max-w-[70%]',
        isOwn ? 'items-end' : 'items-start',
      ]"
    >
      <span
        v-if="!isOwn && isGroup"
        class="mb-1 px-1 text-xs font-medium text-emerald-600 dark:text-emerald-400"
      >
        {{ message.sender_display_name }}
      </span>

      <div
        :class="[
          'min-w-0 max-w-full whitespace-pre-wrap break-words [overflow-wrap:anywhere]',
          'rounded-3xl px-3.5 py-2.5 text-[15px] leading-relaxed',
          isOwn
            ? 'rounded-br-md border bg-emerald-200 dark:bg-emerald-800 shadow-sm shadow-emerald-600/10 dark:shadow-emerald-400/20 border-slate-200/70 dark:border-white/10'
            : 'rounded-bl-md border border-slate-200/70 bg-white text-slate-800 shadow-sm shadow-emerald-600/10 dark:shadow-emerald-200/10 dark:border-white/5 dark:bg-slate-900 dark:text-slate-100',
        ]"
      >
        {{ message.content }}
      </div>

      <div class="mt-1 flex items-center gap-1 px-1">
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
</template>
