<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    message: string
    confirmText?: string
    cancelText?: string
    danger?: boolean
    loading?: boolean
  }>(),
  {
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    danger: false,
    loading: false,
  },
)

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/30 p-4 backdrop-blur-sm dark:bg-black/50"
    @click.self="emit('cancel')"
  >
    <div
      class="w-full max-w-sm rounded-3xl border border-slate-200/70 bg-white p-6 shadow-2xl shadow-slate-950/10 dark:border-white/10 dark:bg-slate-900 dark:shadow-black/30"
    >
      <h2 class="text-lg font-semibold text-slate-900 dark:text-white">
        {{ title }}
      </h2>

      <p class="mt-2 text-sm leading-relaxed text-slate-500 dark:text-slate-400">
        {{ message }}
      </p>

      <div class="mt-6 flex justify-end gap-2">
        <button
          type="button"
          :disabled="loading"
          class="rounded-xl px-4 py-2 text-sm font-medium text-slate-600 transition-colors hover:cursor-pointer hover:bg-slate-100 disabled:opacity-50 dark:text-slate-300 dark:hover:bg-slate-800"
          @click="emit('cancel')"
        >
          {{ cancelText }}
        </button>

        <button
          type="button"
          :disabled="loading"
          :class="[
            'rounded-xl px-4 py-2 text-sm font-medium text-white transition-all hover:cursor-pointer active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50',
            danger
              ? 'bg-rose-600 hover:bg-rose-700 dark:hover:bg-rose-500'
              : 'bg-emerald-600 hover:bg-emerald-700 dark:hover:bg-emerald-500',
          ]"
          @click="emit('confirm')"
        >
          {{ loading ? 'Please wait...' : confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>
