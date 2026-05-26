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
  // Don't force scope — let caller determine. Backend defaults to 'limited' (key omitted).
  // Admin pages should pass { scope: 'all' } for full channel data.
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

// ---------------------------------------------------------------------------
// Balance & Payment / Topup API
// ---------------------------------------------------------------------------

export interface BalanceInfo {
  cash_balance: number
  quota: number
  total_used: number
}

export async function getBalance(): Promise<ApiResponse<BalanceInfo>> {
  const res = await apiClient.get('/api/user/self/balance')
  return res.data
}

export async function getTopUpInfo(): Promise<ApiResponse<{ min_topup: number; payment_methods: string[] }>> {
  const res = await apiClient.get('/api/user/self/topup/info')
  return res.data
}

export async function getTopUpList(): Promise<ApiResponse<any[]>> {
  const res = await apiClient.get('/api/user/self/topup/list')
  return res.data
}

export async function requestStripeTopUp(amount: number): Promise<ApiResponse<{ url: string }>> {
  const res = await apiClient.post('/api/user/self/topup/stripe', { amount })
  return res.data
}

export async function requestEpayTopUp(amount: number): Promise<ApiResponse<{ url: string }>> {
  const res = await apiClient.post('/api/user/self/topup/epay', { amount })
  return res.data
}

export async function requestCreemTopUp(amount: number): Promise<ApiResponse<{ url: string }>> {
  const res = await apiClient.post('/api/user/self/topup/creem', { amount })
  return res.data
}

export async function requestWaffoTopUp(amount: number): Promise<ApiResponse<{ url: string }>> {
  const res = await apiClient.post('/api/user/self/topup/waffo', { amount })
  return res.data
}

export async function requestBinanceTopUp(amount: number): Promise<ApiResponse<{ url: string }>> {
  const res = await apiClient.post('/api/user/self/topup/binance', { amount })
  return res.data
}

export async function adminTopUp(userId: number, amount: number, remark?: string): Promise<ApiResponse> {
  const res = await apiClient.post('/api/topup', { user_id: userId, amount, remark })
  return res.data
}

// ---------------------------------------------------------------------------
// Billing API (user self billing stats & records)
// ---------------------------------------------------------------------------

export interface BillingStats {
  total_quota: number
  used_quota: number
  remain_quota: number
  request_count: number
  display_in_currency: boolean
  quota_per_unit: number
}

export interface BillingRecord {
  id: number
  amount: number
  action: string
  before_quota: number
  after_quota: number
  status: string
  remark: string
  created_at: string
}

export async function getBillingStats(): Promise<ApiResponse<BillingStats>> {
  const res = await apiClient.get('/api/user/self/billing/stats')
  return res.data
}

export async function getBillingRecords(): Promise<ApiResponse<BillingRecord[]>> {
  const res = await apiClient.get('/api/user/self/billing/records')
  return res.data
}

// ---------------------------------------------------------------------------
// Checkin API
// ---------------------------------------------------------------------------

export interface CheckinStatus {
  checked_in: boolean
  streak_days: number
  total_checkins: number
  today_reward: number
}

export interface CheckinRecord {
  id: number
  reward: number
  created_at: string
}

export async function getCheckinStatus(): Promise<ApiResponse<CheckinStatus>> {
  const res = await apiClient.get('/api/user/self/checkin')
  return res.data
}

export async function getCheckinHistory(): Promise<ApiResponse<CheckinRecord[]>> {
  const res = await apiClient.get('/api/user/self/checkin/history')
  return res.data
}

export async function doCheckin(): Promise<ApiResponse<{ reward: number }>> {
  const res = await apiClient.post('/api/user/self/checkin')
  return res.data
}

// ---------------------------------------------------------------------------
// Withdrawal API
// ---------------------------------------------------------------------------

export interface WithdrawalRecord {
  id: number
  amount: number
  status: string
  remark: string
  created_at: string
}

export async function submitWithdrawal(amount: number): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/self/withdraw', { amount })
  return res.data
}

export async function getMyWithdrawals(): Promise<ApiResponse<WithdrawalRecord[]>> {
  const res = await apiClient.get('/api/user/self/withdraw/list')
  return res.data
}

export async function getMyWithdrawable(): Promise<ApiResponse<number>> {
  const res = await apiClient.get('/api/user/self/withdraw/available')
  return res.data
}

export async function getMyEarningsByChannel(): Promise<ApiResponse<any[]>> {
  const res = await apiClient.get('/api/user/self/withdraw/earnings')
  return res.data
}

// Admin withdrawal management
export async function adminGetWithdrawals(): Promise<ApiResponse<WithdrawalRecord[]>> {
  const res = await apiClient.get('/api/user/withdrawals')
  return res.data
}

export async function adminApproveWithdrawal(id: number): Promise<ApiResponse> {
  const res = await apiClient.post(`/api/user/withdrawals/${id}/approve`)
  return res.data
}

export async function adminRejectWithdrawal(id: number): Promise<ApiResponse> {
  const res = await apiClient.post(`/api/user/withdrawals/${id}/reject`)
  return res.data
}

export async function adminCompleteWithdrawal(id: number): Promise<ApiResponse> {
  const res = await apiClient.post(`/api/user/withdrawals/${id}/complete`)
  return res.data
}

// ---------------------------------------------------------------------------
// Subscription API
// ---------------------------------------------------------------------------

export interface SubscriptionPlan {
  id: number
  name: string
  description: string
  price: number
  quota: number
  duration_days: number
  models: string
  status: number
  created_time: number
}

export interface UserSubscription {
  id: number
  plan_id: number
  plan_name: string
  status: string
  start_time: number
  end_time: number
  remain_quota: number
  total_quota: number
}

export async function getSubscriptionPlans(): Promise<ApiResponse<SubscriptionPlan[]>> {
  const res = await apiClient.get('/api/user/self/subscription/plans')
  return res.data
}

export async function getSubscriptionSelf(): Promise<ApiResponse<UserSubscription[]>> {
  const res = await apiClient.get('/api/user/self/subscription/self')
  return res.data
}

// Admin subscription management
export async function adminListSubscriptionPlans(): Promise<ApiResponse<SubscriptionPlan[]>> {
  const res = await apiClient.get('/api/admin/subscription/plans')
  return res.data
}

export async function adminCreateSubscriptionPlan(data: Partial<SubscriptionPlan>): Promise<ApiResponse> {
  const res = await apiClient.post('/api/admin/subscription/plans', data)
  return res.data
}

export async function adminUpdateSubscriptionPlan(id: number, data: Partial<SubscriptionPlan>): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/admin/subscription/plans/${id}`, data)
  return res.data
}

export async function adminDeleteSubscriptionPlan(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/admin/subscription/plans/${id}`)
  return res.data
}

export async function adminBindSubscription(userId: number, planId: number): Promise<ApiResponse> {
  const res = await apiClient.post('/api/admin/subscription/bind', { user_id: userId, plan_id: planId })
  return res.data
}

// ---------------------------------------------------------------------------
// Notifications API
// ---------------------------------------------------------------------------

export interface Notification {
  id: number
  title: string
  content: string
  is_read: number
  created_at: string
}

export async function getNotifications(): Promise<ApiResponse<Notification[]>> {
  const res = await apiClient.get('/api/user/self/notifications')
  return res.data
}

export async function getUnreadNotificationCount(): Promise<ApiResponse<number>> {
  const res = await apiClient.get('/api/user/self/notifications/unread_count')
  return res.data
}

export async function markNotificationRead(id: number): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/user/self/notifications/${id}/read`)
  return res.data
}

export async function markAllNotificationsRead(): Promise<ApiResponse> {
  const res = await apiClient.put('/api/user/self/notifications/read_all')
  return res.data
}

// ---------------------------------------------------------------------------
// Team API
// ---------------------------------------------------------------------------

export interface TeamMember {
  id: number
  username: string
  display_name: string
  action: string
  amount: number
  created_at: string
}

export async function getMyTeam(): Promise<ApiResponse<TeamMember[]>> {
  const res = await apiClient.get('/api/user/self/team')
  return res.data
}

export async function upgradeToProvider(): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/upgrade')
  return res.data
}

// ---------------------------------------------------------------------------
// Security Activity API
// ---------------------------------------------------------------------------

export interface SecurityActivity {
  id: number
  action: string
  ip: string
  user_agent: string
  created_at: string
}

export async function getSecurityActivity(): Promise<ApiResponse<SecurityActivity[]>> {
  const res = await apiClient.get('/api/user/self/security/activity')
  return res.data
}

// ---------------------------------------------------------------------------
// Transaction Logs API
// ---------------------------------------------------------------------------

export interface TransactionLog {
  id: number
  user_id: number
  amount: number
  action: string
  before_quota: number
  after_quota: number
  status: string
  remark: string
  created_at: string
}

export async function getTransactionLogs(): Promise<ApiResponse<TransactionLog[]>> {
  const res = await apiClient.get('/api/user/self/transaction_logs')
  return res.data
}

// ---------------------------------------------------------------------------
// Two-Factor Authentication (2FA) API
// ---------------------------------------------------------------------------

export interface TwoFAStatus {
  enabled: boolean
  method: string
}

export interface TwoFAInitResponse {
  secret: string
  qr_code: string
}

export async function getTwoFAStatus(): Promise<ApiResponse<TwoFAStatus>> {
  const res = await apiClient.get('/api/user/2fa')
  return res.data
}

export async function initTwoFA(): Promise<ApiResponse<TwoFAInitResponse>> {
  const res = await apiClient.post('/api/user/2fa/init')
  return res.data
}

export async function verifyAndEnableTwoFA(code: string): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/2fa/enable', { code })
  return res.data
}

export async function disableTwoFA(code: string): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/2fa/disable', { code })
  return res.data
}

export async function verifyLoginTwoFA(code: string): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/2fa/verify', { code })
  return res.data
}

// ---------------------------------------------------------------------------
// Password API
// ---------------------------------------------------------------------------

export async function changePassword(oldPassword: string, newPassword: string): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/password/change', { old_password: oldPassword, new_password: newPassword })
  return res.data
}

export async function adminResetUserPassword(userId: number, newPassword: string): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/password/admin_reset', { user_id: userId, new_password: newPassword })
  return res.data
}

export async function emergencyPasswordReset(resetToken: string, newPassword: string): Promise<ApiResponse> {
  const res = await apiClient.post('/api/password/emergency-reset', { reset_token: resetToken, new_password: newPassword })
  return res.data
}

// ---------------------------------------------------------------------------
// Commission API
// ---------------------------------------------------------------------------

export interface CommissionSetting {
  id: number
  level: string
  rate: number
  enabled: number
}

export interface CommissionRecord {
  id: number
  user_id: number
  from_user_id: number
  amount: number
  rate: number
  status: string
  created_at: string
}

export async function getCommissionSetting(): Promise<ApiResponse<CommissionSetting[]>> {
  const res = await apiClient.get('/api/commission/setting')
  return res.data
}

export async function saveCommissionSetting(data: Partial<CommissionSetting>): Promise<ApiResponse> {
  const res = await apiClient.put('/api/commission/setting', data)
  return res.data
}

export async function getMyCommissionRecords(): Promise<ApiResponse<CommissionRecord[]>> {
  const res = await apiClient.get('/api/commission/self/records')
  return res.data
}

export async function requestCommissionWithdrawal(amount: number): Promise<ApiResponse> {
  const res = await apiClient.post('/api/commission/self/withdraw', { amount })
  return res.data
}

// ---------------------------------------------------------------------------
// Distributor API
// ---------------------------------------------------------------------------

export interface Distributor {
  id: number
  name: string
  contact: string
  status: number
  api_key: string
  balance: number
  created_time: number
}

export interface DistributorPricing {
  id: number
  distributor_id: number
  model_name: string
  price: number
}

export interface DistributorRevenue {
  total_revenue: number
  this_month: number
  last_month: number
}

export async function getDistributors(): Promise<ApiResponse<Distributor[]>> {
  const res = await apiClient.get('/api/distributor')
  return res.data
}

export async function createDistributor(data: Partial<Distributor>): Promise<ApiResponse> {
  const res = await apiClient.post('/api/distributor', data)
  return res.data
}

export async function updateDistributor(id: number, data: Partial<Distributor>): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/distributor/${id}`, data)
  return res.data
}

export async function getDistributorPricing(id: number): Promise<ApiResponse<DistributorPricing[]>> {
  const res = await apiClient.get(`/api/distributor/${id}/pricing`)
  return res.data
}

export async function setDistributorPricing(id: number, data: DistributorPricing[]): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/distributor/${id}/pricing`, { pricing: data })
  return res.data
}

export async function getDistributorRevenue(id: number): Promise<ApiResponse<DistributorRevenue>> {
  const res = await apiClient.get(`/api/distributor/${id}/revenue`)
  return res.data
}

export async function getMyDistributor(): Promise<ApiResponse<Distributor>> {
  const res = await apiClient.get('/api/distributor/self')
  return res.data
}

// ---------------------------------------------------------------------------
// Model Catalog API
// ---------------------------------------------------------------------------

export interface CatalogItem {
  model_name: string
  provider: string
  provider_type: number
  input_price: number
  output_price: number
  description: string
  category: string
  tags: string[]
  context_length: number
  max_output: number
  status: number
}

export async function getModelCatalog(): Promise<ApiResponse<CatalogItem[]>> {
  const res = await apiClient.get('/api/model-catalog')
  return res.data
}

export async function getModelDetail(modelName: string): Promise<ApiResponse<CatalogItem>> {
  const res = await apiClient.get(`/api/model-catalog/${encodeURIComponent(modelName)}`)
  return res.data
}

export async function syncModelMetadata(): Promise<ApiResponse> {
  const res = await apiClient.post('/api/models/sync')
  return res.data
}

// ---------------------------------------------------------------------------
// Model Sync API (admin)
// ---------------------------------------------------------------------------

export interface ModelSyncStatus {
  last_sync_time: number
  model_count: number
  auto_sync_enabled: boolean
}

export async function getModelSyncStatus(): Promise<ApiResponse<ModelSyncStatus>> {
  const res = await apiClient.get('/api/admin/model-sync')
  return res.data
}

export async function saveModelSyncSetting(data: Partial<ModelSyncStatus>): Promise<ApiResponse> {
  const res = await apiClient.put('/api/admin/model-sync', data)
  return res.data
}

export async function syncModels(): Promise<ApiResponse> {
  const res = await apiClient.post('/api/admin/model-sync/sync')
  return res.data
}

// ---------------------------------------------------------------------------
// Performance API
// ---------------------------------------------------------------------------

export interface PerformanceStats {
  uptime: number
  goroutines: number
  memory_mb: number
  cpu_usage: number
  request_rate: number
  avg_latency_ms: number
}

export async function getPerformanceStats(): Promise<ApiResponse<PerformanceStats>> {
  const res = await apiClient.get('/api/admin/performance')
  return res.data
}

// ---------------------------------------------------------------------------
// Channel Affinity API
// ---------------------------------------------------------------------------

export interface ChannelAffinitySetting {
  id: number
  model_name: string
  preferred_channel_id: number
  fallback_channel_ids: string
  enabled: number
}

export async function getChannelAffinitySettings(): Promise<ApiResponse<ChannelAffinitySetting[]>> {
  const res = await apiClient.get('/api/admin/channel-affinity')
  return res.data
}

export async function saveChannelAffinitySettings(data: Partial<ChannelAffinitySetting>): Promise<ApiResponse> {
  const res = await apiClient.put('/api/admin/channel-affinity', data)
  return res.data
}

export async function clearChannelAffinityCache(): Promise<ApiResponse> {
  const res = await apiClient.delete('/api/admin/channel-affinity/cache')
  return res.data
}

export async function getChannelAffinityCacheStats(): Promise<ApiResponse> {
  const res = await apiClient.get('/api/admin/channel-affinity/cache/stats')
  return res.data
}

// ---------------------------------------------------------------------------
// Channel Upstream API
// ---------------------------------------------------------------------------

export interface UpstreamUpdateSetting {
  auto_check: boolean
  interval_hours: number
  last_check_time: number
  upstream_version: string
}

export async function getUpstreamUpdateSetting(): Promise<ApiResponse<UpstreamUpdateSetting>> {
  const res = await apiClient.get('/api/admin/upstream')
  return res.data
}

export async function saveUpstreamUpdateSetting(data: Partial<UpstreamUpdateSetting>): Promise<ApiResponse> {
  const res = await apiClient.put('/api/admin/upstream', data)
  return res.data
}

export async function checkUpstreamUpdates(): Promise<ApiResponse<{ has_update: boolean; version: string }>> {
  const res = await apiClient.post('/api/admin/upstream/check')
  return res.data
}

// ---------------------------------------------------------------------------
// Channel Profit API
// ---------------------------------------------------------------------------

export interface ChannelProfitItem {
  channel_id: number
  channel_name: string
  total_quota: number
  cost: number
  revenue: number
  profit: number
  request_count: number
}

export async function getChannelProfit(): Promise<ApiResponse<ChannelProfitItem[]>> {
  const res = await apiClient.get('/api/channel/profit')
  return res.data
}

// ---------------------------------------------------------------------------
// Channel Pricing & Category
// ---------------------------------------------------------------------------

export async function setChannelPricing(channelId: number, costPerUnit: number, sellPriceRate: number): Promise<ApiResponse> {
  const res = await apiClient.post('/api/channel/pricing', { id: channelId, cost_per_unit: costPerUnit, sell_price_rate: sellPriceRate })
  return res.data
}

export async function setChannelCategory(channelId: number, category: string): Promise<ApiResponse> {
  const res = await apiClient.post('/api/channel/category', { id: channelId, category })
  return res.data
}

export async function deleteDisabledChannel(): Promise<ApiResponse> {
  const res = await apiClient.delete('/api/channel/disabled')
  return res.data
}

// ---------------------------------------------------------------------------
// Custom OAuth API
// ---------------------------------------------------------------------------

export interface CustomOAuthProvider {
  id: number
  name: string
  client_id: string
  client_secret: string
  auth_url: string
  token_url: string
  user_info_url: string
  enabled: boolean
  created_time: number
}

export async function listCustomOAuthProviders(): Promise<ApiResponse<CustomOAuthProvider[]>> {
  const res = await apiClient.get('/api/admin/custom-oauth')
  return res.data
}

export async function createCustomOAuthProvider(data: Partial<CustomOAuthProvider>): Promise<ApiResponse> {
  const res = await apiClient.post('/api/admin/custom-oauth', data)
  return res.data
}

export async function updateCustomOAuthProvider(id: number, data: Partial<CustomOAuthProvider>): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/admin/custom-oauth/${id}`, data)
  return res.data
}

export async function deleteCustomOAuthProvider(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/admin/custom-oauth/${id}`)
  return res.data
}

// ---------------------------------------------------------------------------
// Reseller API
// ---------------------------------------------------------------------------

export interface ResellerBalance {
  balance: number
  frozen: number
  total_earned: number
}

export interface ResellerStats {
  total_keys: number
  active_keys: number
  today_requests: number
  total_requests: number
  today_earnings: number
  total_earnings: number
}

export async function getResellerBalance(): Promise<ApiResponse<ResellerBalance>> {
  const res = await apiClient.get('/api/reseller/balance')
  return res.data
}

export async function submitResellerWithdrawal(amount: number): Promise<ApiResponse> {
  const res = await apiClient.post('/api/reseller/withdraw', { amount })
  return res.data
}

export async function getResellerStats(): Promise<ApiResponse<ResellerStats>> {
  const res = await apiClient.get('/api/reseller/stats')
  return res.data
}

export interface ResellerListItem {
  id: number
  name: string
  balance: number
  status: number
  created_time: number
}

export async function listResellers(): Promise<ApiResponse<ResellerListItem[]>> {
  const res = await apiClient.get('/api/admin/resellers')
  return res.data
}

export interface AdminWithdrawalItem {
  id: number
  reseller_id: number
  amount: number
  status: string
  created_time: number
}

export async function listAdminWithdrawals(): Promise<ApiResponse<AdminWithdrawalItem[]>> {
  const res = await apiClient.get('/api/admin/withdrawals')
  return res.data
}

export async function approveWithdrawal(id: number): Promise<ApiResponse> {
  const res = await apiClient.post(`/api/admin/withdrawals/${id}/approve`)
  return res.data
}

// ---------------------------------------------------------------------------
// Task API (Midjourney / Video / Suno)
// ---------------------------------------------------------------------------

export interface TaskInfo {
  id: string
  type: string
  status: string
  prompt: string
  progress: number
  result_url: string
  created_time: number
  updated_time: number
}

export async function createMidjourneyTask(prompt: string, params?: Record<string, unknown>): Promise<ApiResponse<TaskInfo>> {
  const res = await apiClient.post('/api/task/midjourney', { prompt, ...params })
  return res.data
}

export async function getMidjourneyTask(taskId: string): Promise<ApiResponse<TaskInfo>> {
  const res = await apiClient.get(`/api/task/midjourney/${taskId}`)
  return res.data
}

export async function createVideoTask(prompt: string, params?: Record<string, unknown>): Promise<ApiResponse<TaskInfo>> {
  const res = await apiClient.post('/api/task/video', { prompt, ...params })
  return res.data
}

export async function createSunoTask(prompt: string, params?: Record<string, unknown>): Promise<ApiResponse<TaskInfo>> {
  const res = await apiClient.post('/api/task/suno', { prompt, ...params })
  return res.data
}

export async function getTaskStatus(taskId: string): Promise<ApiResponse<TaskInfo>> {
  const res = await apiClient.get(`/api/task/${taskId}`)
  return res.data
}

export async function listUserTasks(): Promise<ApiResponse<TaskInfo[]>> {
  const res = await apiClient.get('/api/task')
  return res.data
}

export async function cancelTask(taskId: string): Promise<ApiResponse> {
  const res = await apiClient.post(`/api/task/${taskId}/cancel`)
  return res.data
}

export async function deleteTask(taskId: string): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/task/${taskId}`)
  return res.data
}

export async function adminGetAllTasks(): Promise<ApiResponse<TaskInfo[]>> {
  const res = await apiClient.get('/api/admin/task')
  return res.data
}

export async function adminPollTasks(): Promise<ApiResponse> {
  const res = await apiClient.post('/api/admin/task/poll')
  return res.data
}

// ---------------------------------------------------------------------------
// RSS / News API
// ---------------------------------------------------------------------------

export interface RssArticle {
  id: number
  title: string
  summary: string
  url: string
  source: string
  published_at: string
}

export async function getRssArticles(): Promise<ApiResponse<RssArticle[]>> {
  const res = await apiClient.get('/api/rss/articles')
  return res.data
}

// ---------------------------------------------------------------------------
// Enterprise Clients API
// ---------------------------------------------------------------------------

export interface EnterpriseClient {
  id: number
  name: string
  contact: string
  email: string
  status: number
  created_time: number
}

export async function listEnterpriseClients(): Promise<ApiResponse<EnterpriseClient[]>> {
  const res = await apiClient.get('/api/enterprise-clients')
  return res.data
}

// ---------------------------------------------------------------------------
// Fusion API
// ---------------------------------------------------------------------------

export interface FusionResult {
  id: string
  model_a: string
  model_b: string
  result: string
  created_time: number
}

export async function handleFusion(data: { model_a: string; model_b: string; prompt: string }): Promise<ApiResponse<FusionResult>> {
  const res = await apiClient.post('/api/fusion', data)
  return res.data
}

// ---------------------------------------------------------------------------
// Quantum API
// ---------------------------------------------------------------------------

export interface QuantumBackend {
  id: number
  name: string
  backend_type: string
  status: string
  qubits: number
  description: string
}

export interface QuantumProvider {
  name: string
  display_name: string
  description: string
  backends: QuantumBackend[]
}

export interface QuantumTaskInfo {
  id: string
  backend: string
  circuit: string
  shots: number
  status: string
  result: string
  created_time: number
}

export async function getQuantumBackends(): Promise<ApiResponse<QuantumBackend[]>> {
  const res = await apiClient.get('/api/quantum/backends')
  return res.data
}

export async function getQuantumProviders(): Promise<ApiResponse<QuantumProvider[]>> {
  const res = await apiClient.get('/api/quantum/providers')
  return res.data
}

export async function submitQuantumTask(data: { backend: string; circuit: string; shots?: number }): Promise<ApiResponse<QuantumTaskInfo>> {
  const res = await apiClient.post('/api/quantum/submit', data)
  return res.data
}

// ---------------------------------------------------------------------------
// Platform Config API
// ---------------------------------------------------------------------------

export interface PlatformConfig {
  key: string
  value: string
  description: string
  updated_at: string
}

export async function getPlatformConfigs(): Promise<ApiResponse<PlatformConfig[]>> {
  const res = await apiClient.get('/api/platform/config')
  return res.data
}

export async function updatePlatformConfig(key: string, value: string): Promise<ApiResponse> {
  const res = await apiClient.put('/api/platform/config', { key, value })
  return res.data
}

// ---------------------------------------------------------------------------
// Site Content API
// ---------------------------------------------------------------------------

export interface SiteContent {
  key: string
  content: string
  type: string
}

export async function getSiteContent(): Promise<ApiResponse<Record<string, string>>> {
  const res = await apiClient.get('/api/site-content')
  return res.data
}

export async function getSiteStats(): Promise<ApiResponse<Record<string, number>>> {
  const res = await apiClient.get('/api/site-stats')
  return res.data
}

export async function getSiteFeatures(): Promise<ApiResponse<Record<string, boolean>>> {
  const res = await apiClient.get('/api/site-features')
  return res.data
}

export async function getSiteProviders(): Promise<ApiResponse<Record<string, string>>> {
  const res = await apiClient.get('/api/site-providers')
  return res.data
}

// ---------------------------------------------------------------------------
// Promo Ads API
// ---------------------------------------------------------------------------

export interface PromoAd {
  id: number
  title: string
  content: string
  image_url: string
  link_url: string
  position: string
  enabled: number
  sort_order: number
  created_time: number
  updated_time: number
}

export async function getPromoAds(): Promise<ApiResponse<PromoAd[]>> {
  const res = await apiClient.get('/api/promo-ads')
  return res.data
}

// Admin promo ads
export async function adminGetAllPromoAds(): Promise<ApiResponse<PromoAd[]>> {
  const res = await apiClient.get('/api/admin/promo-ads')
  return res.data
}

export async function adminCreatePromoAd(data: Partial<PromoAd>): Promise<ApiResponse> {
  const res = await apiClient.post('/api/admin/promo-ads', data)
  return res.data
}

export async function adminUpdatePromoAd(data: Partial<PromoAd>): Promise<ApiResponse> {
  const res = await apiClient.put('/api/admin/promo-ads', data)
  return res.data
}

export async function adminDeletePromoAd(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/admin/promo-ads/${id}`)
  return res.data
}

// ---------------------------------------------------------------------------
// Menu API
// ---------------------------------------------------------------------------

export interface MenuItem {
  id: number
  menu_key: string
  parent_key: string
  menu_type: string
  label_key: string
  icon: string
  path: string
  sort_order: number
  roles: string
  group_name: string
  enabled: boolean
}

export async function getMenus(type?: string): Promise<ApiResponse<MenuItem[]>> {
  const res = await apiClient.get('/api/menus', { params: { type } })
  return res.data
}

export async function adminGetAllMenus(): Promise<ApiResponse<MenuItem[]>> {
  const res = await apiClient.get('/api/admin/menus')
  return res.data
}

export async function adminCreateOrUpdateMenu(data: Partial<MenuItem>): Promise<ApiResponse> {
  const res = await apiClient.post('/api/admin/menus', data)
  return res.data
}

export async function adminDeleteMenu(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/admin/menus/${id}`)
  return res.data
}

// ---------------------------------------------------------------------------
// OAuth Bindings
// ---------------------------------------------------------------------------

export async function oauthGitHub(): Promise<void> {
  window.location.href = '/api/oauth/github'
}

export async function oauthWeChat(): Promise<void> {
  window.location.href = '/api/oauth/wechat'
}

export async function oauthDiscord(): Promise<void> {
  window.location.href = '/api/oauth/discord'
}

export async function oauthLinuxDO(): Promise<void> {
  window.location.href = '/api/oauth/linuxdo'
}

export async function oauthTelegram(): Promise<void> {
  window.location.href = '/api/oauth/telegram'
}

export async function oauthLark(): Promise<void> {
  window.location.href = '/api/oauth/lark'
}

export async function oauthOidc(): Promise<void> {
  window.location.href = '/api/oauth/oidc'
}

export async function weChatBind(): Promise<void> {
  window.location.href = '/api/oauth/wechat/bind'
}

export async function discordBind(): Promise<void> {
  window.location.href = '/api/oauth/discord/bind'
}

export async function linuxDoBind(): Promise<void> {
  window.location.href = '/api/oauth/linuxdo/bind'
}

export async function telegramBind(): Promise<void> {
  window.location.href = '/api/oauth/telegram/bind'
}

// ---------------------------------------------------------------------------
// Email Verification & Password Reset
// ---------------------------------------------------------------------------

export async function sendEmailVerification(email?: string): Promise<ApiResponse> {
  const res = await apiClient.get('/api/verification', { params: { email } })
  return res.data
}

export async function sendPasswordResetEmail(email: string): Promise<ApiResponse> {
  const res = await apiClient.get('/api/reset_password', { params: { email } })
  return res.data
}

export async function resetPassword(data: { token: string; password: string }): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/reset', data)
  return res.data
}

export async function emailBind(email: string, code: string): Promise<ApiResponse> {
  const res = await apiClient.get('/api/oauth/email/bind', { params: { email, code } })
  return res.data
}

// ---------------------------------------------------------------------------
// WebAuthn / Passkey API
// ---------------------------------------------------------------------------

export async function webauthnBeginRegistration(): Promise<ApiResponse> {
  const res = await apiClient.post('/api/webauthn/register/begin')
  return res.data
}

export async function webauthnFinishRegistration(data: unknown): Promise<ApiResponse> {
  const res = await apiClient.post('/api/webauthn/register/finish', data)
  return res.data
}

export async function webauthnGetCredentials(): Promise<ApiResponse> {
  const res = await apiClient.get('/api/user/self/webauthn/credentials')
  return res.data
}

export async function webauthnDeleteCredential(id: string): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/user/self/webauthn/credentials/${id}`)
  return res.data
}

// ---------------------------------------------------------------------------
// Language / Translation API
// ---------------------------------------------------------------------------

export interface LanguageVersion {
  code: string
  name: string
  enabled: boolean
}

export async function getLanguageVersions(): Promise<ApiResponse<LanguageVersion[]>> {
  const res = await apiClient.get('/api/languages')
  return res.data
}

export async function getTranslations(langCode: string): Promise<ApiResponse<Record<string, string>>> {
  const res = await apiClient.get('/api/translations', { params: { lang: langCode } })
  return res.data
}

// ---------------------------------------------------------------------------
// Status / Misc
// ---------------------------------------------------------------------------

export interface SystemStatus {
  version: string
  uptime: number
  db_status: string
  redis_status: string
  memory_usage: number
  total_users: number
  total_channels: number
}

export async function getSystemStatus(): Promise<ApiResponse<SystemStatus>> {
  const res = await apiClient.get('/api/status')
  return res.data
}

export async function getNotice(): Promise<ApiResponse<string>> {
  const res = await apiClient.get('/api/notice')
  return res.data
}

export async function getAbout(): Promise<ApiResponse<Record<string, string>>> {
  const res = await apiClient.get('/api/about')
  return res.data
}

export async function getHomePageContent(): Promise<ApiResponse<Record<string, string>>> {
  const res = await apiClient.get('/api/home_page_content')
  return res.data
}

export async function getPrometheusMetrics(): Promise<string> {
  const res = await apiClient.get('/api/metrics')
  return res.data
}
