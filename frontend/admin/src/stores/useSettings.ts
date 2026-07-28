import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useSettingsStore = defineStore('settings', () => {
  const apiKey = ref<string>(localStorage.getItem('GATEWAY_API_KEY') || '')
  const showSettings = ref(false)

  function setApiKey(key: string) {
    apiKey.value = key
    localStorage.setItem('GATEWAY_API_KEY', key)
  }

  function clearApiKey() {
    apiKey.value = ''
    localStorage.removeItem('GATEWAY_API_KEY')
  }

  function toggleSettings() {
    showSettings.value = !showSettings.value
  }

  return {
    apiKey,
    showSettings,
    setApiKey,
    clearApiKey,
    toggleSettings,
  }
})
