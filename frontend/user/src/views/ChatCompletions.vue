<template>
  <div class="max-w-4xl mx-auto">
    <div class="bg-white rounded-lg shadow p-6">
      <h2 class="text-xl font-semibold text-gray-900 mb-6">Chat Completions API</h2>

      <!-- Model Selector -->
      <div class="mb-6">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          模型 <span class="text-red-500">*</span>
        </label>
        <select
          v-model="form.model"
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
        >
          <option value="">请选择模型</option>
          <option v-for="model in models" :key="model.id" :value="model.id">
            {{ model.id }}
          </option>
        </select>
        <p v-if="errors.model" class="mt-1 text-xs text-red-600">{{ errors.model }}</p>
      </div>

      <!-- Messages -->
      <div class="mb-6">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          消息 <span class="text-red-500">*</span>
        </label>
        <div class="space-y-4">
          <div
            v-for="(msg, idx) in form.messages"
            :key="idx"
            class="flex items-start space-x-3 p-4 bg-gray-50 rounded-lg"
          >
            <select
              v-model="msg.role"
              class="mt-1 px-2 py-1 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="system">System</option>
              <option value="user">User</option>
              <option value="assistant">Assistant</option>
            </select>
            <textarea
              v-model="msg.content"
              rows="3"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              placeholder="输入消息内容..."
            ></textarea>
            <button
              v-if="form.messages.length > 1"
              @click="removeMessage(idx)"
              class="mt-1 p-2 text-gray-400 hover:text-red-600"
              title="删除消息"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
        <button
          @click="addMessage"
          class="mt-2 text-sm text-blue-600 hover:text-blue-800 font-medium"
        >
          + 添加消息
        </button>
        <p v-if="errors.messages" class="mt-1 text-xs text-red-600">{{ errors.messages }}</p>
      </div>

      <!-- Settings -->
      <div class="mb-6 grid grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            温度
            <span class="text-gray-400 text-xs ml-1">(0-2)</span>
          </label>
          <input
            v-model.number="form.temperature"
            type="number"
            step="0.1"
            min="0"
            max="2"
            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            Max Tokens
            <span class="text-gray-400 text-xs ml-1">(可选)</span>
          </label>
          <input
            v-model.number="form.maxTokens"
            type="number"
            min="1"
            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
            placeholder="默认不限制"
          />
        </div>
      </div>

      <!-- Streaming Toggle -->
      <div class="flex items-center mb-6">
        <input
          v-model="form.stream"
          type="checkbox"
          class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
        />
        <span class="ml-2 text-sm text-gray-700">启用流式响应 (stream)</span>
      </div>

      <!-- Submit -->
      <div class="flex items-center justify-end space-x-3">
        <button
          @click="resetForm"
          class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
        >
          重置
        </button>
        <button
          @click="handleSubmit"
          :disabled="loading || !form.model || !form.messages.length"
          class="px-6 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ loading ? '生成中...' : '发送请求' }}
        </button>
      </div>

      <!-- Response -->
      <div v-if="response" class="mt-6 p-4 bg-gray-50 rounded-lg border border-gray-200">
        <h3 class="text-sm font-medium text-gray-700 mb-2">响应:</h3>
        <div class="bg-white p-4 rounded border border-gray-200">
          <div v-if="responseError" class="text-red-600">
            <p class="font-medium">{{ responseError.type || 'Error' }}</p>
            <p class="text-sm mt-1">{{ responseError.message }}</p>
          </div>
          <div v-else-if="form.stream && streamedContent">
            <div class="prose prose-sm max-w-none">
              <p class="whitespace-pre-wrap">{{ streamedContent }}</p>
            </div>
            <div v-if="usage" class="mt-4 text-xs text-gray-500 border-t pt-2">
              <p>Usage: {{ usage.prompt_tokens }} prompt + {{ usage.completion_tokens }} completion = {{ usage.total_tokens }} total</p>
            </div>
          </div>
          <div v-else>
            <pre class="whitespace-pre-wrap text-sm bg-gray-100 p-3 rounded text-gray-800">{{ formattedResponse }}</pre>
            <div v-if="responseUsage" class="mt-2 text-xs text-gray-500">
              <p>Usage: {{ responseUsage.prompt_tokens }} prompt + {{ responseUsage.completion_tokens }} completion = {{ responseUsage.total_tokens }} total</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { chatApi, modelsApi } from '../api/client'
import type { ChatRequest, ChatResponse, ErrorResponse, ChatMessage } from '../api/types'

const models = ref<ChatResponse['choices']>([])
const loading = ref(false)
const response = ref<ChatResponse | null>(null)
const responseError = ref<ErrorResponse['error'] | null>(null)
const streamedContent = ref('')
const usage = ref<ChatResponse['usage'] | null>(null)
const responseUsage = ref<ChatResponse['usage'] | null>(null)

const form = reactive<ChatRequest>({
  model: '',
  messages: [{ role: 'user', content: '' }],
  temperature: 0.7,
  maxTokens: undefined,
  stream: false,
})

const errors = reactive<Record<string, string>>({})

async function loadModels() {
  try {
    const res = await modelsApi.list()
    models.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load models:', e)
  }
}

function addMessage() {
  form.messages.push({ role: 'user', content: '' })
}

function removeMessage(index: number) {
  if (form.messages.length > 1) {
    form.messages.splice(index, 1)
  }
}

function resetForm() {
  form.model = ''
  form.messages = [{ role: 'user', content: '' }]
  form.temperature = 0.7
  form.maxTokens = undefined
  form.stream = false
  response.value = null
  responseError.value = null
  streamedContent.value = ''
  usage.value = null
}

async function handleSubmit() {
  errors.model = !form.model ? '请选择模型' : ''
  const hasValidMessage = form.messages.some(m => m.content.trim())
  errors.messages = !hasValidMessage ? '至少需要一条有效消息' : ''

  if (Object.values(errors).some(e => e)) {
    return
  }

  loading.value = true
  response.value = null
  responseError.value = null
  streamedContent.value = ''

  try {
    const result = await chatApi.create(form)

    if (form.stream) {
      // Handle streaming response
      const reader = result.data.getReader()
      const decoder = new TextDecoder()
      streamedContent.value = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split('\n').filter(line => line.trim())

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6)
            if (data === '[DONE]') {
              continue
            }
            try {
              const parsed = JSON.parse(data)
              if (parsed.choices && parsed.choices[0].delta.content) {
                streamedContent.value += parsed.choices[0].delta.content
              }
              if (parsed.usage) {
                usage.value = parsed.usage
              }
            } catch (e) {
              // Ignore parse errors
            }
          }
        }
      }
    } else {
      // Handle non-streaming response
      response.value = result.data
      responseUsage.value = result.data.usage || null
    }
  } catch (err: any) {
    responseError.value = {
      message: err.message || '请求失败',
      type: err.type || 'api_error',
      code: err.code || 'internal_error',
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadModels()
})
</script>
