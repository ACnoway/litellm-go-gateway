import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Deployment, DeploymentRequest } from '../api/types'
import { deploymentsApi } from '../api/client'

export const useDeploymentsStore = defineStore('deployments', () => {
  const deployments = ref<Deployment[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loaded = computed(() => deployments.value.length > 0)

  async function listDeployments() {
    loading.value = true
    error.value = null
    try {
      const response = await deploymentsApi.list()
      deployments.value = response.data.data || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取 deployments 失败'
    } finally {
      loading.value = false
    }
  }

  async function getDeployment(id: number) {
    loading.value = true
    error.value = null
    try {
      const response = await deploymentsApi.get(id)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取 deployment 失败'
      return null
    } finally {
      loading.value = false
    }
  }

  async function createDeployment(data: DeploymentRequest) {
    loading.value = true
    error.value = null
    try {
      const response = await deploymentsApi.create(data)
      deployments.value.push(response.data)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '创建 deployment 失败'
      return null
    } finally {
      loading.value = false
    }
  }

  async function updateDeployment(id: number, data: DeploymentRequest) {
    loading.value = true
    error.value = null
    try {
      const response = await deploymentsApi.update(id, data)
      const index = deployments.value.findIndex(d => d.id === id)
      if (index !== -1) {
        deployments.value[index] = response.data
      }
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新 deployment 失败'
      return null
    } finally {
      loading.value = false
    }
  }

  async function deleteDeployment(id: number) {
    loading.value = true
    error.value = null
    try {
      await deploymentsApi.delete(id)
      deployments.value = deployments.value.filter(d => d.id !== id)
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : '删除 deployment 失败'
      return false
    } finally {
      loading.value = false
    }
  }

  function getEnabledDeployments() {
    return deployments.value.filter(d => d.enabled)
  }

  function clear() {
    deployments.value = []
  }

  return {
    deployments,
    loading,
    error,
    loaded,
    listDeployments,
    getDeployment,
    createDeployment,
    updateDeployment,
    deleteDeployment,
    getEnabledDeployments,
    clear,
  }
})
