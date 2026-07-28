<template>
  <div class="max-w-6xl mx-auto">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Providers</h1>

    <!-- Providers Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="provider in providers"
        :key="provider.name"
        class="bg-white rounded-lg shadow p-6 border-l-4"
        :class="provider.available ? 'border-green-500' : 'border-gray-300'"
      >
        <div class="flex items-start justify-between">
          <div class="flex items-center">
            <div class="w-12 h-12 rounded-lg flex items-center justify-center"
              :class="provider.available ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-500'">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <div class="ml-4">
              <h3 class="text-lg font-semibold text-gray-900">{{ provider.name }}</h3>
              <p class="text-sm text-gray-500">Type: {{ provider.type }}</p>
            </div>
          </div>
          <span
            :class="[
              'inline-flex items-center px-3 py-1 rounded-full text-xs font-medium',
              provider.available
                ? 'bg-green-100 text-green-800'
                : 'bg-gray-100 text-gray-600'
            ]"
          >
            {{ provider.available ? '可用' : '未配置' }}
          </span>
        </div>

        <div class="mt-4 pt-4 border-t border-gray-100">
          <div class="flex items-center text-sm text-gray-500">
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
            <span>Base URL: {{ provider.base_url || '默认' }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Info Note -->
    <div class="mt-8 bg-blue-50 border border-blue-200 rounded-md p-4">
      <div class="flex">
        <svg class="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
        </svg>
        <div class="ml-3 flex-1">
          <h3 class="text-sm font-medium text-blue-800">提示</h3>
          <div class="mt-2 text-sm text-blue-700">
            <p>Provider 在配置相应的 API Key 后会自动注册。例如配置 OPENAI_API_KEY 后，OpenAI provider 会自动可用。</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useProvidersStore } from '../stores/useProviders'

const providersStore = useProvidersStore()

const providers = providersStore.providers
const loading = providersStore.loading
const error = providersStore.error

onMounted(async () => {
  await providersStore.listProviders()
})
</script>
