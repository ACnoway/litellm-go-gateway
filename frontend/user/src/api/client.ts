import axios from 'axios'
import type { ErrorResponse } from './types'

const api = axios.create({
  baseURL: '/v1',
  timeout: 60000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const apiKey = localStorage.getItem('GATEWAY_API_KEY')
    if (apiKey && config.headers) {
      config.headers.Authorization = `Bearer ${apiKey}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      const err = new Error(error.response.data?.error?.message || '请求失败') as Error & {
        status?: number
        code?: string
        type?: string
      }
      err.status = error.response.status
      err.code = error.response.data?.error?.code
      err.type = error.response.data?.error?.type
      return Promise.reject(err)
    } else if (error.request) {
      const err = new Error('网络错误：未收到服务器响应') as Error & { code?: string }
      err.code = 'NO_RESPONSE'
      return Promise.reject(err)
    }
    return Promise.reject(error.message)
  }
)

// Chat Completions API
export const chatApi = {
  create: (data: any) => api.post('/chat/completions', data, { timeout: 30000 }),
}

// Models API
export const modelsApi = {
  list: () => api.get('/models'),
  get: (id: string) => api.get(`/models/${id}`),
}

export default api
