import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ModelInfo } from '../api/types'
import { modelsApi } from '../api/client'

export const useModelsStore = defineStore('models', () => {
  const models = ref<ModelInfo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loaded = computed(() => models.value.length > 0)

  async function listModels() {
    loading.value = true
    error.value = null
    try {
      const response = await modelsApi.list()
      models.value = response.data.data || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取 models 失败'
    } finally {
      loading.value = false
    }
  }

  function getReadyModels(): ModelInfo[] {
    return models.value.filter(m => m.ready)
  }

  function clear() {
    models.value = []
  }

  return {
    models,
    loading,
    error,
    loaded,
    listModels,
    getReadyModels,
    clear,
  }
})
