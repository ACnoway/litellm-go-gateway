// Admin API 类型定义
export interface ProviderInfo {
  name: string
  available: boolean
  type: string
  base_url?: string
}

export interface Deployment {
  id: number
  name: string
  actual_model: string
  providers: string[]
  strategy: 'priority' | 'round-robin' | 'weighted'
  weights?: number[]
  max_tokens?: number
  description?: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface DeploymentRequest {
  name: string
  actual_model: string
  providers: string[]
  strategy?: 'priority' | 'round-robin' | 'weighted'
  weights?: number[]
  max_tokens?: number
  description?: string
  enabled?: boolean
}

export interface ModelInfo {
  id: string
  object: string
  created: number
  owned_by: string
  ready: boolean
  description?: string
  providers: string[]
}

export interface RoutingRule {
  id: number
  pattern: string
  providers: string[]
  created_at: string
  updated_at: string
}

export interface RoutingRuleRequest {
  pattern: string
  providers: string[]
}

export interface ErrorResponse {
  error: {
    message: string
    type: string
    code: string
  }
}

export interface PaginatedResponse<T> {
  object: string
  data: T[]
}
