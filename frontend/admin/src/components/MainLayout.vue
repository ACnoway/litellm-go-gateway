<template>
  <div class="min-h-screen bg-slate-50 flex">
    <Sidebar />

    <div class="lg:ml-64 flex flex-col flex-1">
      <!-- Header -->
      <header class="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-6 sticky top-0 z-10">
        <div class="flex items-center">
          <span class="text-gray-700 font-medium">{{ pageTitle }}</span>
        </div>

        <div class="flex items-center space-x-4">
          <div class="flex items-center px-3 py-1.5 bg-gray-50 rounded-md">
            <svg class="w-4 h-4 mr-2 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
            </svg>
            <span class="text-xs text-gray-600">
              {{ hasApiKey ? '已配置' : '未配置' }}
            </span>
          </div>

          <button
            @click="showSettings"
            class="p-2 text-gray-500 hover:bg-gray-50 rounded-md"
            title="设置"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </button>
        </div>
      </header>

      <!-- Settings Modal -->
      <SettingsModal v-if="settingsStore.showSettings" @close="settingsStore.toggleSettings()" />

      <!-- Main Content -->
      <main class="flex-1 p-6">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import Sidebar from './Sidebar.vue'
import SettingsModal from './SettingsModal.vue'
import { useSettingsStore } from '../stores/useSettings'

const route = useRoute()
const settingsStore = useSettingsStore()

const hasApiKey = computed(() => !!localStorage.getItem('GATEWAY_API_KEY'))

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/admin': '仪表盘',
    '/admin/deployments': 'Deployments',
    '/admin/models': 'Models',
    '/admin/providers': 'Providers',
  }
  return titles[route.path] || '管理后台'
})

function showSettings() {
  settingsStore.toggleSettings()
}

onMounted(() => {
  // Check if API key is set on mount
  if (!hasApiKey.value) {
    settingsStore.toggleSettings()
  }
})
</script>
