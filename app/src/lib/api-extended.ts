/**
 * QuantumClaw Mobile - Re-exported API functions
 * 
 * Direct import from the existing web frontend api-extended.ts.
 * This file has NO DOM dependency — safe for React Native.
 */

import apiClient, { ApiResponse } from './api'
export type { ApiResponse }

// Re-export the helper
function extractParams<T>(arg1: unknown, arg2?: unknown): T | undefined {
  if (arg1 && typeof arg1 === 'object') {
    const obj = arg1 as Record<string, unknown>
    if ('queryKey' in obj || 'signal' in obj) {
      return arg2 as T
    }
  }
  return arg1 as T
}

// ═══════════════════════════════════════════════════════════════════════════
// Authentication
// ═══════════════════════════════════════════════════════════════════════════

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
  display_name?: string
  email?: string
  verification_code?: string
  wechat_id?: string
  ref?: string      // 邀请码
}): Promise<ApiResponse> {
  const res = await apiClient.post('/api/user/register', data)
  return res.data
}

export async function getSelf(): Promise<ApiResponse<{
  id: number
  username: string
  email: string
  display_name: string
  role: number
  status: number
  balance: number
  used_quota: number
  quota: number
  request_count: number
  group: string
  aff_code: string
  user_type?: string
  organization?: { id: number; name: string; tier: string }
}>> {
  const res = await apiClient.get('/api/user/self', {
    skipErrorHandler: true,
  } as unknown as Record<string, unknown>)
  return res.data
}

// ═══════════════════════════════════════════════════════════════════════════
// User & Profile
// ═══════════════════════════════════════════════════════════════════════════

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

export async function updateUser(data: Partial<User>): Promise<ApiResponse> {
  const res = await apiClient.put('/api/user/self', data)
  return res.data
}

// ═══════════════════════════════════════════════════════════════════════════
// API Keys
// ═══════════════════════════════════════════════════════════════════════════

export interface Token {
  id: number
  user_id: number
  key: string
  name: string
  created_time: number
  accessed_time: number
  expired_time: number
  remain_quota: number
  unlimited_quota: boolean
  status: number
  models?: string
  subnet?: string
  used_quota?: number
}

export async function getTokens(): Promise<ApiResponse<Token[]>> {
  const res = await apiClient.get('/api/token')
  return res.data
}

export async function createToken(data: {
  name: string
  remain_quota?: number
  expired_time?: number
  unlimited_quota?: boolean
  models?: string
  subnet?: string
}): Promise<ApiResponse<Token>> {
  const res = await apiClient.post('/api/token', data)
  return res.data
}

export async function deleteToken(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/token/${id}`)
  return res.data
}

export async function updateToken(id: number, data: Partial<Token>): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/token/${id}`, data)
  return res.data
}

// ═══════════════════════════════════════════════════════════════════════════
// Dashboard / Stats
// ═══════════════════════════════════════════════════════════════════════════

export interface DashboardData {
  quota: number
  used_quota: number
  balance: number
  request_count: number
  token_count: number
  today_quota: number
  today_used_quota: number
  today_request_count: number
  daily_usage?: Array<{ date: string; amount: number }>
  model_usage?: Array<{ model: string; count: number; quota: number }>
}

export async function getDashboardData(): Promise<ApiResponse<DashboardData>> {
  const res = await apiClient.get('/api/user/dashboard')
  return res.data
}

// ═══════════════════════════════════════════════════════════════════════════
// Wallet / Balance
// ═══════════════════════════════════════════════════════════════════════════

export interface Transaction {
  id: number
  user_id: number
  amount: number
  type: string
  description: string
  created_time: number
}

export async function getTransactions(
  page: number = 1,
  pageSize: number = 20
): Promise<ApiResponse<Transaction[]>> {
  const res = await apiClient.get('/api/user/transactions', {
    params: { page, page_size: pageSize },
  })
  return res.data
}

export interface TopUpResponse {
  url: string        // payment redirect URL
  order_no: string
}

export async function createTopUpOrder(
  amount: number,
  paymentMethod: string = 'alipay'
): Promise<ApiResponse<TopUpResponse>> {
  const res = await apiClient.post('/api/user/topup', {
    amount,
    payment_method: paymentMethod,
  })
  return res.data
}

// ═══════════════════════════════════════════════════════════════════════════
// Provider / Channel APIs
// ═══════════════════════════════════════════════════════════════════════════

export interface Channel {
  id: number
  name: string
  type: number
  key: string
  base_url: string
  models: string
  group: string
  priority: number
  weight: number
  status: number
  created_time: number
  response_time?: number
}

export async function getChannels(): Promise<ApiResponse<Channel[]>> {
  const res = await apiClient.get('/api/channel')
  return res.data
}

export async function createChannel(data: Partial<Channel>): Promise<ApiResponse> {
  const res = await apiClient.post('/api/channel', data)
  return res.data
}

export async function updateChannel(id: number, data: Partial<Channel>): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/channel/${id}`, data)
  return res.data
}

export async function deleteChannel(id: number): Promise<ApiResponse> {
  const res = await apiClient.delete(`/api/channel/${id}`)
  return res.data
}

export async function testChannel(id: number): Promise<ApiResponse> {
  const res = await apiClient.get(`/api/channel/${id}/test`)
  return res.data
}

// ═══════════════════════════════════════════════════════════════════════════
// Provider Store (店铺)
// ═══════════════════════════════════════════════════════════════════════════

export interface ProviderStore {
  id: number
  user_id: number
  slug: string
  name: string
  description: string
  logo_url: string
  contact_info: string
  is_active: boolean
  created_at: string
}

export async function getMyStore(): Promise<ApiResponse<ProviderStore>> {
  const res = await apiClient.get('/api/provider/store')
  return res.data
}

export async function updateStore(data: Partial<ProviderStore>): Promise<ApiResponse> {
  const res = await apiClient.put('/api/provider/store', data)
  return res.data
}

export async function getPublicStore(slug: string): Promise<ApiResponse<ProviderStore>> {
  const res = await apiClient.get(`/api/stores/${slug}`)
  return res.data
}

// ═══════════════════════════════════════════════════════════════════════════
// Provider Analytics
// ═══════════════════════════════════════════════════════════════════════════

export interface ProviderStats {
  total_revenue: number
  total_commission: number
  total_requests: number
  active_customers: number
  daily_revenue: Array<{ date: string; amount: number }>
  top_customers: Array<{ user_id: number; username: string; revenue: number }>
}

export async function getProviderStats(): Promise<ApiResponse<ProviderStats>> {
  const res = await apiClient.get('/api/provider/stats')
  return res.data
}

// ═══════════════════════════════════════════════════════════════════════════
// Enterprise APIs (企业)
// ═══════════════════════════════════════════════════════════════════════════

export interface Department {
  id: number
  org_id: number
  name: string
  description: string
  head_user_id: number
  monthly_budget: number
  alert_threshold: number
  status: number
}

export async function getDepartments(orgId: number): Promise<ApiResponse<Department[]>> {
  const res = await apiClient.get(`/api/org/${orgId}/departments`)
  return res.data
}

export async function createDepartment(orgId: number, data: {
  name: string; description?: string; head_user_id?: number
  monthly_budget?: number; alert_threshold?: number
}): Promise<ApiResponse> {
  const res = await apiClient.post(`/api/org/${orgId}/departments`, data)
  return res.data
}

export async function getEnterpriseTokenPolicy(orgId: number): Promise<ApiResponse> {
  const res = await apiClient.get(`/api/org/${orgId}/policy`)
  return res.data
}

export async function saveEnterprisePolicy(orgId: number, data: Record<string, unknown>): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/org/${orgId}/policy`, data)
  return res.data
}

export interface EnterpriseApproval {
  id: number
  org_id: number
  type: string
  status: string
  request_by: number
  content: string
  reason: string
  created_at: string
}

export async function getApprovals(orgId: number, status?: string): Promise<ApiResponse<EnterpriseApproval[]>> {
  const res = await apiClient.get(`/api/org/${orgId}/approvals`, {
    params: { status },
  })
  return res.data
}

export async function processApproval(
  orgId: number, approvalId: number, 
  status: string, remark?: string
): Promise<ApiResponse> {
  const res = await apiClient.post(`/api/org/${orgId}/approvals/${approvalId}`, {
    status, remark,
  })
  return res.data
}

// ═══════════════════════════════════════════════════════════════════════════
// Message / Invite APIs (移动端新增)
// ═══════════════════════════════════════════════════════════════════════════

export interface MessageJob {
  id: number
  user_id: number
  channel: string
  batch_limit: number
  total_targets: number
  sent_count: number
  fail_count: number
  current_batch: number
  status: string
  agreement_version: string
  created_at: string
  completed_at?: string
}

export interface MessageLog {
  id: number
  job_id: number
  batch: number
  target: string
  target_name: string
  content: string
  aff_code: string
  status: string
  error_msg?: string
  created_at: string
}

export interface MessageAgreement {
  id: number
  version: string
  title: string
  content: string
  is_active: boolean
  created_at: string
}

export async function createMessageJob(data: {
  channel: string
  total_targets: number
  agreement_version: string
}): Promise<ApiResponse<MessageJob>> {
  const res = await apiClient.post('/api/message/jobs', data)
  return res.data
}

export async function getMessageJobs(): Promise<ApiResponse<MessageJob[]>> {
  const res = await apiClient.get('/api/message/jobs')
  return res.data
}

export async function getMessageJob(id: number): Promise<ApiResponse<MessageJob>> {
  const res = await apiClient.get(`/api/message/jobs/${id}`)
  return res.data
}

export async function updateMessageJobProgress(
  id: number, data: {
    sent_count: number; fail_count: number; current_batch: number
  }
): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/message/jobs/${id}/progress`, data)
  return res.data
}

export async function completeMessageJob(id: number): Promise<ApiResponse> {
  const res = await apiClient.put(`/api/message/jobs/${id}/complete`)
  return res.data
}

export async function batchCreateMessageLogs(
  jobId: number, logs: Array<Omit<MessageLog, 'id' | 'job_id' | 'created_at'>>
): Promise<ApiResponse> {
  const res = await apiClient.post('/api/message/logs/batch', {
    job_id: jobId, logs,
  })
  return res.data
}

export async function getJobLogs(
  jobId: number, offset?: number, limit?: number
): Promise<ApiResponse<MessageLog[]>> {
  const res = await apiClient.get(`/api/message/jobs/${jobId}/logs`, {
    params: { offset, limit },
  })
  return res.data
}

export async function getActiveAgreement(): Promise<ApiResponse<MessageAgreement>> {
  const res = await apiClient.get('/api/message/agreements/active')
  return res.data
}

// ═══════════════════════════════════════════════════════════════════════════
// Status / Misc
// ═══════════════════════════════════════════════════════════════════════════

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
