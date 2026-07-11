<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/useAuthStore'
import { MoonIcon, SunIcon } from '@heroicons/vue/24/outline'
import { useDark, useToggle } from '@vueuse/core'

const router = useRouter()
const auth = useAuthStore()

const isDark = useDark()
const toggleDark = useToggle(isDark)

const mode = ref<'login' | 'register'>('login')
const email = ref('')
const username = ref('')
const displayName = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

function changeMode(m: 'login' | 'register') {
  mode.value = m
  error.value = ''
}

/**
 * Handles user authentication by executing either login or registration based on the current mode.
 */
async function submit() {
  error.value = ''
  loading.value = true
  try {
    if (mode.value === 'login') {
      await auth.login(email.value, password.value)
    } else {
      await auth.register(email.value, username.value, displayName.value, password.value)
    }
    router.push('/chat')
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div
    class="relative min-h-screen overflow-hidden bg-emerald-50/30 flex items-center justify-center p-4 dark:bg-slate-950 transition-colors"
  >
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div
        class="absolute -top-40 -left-40 h-[36rem] w-[36rem] rounded-full bg-emerald-300/20 blur-3xl dark:bg-emerald-500/10"
      ></div>
      <div
        class="absolute -bottom-48 -right-40 h-[36rem] w-[36rem] rounded-full bg-teal-300/20 blur-3xl dark:bg-teal-500/10"
      ></div>
    </div>

    <button
      @click="toggleDark()"
      class="absolute top-4 right-4 z-10 p-2 rounded-xl text-slate-900 hover:text-emerald-600 dark:text-slate-300 dark:hover:text-emerald-400 hover:cursor-pointer transform transition-transform active:scale-110"
      aria-label="Toggle dark mode"
    >
      <SunIcon v-if="isDark" class="w-6 h-6" />
      <MoonIcon v-else class="w-6 h-6" />
    </button>

    <div class="relative z-10 w-full max-w-sm">
      <div class="flex flex-col items-center justify-center gap-2 mb-1">
        <img
          src="/icon.svg"
          alt="LightChat logo"
          class="size-11 drop-shadow-[0_0_20px_rgba(16,185,129,0.8)]"
        />

        <h1 class="text-3xl font-bold tracking-tight dark:text-slate-100">
          Light<span class="text-emerald-500 dark:text-emerald-400">Chat</span>
        </h1>
      </div>
      <p class="text-slate-500 text-sm text-center mb-7 dark:text-slate-400">
        Fast, minimal messaging.
      </p>

      <div
        class="flex bg-white rounded-2xl p-1 mb-5 border border-emerald-200 dark:bg-slate-900 dark:border-slate-800 transition-colors"
      >
        <button
          v-for="m in ['login', 'register']"
          :key="m"
          @click="changeMode(m)"
          :class="[
            'flex-1 py-1.5 border border-transparent rounded-xl text-sm font-medium hover:cursor-pointer transition-all capitalize',
            mode === m
              ? 'bg-emerald-600 text-white shadow dark:bg-emerald-600'
              : 'text-gray-600 hover:border-emerald-300 dark:text-slate-400 dark:hover:border-slate-700',
          ]"
        >
          {{ m === 'login' ? 'Sign In' : 'Register' }}
        </button>
      </div>

      <form @submit.prevent="submit" class="space-y-3">
        <div>
          <label class="block text-xs text-gray-600 mb-1 dark:text-slate-400">Email</label>
          <input
            v-model="email"
            type="email"
            required
            placeholder="you@example.com"
            class="w-full bg-white border border-emerald-200 rounded-2xl px-3 py-2.5 text-black text-sm placeholder-gray-400 focus:outline-none focus:border-emerald-500 transition-colors dark:bg-slate-800 dark:border-slate-700 dark:text-slate-100 dark:placeholder-slate-500 dark:focus:border-emerald-500"
          />
        </div>

        <template v-if="mode === 'register'">
          <div>
            <label class="block text-xs text-gray-600 mb-1 dark:text-slate-400">Username</label>
            <input
              v-model="username"
              type="text"
              required
              placeholder="your_username"
              class="w-full bg-white border border-emerald-200 rounded-2xl px-3 py-2.5 text-black text-sm placeholder-gray-400 focus:outline-none focus:border-emerald-500 transition-colors dark:bg-slate-800 dark:border-slate-700 dark:text-slate-100 dark:placeholder-slate-500 dark:focus:border-emerald-500"
            />
          </div>
          <div>
            <label class="block text-xs text-gray-600 mb-1 dark:text-slate-400">Display Name</label>
            <input
              v-model="displayName"
              type="text"
              required
              placeholder="Your Name"
              class="w-full bg-white border border-emerald-200 rounded-2xl px-3 py-2.5 text-black text-sm placeholder-gray-400 focus:outline-none focus:border-emerald-500 transition-colors dark:bg-slate-800 dark:border-slate-700 dark:text-slate-100 dark:placeholder-slate-500 dark:focus:border-emerald-500"
            />
          </div>
        </template>

        <div>
          <label class="block text-xs text-gray-600 mb-1 dark:text-slate-400">Password</label>
          <input
            v-model="password"
            type="password"
            required
            placeholder="••••••••••"
            class="w-full bg-white border border-emerald-200 rounded-2xl px-3 py-2.5 text-black text-sm placeholder-gray-400 focus:outline-none focus:border-emerald-500 transition-colors dark:bg-slate-800 dark:border-slate-700 dark:text-slate-100 dark:placeholder-slate-500 dark:focus:border-emerald-500"
          />
        </div>

        <p v-if="error" class="text-rose-600 text-xs pt-1 dark:text-rose-400">{{ error }}</p>

        <button
          type="submit"
          :disabled="loading"
          class="w-full hover:shadow bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed hover:cursor-pointer text-white font-semibold rounded-xl py-2.5 text-sm transition-colors mt-2 dark:bg-emerald-600 dark:hover:bg-emerald-500"
        >
          {{ loading ? 'Loading…' : mode === 'login' ? 'Sign In' : 'Create Account' }}
        </button>
      </form>
    </div>
  </div>
</template>
