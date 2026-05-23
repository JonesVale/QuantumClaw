import apiClient, { ApiResponse } from './api'
export type { ApiResponse }

// Helper to extract parameters safely (handles TanStack Query context object)
function extractParams<T>(arg1: unknown, arg2?: unknown): T | undefined {
  // Check if arg1 is TanStack Query context object (has queryKey or signal)
  if (arg1 && typeof arg1 === 'object') {
    const obj = arg1 as Record<string, unknown>
    if ('queryKey' in obj || 'signal' in obj) {
      return arg2 as T
    }
  }
  return arg1 as T
}

// ---------------------------------------------------------------------------
// Authentication API
// ---------------------------------------------------------------------------

export async function signIn(username: string, password: string): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/login', { username, password })
  return res.data
}

export async function signOut(): Promise<ApiResponse> {
  const res = await apiClient.get('/api/user/logout')
  return res.data
}

export async function register(data: {
  username: string
  password: string
  email?: string
  display_name?: string
  aff_code?: string
}): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/register', data)
  return res.data
}

// ---------------------------------------------------------------------------
// Channel Management API
// ---------------------------------------------------------------------------

export interface Channel {
  id: number
  name: string
  group: string
  type: number
  key: string
  base_url: string
  models: string
  status: number
  weight: number
  cost_per_unit: number
  sell_price_rate: number
  used_quota: number
  created_time: number
  category: string
  config?: string
}

export interface ChannelTestResult {
  status: 'success' | 'error'
  latency_ms?: number
  error?: string
  models?: string[]
}

export async function getChannels(
  arg1?: unknown,
  arg2?: {
    group?: string
    status?: number
    type?: string
    search?: string
  }
): Promise<ApiResponse<Channel[]>> {
  const params: Record<string, unknown> = extractParams<typeof arg2>(arg1, arg2) || {}
  // Fetch all channels (no pagination)
  params.scope = 'all'
  const res = await apiClient.get('/api/channel', { params })
  return res.data
}

export async function getChannel(id: number): Promise<ApiResponse<Channel>> {
  const res = await apiClient.get(`/api/channel/${id}`)
  return res.data
}

export async function createChannel(data: ChannelFormData): Promise<ApiResponse> {
  const payload: Record<string, unknown> = { ...data }
  // Serialize config fields into config JSON
  const configParts: Record<string, unknown> = {}
  if (data.cache_billing_ratio !== undefined) configParts.cache_billing_ratio = data.cache_billing_ratio
  if (data.thinking_to_content !== undefined) configParts.thinking_to_content = data.thinking_to_content
  if (Object.keys(configParts).length > 0) {
    payload.config = JSON.stringify(configParts)
  }
  delete payload.cache_billing_ratio
  delete payload.thinking_to_content
  const res = await apiClient.post('/api/channel', payload)
  return res.data
}

export async function updateChannel(data: ChannelFormData & { id: number }): Promise<ApiResponse> {
  const payload: Record<string, unknown> = { ...data }
  const configParts: Record<string, unknown> = {}
  if (data.cache_billing_ratio !== undefined) configParts.cache_billing_ratio = data.cache_billing_ratio
  if (data.thinking_to_content !== undefined) configParts.thinking_to_content = data.thinking_to_content
  if (Object.keys(configParts).length > 0) {
    payload.config = JSON.stringify(configParts)
  }
  delete payload.cache_billing_ratio
  delete payload.thinking_to_content
  const res = await apiClient.put('/api/channel', payload)
  return res.data
}

export async function deleteChannel(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/channel/${id}`)
  return res.data
}

export async function testChannel(id: number): Promise<ApiResponse<ChannelTestResult>> {
  const res = await apiClient.get(`/api/channel/test/${id}`)
  return res.data
}

export async function testAllChannels(): Promise<ApiResponse<Record<number, ChannelTestResult>>> {
  const res = await apiClient.get('/api/channel/test')
  return res.data
}

export async function updateChannelBalance(id: number): Promise<ApiResponse> {
  const res = await apiClient.get(`/api/channel/update_balance/${id}`)
  return res.data
}

export async function updateAllChannelsBalance(): Promise<ApiResponse> {
  const res = await apiClient.get('/api/channel/update_balance')
  return res.data
}

// 获取所有渠道类型名称映射（动态，避免前端硬编码）
export async function getChannelTypes(): Promise<Record<string, string>> {
  const res = await apiClient.get('/api/channel/types')
  return res.data
}

export async function getChannelModels(): Promise<ApiResponse<string[]>> {
  const res = await apiClient.get('/api/channel/models')
  return res.data
}

// ---------------------------------------------------------------------------
// Token Management API
// ---------------------------------------------------------------------------

export interface Token {
  id: number
  name: string
  key: string
  status: number
  created_time: number
  remain_quota?: number
  remaining_quota?: number
  unlimited_quota?: boolean
  used_quota?: number
  request_count?: number
  models?: string[]
  group?: string
}

export interface TokenFormData {
  name?: string
  remain_quota?: number
  expired_time?: number
  unlimited_quota?: boolean
  models?: string
  subnet?: string
  group?: string
}

export async function getTokens(
  arg1?: unknown,
  arg2?: {
    status?: number
    group?: string
    search?: string
  }
): Promise<ApiResponse<Token[]>> {
  const params = extractParams<typeof arg2>(arg1, arg2)
  const res = await apiClient.get('/api/token', { params })
  return res.data
}

export async function getToken(id: number): Promise<ApiResponse<Token>> {
  const res = await apiClient.get(`/api/token/${id}`)
  return res.data
}

export async function createToken(data: Partial<Token>): Promise<ApiResponse> {
  const res = await apiClient.post('/api/token', data)
  return res.data
}

export async function updateToken(data: Partial<Token>): Promise<ApiResponse> {
  const res = await apiClient.put('/api/token', data)
  return res.data
}

export async function deleteToken(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/token/${id}`)
  return res.data
}

export async function manageToken(id: number, action: string): Promise<ApiResponse> {
  const res = await apiClient.post('/api/token/manage', { id, action })
  return res.data
}

// ---------------------------------------------------------------------------
// Log Management API
// ---------------------------------------------------------------------------

export interface Log {
  id: number
  user_id: number
  username?: string
  channel_id: number
  channel_name?: string
  type: string
  type_?: number
  model_name?: string
  model?: string
  prompt_tokens?: number
  completion_tokens?: number
  request_tokens?: number
  response_tokens?: number
  quota?: number
  created_at?: number
  created_time?: number
}

export type LogEntry = Log

export interface LogStat {
  today_requests: number
  today_cost: number
  total_requests: number
  total_cost: number
}

export async function getLogs(
  arg1?: unknown,
  arg2?: {
    channel_id?: number
    model?: string
    start_date?: string
    end_date?: string
    page?: number
    page_size?: number
  }
): Promise<ApiResponse<Log[]>> {
  const params = extractParams<typeof arg2>(arg1, arg2)
  const res = await apiClient.get('/api/log/self', { params })
  return res.data
}

export async function getLogsStat(): Promise<ApiResponse<LogStat>> {
  const res = await apiClient.get('/api/log/self/stat')
  return res.data
}

export async function getAllLogs(
  arg1?: unknown,
  arg2?: {
    user_id?: number
    channel_id?: number
    model?: string
    start_date?: string
    end_date?: string
    page?: number
    page_size?: number
  }
): Promise<ApiResponse<Log[]>> {
  const params = extractParams<typeof arg2>(arg1, arg2)
  const res = await apiClient.get('/api/log', { params })
  return res.data
}

export async function getAllLogsStat(): Promise<ApiResponse<LogStat>> {
  const res = await apiClient.get('/api/log/stat')
  return res.data
}

export async function searchLogs(
  arg1: unknown,
  arg2?: {
    user_id?: number
    channel_id?: number
    model?: string
    start_date?: string
    end_date?: string
    search?: string
    page?: number
    page_size?: number
  }
): Promise<ApiResponse<Log[]>> {
  const params = extractParams<typeof arg2>(arg1, arg2)
  const res = await apiClient.get('/api/log/self/search', { params })
  return res.data
}

export async function deleteHistoryLogs(before_date: string): Promise<ApiResponse> {
  const res = await apiClient.delete('/api/log', {
    data: { before_date },
    skipBusinessError: true,
  } as unknown as Record<string, unknown>)
  return res.data
}

// ---------------------------------------------------------------------------
// Model Management API
// ---------------------------------------------------------------------------

export interface Model {
  id: number
  name: string
  type: string
  status: number
  channel_id: number
  channel_name: string
  created_at: number
}

export type ModelInfo = Model

// Enhanced model with full pricing and provider info (new /api/models endpoint)
export interface EnhancedModel {
  name: string
  channel_id: number
  channel_name: string
  provider: string
  provider_type: number
  cost_per_unit: number
  sell_price_rate: number
  input_price: number
  output_price: number
  status: number
  group: string
}

// Model usage ranking (GET /api/models/rankings)
export interface ModelRanking {
  model: string
  provider: string
  channel_name: string
  tokens_7d: number
  trend_percent: number
  avg_speed_ms: number
  price_per_1k: number
  request_count_7d: number
}

export async function getModels(
  arg1?: unknown,
  arg2?: {
    type?: string
    status?: number
    search?: string
  }
): Promise<ApiResponse<ModelInfo[]>> {
  const params = extractParams<typeof arg2>(arg1, arg2)
  const res = await apiClient.get('/api/models', { params })
  // Backend returns { success: true, data: { "1": ["gpt-4", ...] } }
  // Transform dict format to ModelInfo[] array
  const raw = res.data || {}
  // Check if there's a nested data dict of channel_id -> string[]
  let dict: Record<string, any> | null = null
  if (raw.data && typeof raw.data === 'object' && !Array.isArray(raw.data)) {
    dict = raw.data as Record<string, any>
  } else if (typeof raw === 'object' && !Array.isArray(raw) && !('success' in raw)) {
    dict = raw as Record<string, any>
  }
  if (dict) {
    const entries = Object.entries(dict)
    if (entries.length > 0) {
      const firstVal = entries[0][1]
      if (Array.isArray(firstVal)) {
        const result: ModelInfo[] = []
        let modelId = 1
        for (const [channelId, modelNames] of entries) {
          if (Array.isArray(modelNames)) {
            for (const name of modelNames as string[]) {
              result.push({
                id: modelId++,
                name,
                type: '',
                status: 1,
                channel_id: parseInt(channelId, 10) || 0,
                channel_name: 'Channel ' + channelId,
                created_at: Date.now(),
              })
            }
          }
        }
        return { success: true, message: '', data: result }
      }
    }
  }
  // Direct array format
  if (Array.isArray(raw)) {
    return { success: true, message: '', data: raw as any[] }
  }
  if (Array.isArray(raw.data)) {
    return { success: true, message: '', data: raw.data as any[] }
  }
  // Fallback
  return raw as any
}

// Fetch enhanced models with full pricing info
export async function getEnhancedModels(): Promise<ApiResponse<EnhancedModel[]>> {
  const res = await apiClient.get('/api/models')
  return res.data
}

// Fetch model usage rankings
export async function getModelRankings(): Promise<ApiResponse<ModelRanking[]>> {
  const res = await apiClient.get('/api/models/rankings')
  return res.data
}

// ---------------------------------------------------------------------------
// User Management API
// ---------------------------------------------------------------------------

export interface User {
  id: number
  username: string
  email: string
  display_name: string
  role: number
  status: number
  balance: number
  used_quota: number
  quota?: number
  request_count?: number
  created_time?: number
  created_at?: number
  updated_at?: number
}

export interface UserFormData {
  username?: string
  password?: string
  display_name?: string
  email?: string
  role?: number
  group?: string
  quota?: number
}

export interface ChannelFormData {
  type?: number
  key?: string
  name?: string
  base_url?: string
  topup_url?: string
  models?: string
  group?: string
  model_mapping?: string
  priority?: number
  weight?: number
  cache_billing_ratio?: number
  thinking_to_content?: boolean
  cost_per_unit?: number
  sell_price_rate?: number
  category?: string
}

export async function getUsers(
  arg1?: unknown,
  arg2?: {
    role?: number
    status?: number
    search?: string
    page?: number
    page_size?: number
  }
): Promise<ApiResponse<User[]>> {
  const params = extractParams<typeof arg2>(arg1, arg2)
  const res = await apiClient.get('/api/user', { params })
  return res.data
}

export async function getUser(id: number): Promise<ApiResponse<User>> {
  const res = await apiClient.get(`/api/user/${id}`)
  return res.data
}

export async function createUser(data: Partial<User>): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user', data)
  return res.data
}

export async function updateUser(data: Partial<User>): Promise<ApiResponse> {
  const res = await apiClient.put('/api/user', data)
  return res.data
}

export async function deleteUser(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/user/${id}`)
  return res.data
}

export async function manageUser(data: {
  id: number
  action: 'disable' | 'enable' | 'reset_quota' | 'reset_used_quota'
}): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/manage', data)
  return res.data
}

// ---------------------------------------------------------------------------
// Redemption API
// ---------------------------------------------------------------------------

export interface Redemption {
  id: number
  code: string
  quota: number
  status: number
  used_count: number
  max_count: number
  created_at: number
  expires_at: number
}

export async function getRedemptions(
  arg1?: unknown,
  arg2?: {
    status?: number
    search?: string
  }
): Promise<ApiResponse<Redemption[]>> {
  const params = extractParams<typeof arg2>(arg1, arg2)
  const res = await apiClient.get('/api/redemption', { params })
  return res.data
}

export async function getRedemptionCodes(): Promise<ApiResponse<Redemption[]>> {
  const res = await apiClient.get('/api/redemption', { params: { status: 1 } })
  return res.data
}

export async function createRedemption(data: Partial<Redemption>): Promise<ApiResponse> {
  const res = await apiClient.post('/api/redemption', data)
  return res.data
}

export async function createRedemptionCode(data: Partial<Redemption>): Promise<ApiResponse> {
  const res = await apiClient.post('/api/redemption', data)
  return res.data
}

export async function updateRedemption(data: Partial<Redemption>): Promise<ApiResponse> {
  const res = await apiClient.put('/api/redemption', data)
  return res.data
}

export async function deleteRedemption(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/redemption/${id}`)
  return res.data
}

export async function deleteRedemptionCode(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/redemption/${id}`)
  return res.data
}

// ---------------------------------------------------------------------------
// Group Management API
// ---------------------------------------------------------------------------

export interface Group {
  id: number
  name: string
  priority: number
  created_at: number
}

export async function getGroups(): Promise<ApiResponse<Group[]>> {
  const res = await apiClient.get('/api/group')
  return res.data
}

// ---------------------------------------------------------------------------
// Settings API
// ---------------------------------------------------------------------------

export interface Option {
  key: string
  value: string
  type: string
  description: string
}

export type SystemOption = Option

export async function getOptions(): Promise<ApiResponse<Option[]>> {
  const res = await apiClient.get('/api/option')
  return res.data
}

export async function updateOption(key: string, value: string): Promise<ApiResponse> {
  const res = await apiClient.put('/api/option', { key, value })
  return res.data
}

export async function updateOptions(updates: Record<string, string>): Promise<ApiResponse> {
  const promises = Object.entries(updates).map(([key, value]) =>
    apiClient.put('/api/option', { key, value })
  )
  await Promise.all(promises)
  return { success: true }
}

// ---------------------------------------------------------------------------
// Dashboard API
// ---------------------------------------------------------------------------

export interface DashboardStats {
  total_users: number
  total_requests: number
  total_cost: number
  total_channels: number
  active_channels: number
  total_tokens: number
  total_quota: number
  today_requests: number
  today_quota: number
}

export async function getDashboardStats(): Promise<ApiResponse<DashboardStats>> {
  const res = await apiClient.get('/api/user/self/dashboard')
  return res.data
}

// ---------------------------------------------------------------------------
// Enhanced Dashboard API (with charts)
// ---------------------------------------------------------------------------

export interface DailyRequest {
  date: string
  count: number
  cost: number
}

export interface ModelBreakdown {
  model: string
  percentage: number
  requests: number
}

export interface ProviderBreakdown {
  provider: string
  requests: number
  percentage: number
}

export interface EnhancedDashboard {
  total_requests: number
  today_requests: number
  total_cost: number
  model_count: number
  user_count?: number
  total_tokens?: number
  model_usage: { model_name: string; request_count: number; total_tokens?: number }[]
  daily_requests: DailyRequest[]
  model_breakdown: ModelBreakdown[]
  provider_breakdown: ProviderBreakdown[]
}

export async function getEnhancedDashboard(): Promise<ApiResponse<EnhancedDashboard>> {
  const res = await apiClient.get('/api/user/self/dashboard')
  return res.data
}

// ---------------------------------------------------------------------------
// Personal Profile API
// ---------------------------------------------------------------------------

export async function updateSelf(data: {
  display_name?: string
  email?: string
  password?: string
  old_password?: string
}): Promise<ApiResponse> {
  const res = await apiClient.put('/api/user/self', data)
  return res.data
}

export function generateApiKey(): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  const prefix = 'qc-'
  let result = prefix
  for (let i = 0; i < 48; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

// ---------------------------------------------------------------------------
// Settlement Config API
// ---------------------------------------------------------------------------

export interface SettlementConfigItem {
  id: number
  model_name: string
  unified_cost: number
  commission_rate: number
  platform_fee_rate: number
  enabled: number
  created_time: number
  updated_time: number
}

export async function getSettlementConfigs(): Promise<ApiResponse<SettlementConfigItem[]>> {
  const res = await apiClient.get('/api/settlement/config')
  return res.data
}

export async function createSettlementConfig(data: {
  model_name: string
  unified_cost?: number
  commission_rate?: number
  platform_fee_rate?: number
}): Promise<ApiResponse<SettlementConfigItem>> {
  const res = await apiClient.post('/api/settlement/config', data)
  return res.data
}

export async function updateSettlementConfig(id: number, data: Partial<SettlementConfigItem>): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/settlement/config/${id}`, data)
  return res.data
}

export async function deleteSettlementConfig(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/settlement/config/${id}`)
  return res.data
}

// ---------------------------------------------------------------------------
// Transaction API
// ---------------------------------------------------------------------------

export interface TransactionItem {
  id: number
  user_id: number
  model_name: string
  channel_id: number
  channel_owner_id: number
  promoter_id: number
  is_fallback: number
  unit_price: number
  total_amount: number
  unified_cost: number
  commission_amount: number
  platform_fee: number
  created_time: number
}

export interface TransactionsResponse {
  transactions: TransactionItem[]
  total: number
  page: number
  page_size: number
}

export async function getTransactions(params?: {
  user_id?: number
  promoter_id?: number
  channel_owner_id?: number
  model?: string
  page?: number
  page_size?: number
}): Promise<ApiResponse<TransactionsResponse>> {
  const res = await apiClient.get('/api/transactions', { params })
  return res.data
}

export interface TelegramWidgetInfo {
  bot_username: string
  callback_url: string
}

export async function getTelegramWidgetInfo(): Promise<ApiResponse<{ data?: TelegramWidgetInfo }>> {
  const res = await apiClient.get('/api/oauth/telegram/widget')
  return res.data
}
