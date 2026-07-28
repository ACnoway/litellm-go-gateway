<template>
  <div class="max-w-4xl mx-auto">
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-900">
        {{ isEdit ? '编辑 Deployment' : '创建 Deployment' }}
      </h1>
      <button
        @click="goBack"
        class="inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
      >
        <svg class="-ml-1 mr-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
        </svg>
        返回
      </button>
    </div>

    <form @submit.prevent="handleSubmit" class="bg-white shadow rounded-lg p-6 space-y-6">
      <!-- Deployment Name -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">
          逻辑模型名 <span class="text-red-500">*</span>
        </label>
        <input
          v-model="form.name"
          type="text"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          placeholder="例如: gpt-4-turbo"
          :class="{ 'border-red-300': errors.name }"
        />
        <p v-if="errors.name" class="mt-1 text-xs text-red-600">{{ errors.name }}</p>
        <p class="mt-1 text-xs text-gray-500">
          用户请求时使用的模型名称（必须与请求中的 model 字段完全匹配）
        </p>
      </div>

      <!-- Actual Model Name -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">
          真实模型名 <span class="text-red-500">*</span>
        </label>
        <input
          v-model="form.actualModel"
          type="text"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          placeholder="例如: gpt-4-turbo-2024-04-09"
          :class="{ 'border-red-300': errors.actualModel }"
        />
        <p v-if="errors.actualModel" class="mt-1 text-xs text-red-600">{{ errors.actualModel }}</p>
        <p class="mt-1 text-xs text-gray-500">
          实际转发给上游 provider 的模型名称
        </p>
      </div>

      <!-- Providers -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">
          Providers <span class="text-red-500">*</span>
        </label>
        <div class="flex flex-wrap gap-2 mb-2">
          <span
            v-for="provider in availableProviders" :key="provider"
            :class="[
              'inline-flex items-center px-3 py-1.5 rounded-md text-sm cursor-pointer transition-colors',
              form.providers.includes(provider)
                ? 'bg-blue-100 text-blue-800 border border-blue-300'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            ]"
            @click="toggleProvider(provider)"
          >
            {{ provider }}
            <svg
              v-if="form.providers.includes(provider)"
              class="w-4 h-4 ml-1"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </span>
        </div>
        <p class="text-xs text-gray-500">
          选择可用的 provider，按优先级顺序排列
        </p>
        <p v-if="errors.providers" class="mt-1 text-xs text-red-600">{{ errors.providers }}</p>
      </div>

      <!-- Strategy -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">
          负载均衡策略
        </label>
        <div class="grid grid-cols-3 gap-3">
          <label
            v-for="strategy in strategies"
            :key="strategy"
            class="flex items-center p-3 border rounded-md cursor-pointer transition-colors hover:bg-gray-50"
            :class="form.strategy === strategy ? 'border-blue-500 bg-blue-50' : 'border-gray-300'"
          >
            <input
              v-model="form.strategy"
              type="radio"
              name="strategy"
              :value="strategy"
              class="text-blue-600 focus:ring-blue-500"
            />
            <span class="ml-2 text-sm text-gray-700">{{ strategy }}</span>
          </label>
        </div>
        <p class="mt-1 text-xs text-gray-500">
          priority: 按顺序尝试，失败则切换下一个<br>
          round-robin: 轮询选择 provider<br>
          weighted: 按权重随机选择
        </p>
      </div>

      <!-- Weights (only show for weighted strategy) -->
      <div v-if="form.strategy === 'weighted'" class="p-4 bg-gray-50 rounded-md border border-gray-200">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          权重配置
        </label>
        <div v-for="(provider, idx) in form.providers" :key="provider" class="flex items-center mb-2">
          <span class="w-24 text-sm text-gray-600">{{ provider }}</span>
          <input
            v-model="form.weights[idx]"
            type="number"
            min="1"
            class="w-20 px-2 py-1 border border-gray-300 rounded-md text-sm"
            @input="validateWeights"
          />
          <span class="ml-2 text-xs text-gray-500">weight</span>
        </div>
        <p v-if="errors.weights" class="mt-1 text-xs text-red-600">{{ errors.weights }}</p>
      </div>

      <!-- Max Tokens -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">
          最大 Token 限制
        </label>
        <input
          v-model="form.maxTokens"
          type="number"
          min="0"
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          placeholder="例如: 128000"
        />
        <p class="mt-1 text-xs text-gray-500">
          可选：设置模型的最大 token 限制
        </p>
      </div>

      <!-- Description -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">
          描述
        </label>
        <textarea
          v-model="form.description"
          rows="3"
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          placeholder="此 deployment 的描述..."
        ></textarea>
      </div>

      <!-- Enabled Toggle -->
      <div class="flex items-center">
        <input
          v-model="form.enabled"
          type="checkbox"
          class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
        />
        <span class="ml-2 text-sm text-gray-700">启用此 Deployment</span>
      </div>

      <!-- Submit -->
      <div class="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
        <button
          type="button"
          @click="goBack"
          class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
        >
          取消
        </button>
        <button
          type="submit"
          :disabled="loading || !form.name || !form.actualModel || !form.providers.length"
          class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ loading ? '保存中...' : (isEdit ? '保存更改' : '创建 Deployment') }}
        </button>
      </div>

      <!-- Error Alert -->
      <div v-if="saveError" class="bg-red-50 border border-red-200 rounded-md p-4">
        <div class="flex">
          <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
          <div class="ml-3">
            <p class="text-sm text-red-700">{{ saveError }}</p>
          </div>
        </div>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useDeploymentsStore } from '../stores/useDeployments'
import { useProvidersStore } from '../stores/useProviders'

const router = useRouter()
const route = useRoute()
const deploymentsStore = useDeploymentsStore()
const providersStore = useProvidersStore()

const isEdit = route.name === 'admin-deployments-edit' as boolean
const deploymentId = isEdit ? Number(route.params.id) : null

const availableProviders = providersStore.providers
  .filter(p => p.available)
  .map(p => p.name)

const strategies = ['priority', 'round-robin', 'weighted']

const form = reactive<DeploymentRequest>({
  name: '',
  actualModel: '',
  providers: [],
  strategy: 'priority',
  weights: [],
  maxTokens: 0,
  description: '',
  enabled: true,
})

const loading = ref(false)
const saveError = ref<string | null>(null)
const errors = reactive<Record<string, string>>({})

function toggleProvider(provider: string) {
  const idx = form.providers.indexOf(provider)
  if (idx > -1) {
    form.providers.splice(idx, 1)
  } else {
    form.providers.push(provider)
  }
}

function validateWeights() {
  if (form.weights && form.weights.some(w => w <= 0)) {
    errors.weights = '权重必须为正整数'
  } else {
    delete errors.weights
  }
}

async function handleSubmit() {
  errors.name = !form.name ? '逻辑模型名不能为空' : ''
  errors.actualModel = !form.actualModel ? '真实模型名不能为空' : ''
  errors.providers = !form.providers.length ? '至少选择一个 provider' : ''

  if (form.strategy === 'weighted' && form.weights) {
    if (form.weights.length !== form.providers.length) {
      errors.weights = '权重数量必须与 provider 数量一致'
    } else if (form.weights.some(w => w <= 0)) {
      errors.weights = '权重必须为正整数'
    }
  }

  if (Object.values(errors).some(e => e)) {
    return
  }

  loading.value = true
  saveError.value = null

  try {
    if (isEdit && deploymentId) {
      await deploymentsStore.updateDeployment(deploymentId, form)
      router.push('/admin/deployments')
    } else {
      await deploymentsStore.createDeployment(form)
      router.push('/admin/deployments')
    }
  } catch (err) {
    saveError.value = err instanceof Error ? err.message : '保存失败'
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push('/admin/deployments')
}

onMounted(async () => {
  if (isEdit && deploymentId) {
    const deployment = await deploymentsStore.getDeployment(deploymentId)
    if (deployment) {
      form.name = deployment.name
      form.actualModel = deployment.actual_model
      form.providers = [...deployment.providers]
      form.strategy = deployment.strategy
      form.weights = deployment.weights || []
      form.maxTokens = deployment.max_tokens || 0
      form.description = deployment.description || ''
      form.enabled = deployment.enabled
    }
  }
  await providersStore.listProviders()
})
</script>
