import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ProviderInfo } from '../api/types'
import { providersApi } from '../api/client'

export const useProvidersStore = defineStore('providers', () => {
  const providers = ref<ProviderInfo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loaded = computed(() => providers.value.length > 0)

  async function listProviders() {
    loading.value = true
    error.value = null
    try {
      const response = await providersApi.list()
      providers.value = response.data.data || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取 providers 失败'
    } finally {
      loading.value = false
    }
  }

  function getProvider(name: string): ProviderInfo | undefined {
    return providers.value.find(p => p.name === name)
  }

  function getAvailableProviders(): ProviderInfo[] {
    return providers.value.filter(p => p.available)
  }

  function clear() {
    providers.value = []
  }

  return {
    providers,
    loading,
    error,
    loaded,
    listProviders,
    getProvider,
    getAvailableProviders,
    clear,
  }
})
