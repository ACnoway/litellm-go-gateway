<template>
  <div class="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
    <!-- Header -->
    <header class="sticky top-0 z-30 bg-white/80 backdrop-blur-md border-b border-gray-200">
      <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <!-- Logo -->
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 bg-gradient-to-br from-blue-600 to-indigo-700 rounded-xl flex items-center justify-center shadow-lg shadow-blue-500/20">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <h1 class="text-xl font-bold text-gray-900 tracking-tight">LiteLLM Gateway</h1>
              <p class="text-xs text-gray-500 font-medium">AI API Proxy</p>
            </div>
          </div>

          <!-- Status & Settings -->
          <div class="flex items-center space-x-3">
            <!-- API Status -->
            <div class="hidden sm:flex items-center px-3 py-1.5 bg-gray-50 border border-gray-200 rounded-lg">
              <div :class="[
                'w-2 h-2 rounded-full mr-2',
                hasApiKey ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]' : 'bg-gray-300'
              ]"></div>
              <span class="text-sm font-medium text-gray-700">
                {{ hasApiKey ? 'API 已配置' : '未配置 API' }}
              </span>
            </div>

            <!-- Settings Button -->
            <button
              @click="showSettings"
              class="w-10 h-10 flex items-center justify-center rounded-lg text-gray-600 hover:bg-gray-100 hover:text-gray-900 transition-colors relative group"
            >
              <svg class="w-5 h-5 transition-transform group-hover:rotate-90" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- Settings Modal -->
    <SettingsModal v-if="settingsStore.showSettings" @close="settingsStore.toggleSettings()" />

    <!-- Main Content -->
    <main class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <router-view />
    </main>

    <!-- Footer -->
    <footer class="border-t border-gray-200 bg-white mt-auto">
      <div class="max-w-5xl mx-auto px-4 py-6 flex flex-col md:flex-row items-center justify-between text-sm text-gray-500">
        <p>© 2026 LiteLLM Gateway</p>
        <div class="flex items-center space-x-4 mt-2 md:mt-0">
          <span class="flex items-center">
            <span class="w-2 h-2 bg-green-500 rounded-full mr-2 animate-pulse"></span>
            Service Available
          </span>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import SettingsModal from './SettingsModal.vue'
import { useSettingsStore } from '../stores/useSettings'

const settingsStore = useSettingsStore()

const hasApiKey = computed(() => !!localStorage.getItem('GATEWAY_API_KEY'))

function showSettings() {
  settingsStore.toggleSettings()
}
</script>
