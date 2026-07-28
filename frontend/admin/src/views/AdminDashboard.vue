<template>
  <div class="max-w-6xl mx-auto">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">仪表盘</h1>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
      <div class="bg-white rounded-lg shadow p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-medium text-gray-500">可用 Models</h3>
          <div class="p-2 bg-blue-100 rounded-lg">
            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 7.25v11.25m0 0a2.25 2.25 0 01-2.25-2.25M21 7.25a2.25 2.25 0 00-2.25-2.25h-15a2.25 2.25 0 00-2.25 2.25v11.25a2.25 2.25 0 002.25 2.25h15a2.25 2.25 0 002.25-2.25" />
            </svg>
          </div>
        </div>
        <div class="text-3xl font-bold text-gray-900">{{ stats.modelsCount }}</div>
        <div class="mt-1 text-sm text-gray-500">{{ stats.readyModels }} 个就绪</div>
      </div>

      <div class="bg-white rounded-lg shadow p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-medium text-gray-500">Deployments</h3>
          <div class="p-2 bg-green-100 rounded-lg">
            <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
        </div>
        <div class="text-3xl font-bold text-gray-900">{{ stats.deploymentsCount }}</div>
        <div class="mt-1 text-sm text-gray-500">{{ stats.enabledDeployments }} 个启用</div>
      </div>

      <div class="bg-white rounded-lg shadow p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-medium text-gray-500">Providers</h3>
          <div class="p-2 bg-purple-100 rounded-lg">
            <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </div>
        </div>
        <div class="text-3xl font-bold text-gray-900">{{ stats.providersCount }}</div>
        <div class="mt-1 text-sm text-gray-500">{{ stats.availableProviders }} 个可用</div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="bg-white rounded-lg shadow p-6 mb-8">
      <h2 class="text-lg font-semibold text-gray-900 mb-4">快捷操作</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <button
          @click="createDeployment"
          class="flex flex-col items-center p-4 border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 transition-all"
        >
          <div class="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-3">
            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
          </div>
          <h4 class="text-sm font-medium text-gray-900 mb-1">创建 Deployment</h4>
          <p class="text-xs text-gray-500 text-center">为模型配置上游 provider</p>
        </button>

        <button
          @click="viewModels"
          class="flex flex-col items-center p-4 border border-gray-200 rounded-lg hover:border-green-300 hover:bg-green-50 transition-all"
        >
          <div class="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-3">
            <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 7.25v11.25m0 0a2.25 2.25 0 01-2.25-2.25M21 7.25a2.25 2.25 0 00-2.25-2.25h-15a2.25 2.25 0 00-2.25 2.25v11.25a2.25 2.25 0 002.25 2.25h15a2.25 2.25 0 002.25-2.25" />
            </svg>
          </div>
          <h4 class="text-sm font-medium text-gray-900 mb-1">查看 Models</h4>
          <p class="text-xs text-gray-500 text-center">浏览所有可用的逻辑模型</p>
        </button>

        <button
          @click="viewProviders"
          class="flex flex-col items-center p-4 border border-gray-200 rounded-lg hover:border-purple-300 hover:bg-purple-50 transition-all"
        >
          <div class="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-3">
            <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </div>
          <h4 class="text-sm font-medium text-gray-900 mb-1">配置 Provider</h4>
          <p class="text-xs text-gray-500 text-center">管理上游 LLM provider</p>
        </button>
      </div>
    </div>

    <!-- Recent Activity -->
    <div class="bg-white rounded-lg shadow">
      <h2 class="text-lg font-semibold text-gray-900 p-6 border-b border-gray-200">最近活动</h2>
      <div class="p-6">
        <div class="text-center py-8 text-gray-500">
          <svg class="w-12 h-12 mx-auto mb-3 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p>暂无活动记录</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDeploymentsStore } from '../stores/useDeployments'
import { useProvidersStore } from '../stores/useProviders'
import { useModelsStore } from '../stores/useModels'

const router = useRouter()
const deploymentsStore = useDeploymentsStore()
const providersStore = useProvidersStore()
const modelsStore = useModelsStore()

const stats = ref({
  modelsCount: 0,
  readyModels: 0,
  deploymentsCount: 0,
  enabledDeployments: 0,
  providersCount: 0,
  availableProviders: 0,
})

function createDeployment() {
  router.push('/admin/deployments/new')
}

function viewModels() {
  router.push('/admin/models')
}

function viewProviders() {
  router.push('/admin/providers')
}

onMounted(async () => {
  await Promise.all([
    deploymentsStore.listDeployments(),
    providersStore.listProviders(),
    modelsStore.listModels(),
  ])

  stats.value = {
    modelsCount: modelsStore.models.length,
    readyModels: modelsStore.getReadyModels().length,
    deploymentsCount: deploymentsStore.deployments.length,
    enabledDeployments: deploymentsStore.getEnabledDeployments().length,
    providersCount: providersStore.providers.length,
    availableProviders: providersStore.getAvailableProviders().length,
  }
})
</script>
