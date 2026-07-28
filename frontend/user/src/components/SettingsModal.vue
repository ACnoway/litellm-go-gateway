<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-gray-900/50 backdrop-blur-sm"
    @click.self="emit('close')"
  >
    <div class="bg-white rounded-lg shadow-xl w-full max-w-md overflow-hidden">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-900">设置</h3>
        <button
          @click="emit('close')"
          class="text-gray-400 hover:text-gray-600 transition-colors"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Body -->
      <div class="px-6 py-6 space-y-6">
        <!-- API Key -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            API Key
            <span class="text-red-500 ml-1">*</span>
          </label>
          <div class="relative">
            <input
              v-model="apiKeyInput"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              placeholder="sk-..."
              autocomplete="off"
            />
            <button
              v-if="apiKeyInput"
              @click="copyApiKey"
              class="absolute right-2 top-2 p-1 text-gray-400 hover:text-gray-600"
              title="复制"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
              </svg>
            </button>
          </div>
          <p class="mt-1 text-xs text-gray-500">
            使用您的 OpenAI API Key 或其他支持的 LLM 提供商 API Key。
          </p>
        </div>

        <!-- Actions -->
        <div class="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
          <button
            v-if="hasApiKey"
            @click="clearApiKey"
            class="px-4 py-2 text-sm font-medium text-red-700 bg-red-50 hover:bg-red-100 rounded-md transition-colors"
          >
            清除
          </button>
          <button
            @click="saveSettings"
            :disabled="!apiKeyInput.trim() || saving"
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ saving ? '保存中...' : '保存设置' }}
          </button>
        </div>

        <!-- Status -->
        <div v-if="saveStatus" :class="[
          'text-sm px-3 py-2 rounded-md',
          saveStatus.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
        ]">
          {{ saveStatus.message }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useSettingsStore } from '../stores/useSettings'

const emit = defineEmits(['close'])
const settingsStore = useSettingsStore()
const apiKeyInput = ref('')
const saving = ref(false)
const saveStatus = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const hasApiKey = ref(false)

watch(() => settingsStore.showSettings, (visible) => {
  if (visible) {
    apiKeyInput.value = localStorage.getItem('GATEWAY_API_KEY') || ''
    hasApiKey.value = !!apiKeyInput.value
  }
})

onMounted(() => {
  apiKeyInput.value = localStorage.getItem('GATEWAY_API_KEY') || ''
  hasApiKey.value = !!apiKeyInput.value
})

function copyApiKey() {
  if (apiKeyInput.value) {
    navigator.clipboard.writeText(apiKeyInput.value)
    showStatus('success', '已复制到剪贴板')
  }
}

function clearApiKey() {
  settingsStore.clearApiKey()
  apiKeyInput.value = ''
  hasApiKey.value = false
  showStatus('success', 'API Key 已清除')
}

function saveSettings() {
  if (!apiKeyInput.value.trim()) {
    showStatus('error', 'API Key 不能为空')
    return
  }
  saving.value = true
  settingsStore.setApiKey(apiKeyInput.value)
  hasApiKey.value = true
  showStatus('success', '设置已保存')
  setTimeout(() => {
    saving.value = false
    settingsStore.toggleSettings()
  }, 1000)
}

function showStatus(type: 'success' | 'error', message: string) {
  saveStatus.value = { type, message }
  setTimeout(() => {
    saveStatus.value = null
  }, 3000)
}
</script>
