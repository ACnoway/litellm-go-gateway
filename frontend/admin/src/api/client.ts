import axios from 'axios'
import type { ErrorResponse } from './types'

const api = axios.create({
  baseURL: '', // 同源，使用相对路径
  timeout: 30000,
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
      const status = error.response.status
      const errorMessage = error.response.data?.error?.message || '请求失败'

      if (status === 401) {
        localStorage.removeItem('GATEWAY_API_KEY')
      }

      const err = new Error(errorMessage) as Error & { status?: number; code?: string }
      err.status = status
      err.code = error.response.data?.error?.code
      return Promise.reject(err)
    } else if (error.request) {
      const err = new Error('网络错误：未收到服务器响应') as Error & { code?: string }
      err.code = 'NO_RESPONSE'
      return Promise.reject(err)
    }
    return Promise.reject(error.message)
  }
)

// Provider API
export const providersApi = {
  list: () => api.get('/admin/providers'),
  get: (name: string) => api.get(`/admin/providers/${name}`),
}

// Deployment API
export const deploymentsApi = {
  list: () => api.get('/admin/deployments'),
  get: (id: number) => api.get(`/admin/deployments/${id}`),
  create: (data: any) => api.post('/admin/deployments', data),
  update: (id: number, data: any) => api.put(`/admin/deployments/${id}`, data),
  delete: (id: number) => api.delete(`/admin/deployments/${id}`),
}

// Model API - uses /v1/models
export const modelsApi = {
  list: () => api.get('/v1/models'),
}

// Routing Rule API
export const routingRulesApi = {
  list: () => api.get('/admin/routing/rules'),
  get: (id: number) => api.get(`/admin/routing/rules/${id}`),
  create: (data: any) => api.post('/admin/routing/rules', data),
  update: (id: number, data: any) => api.put(`/admin/routing/rules/${id}`, data),
  delete: (id: number) => api.delete(`/admin/routing/rules/${id}`),
}

export default api
