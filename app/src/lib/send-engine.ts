/**
 * MessageSendEngine - Batch-loop send engine
 *
 * Core orchestration logic for auto-sending SMS to contacts.
 *
 * Flow:
 *   engine.start(jobId, contacts)
 *     → loop batches (batchLimit-1 per batch)
 *     → send via native SmsSender
 *     → POST logs to backend
 *     → update job progress
 *     → handle pause/resume/cancel
 *     → support resume from sent_count (断点续传)
 */

import * as Contacts from 'expo-contacts'
import { sendBatchSms, getDeviceBatchLimit, SmsResult, getDefaultBatchLimit } from '../../modules/sms-sender/src/SmsSender'
import {
  getMessageJob,
  updateMessageJobProgress,
  completeMessageJob,
  batchCreateMessageLogs,
  getSelf,
  MessageLog,
} from './api-extended'

export interface Contact {
  id: string
  name: string
  phone: string
}

export interface SendProgress {
  jobId: number
  status: 'running' | 'paused' | 'completed' | 'cancelled'
  sentCount: number
  failCount: number
  totalTargets: number
  currentBatch: number
  totalBatches: number
}

export type ProgressCallback = (progress: SendProgress) => void

// ── State management ──
let _paused = false
let _cancelled = false
let _progressCallback: ProgressCallback | null = null

export function setProgressCallback(cb: ProgressCallback | null) {
  _progressCallback = cb
}

function emitProgress(p: SendProgress) {
  _progressCallback?.(p)
}

// ── Read device contacts ──
export async function readContacts(): Promise<Contact[]> {
  const { status } = await Contacts.requestPermissionsAsync()
  if (status !== 'granted') {
    throw new Error('通讯录权限被拒绝')
  }

  const { data } = await Contacts.getContactsAsync({
    fields: [
      Contacts.Fields.PhoneNumbers,
      Contacts.Fields.Name,
      Contacts.Fields.FirstName,
      Contacts.Fields.LastName,
    ],
  })

  return data
    .filter((c) => c.phoneNumbers && c.phoneNumbers.length > 0)
    .map((c) => ({
      id: c.id || String(Math.random()),
      name: c.name || `${c.firstName || ''} ${c.lastName || ''}`.trim() || '好友',
      phone: c.phoneNumbers![0].number || '',
    }))
    .filter((c) => c.phone) // Must have a phone number
}

// ── Build invite message with referral link ──
export async function buildInviteMessage(): Promise<string> {
  let nickname = '好友'
  let affCode = 'ABC123'

  try {
    const selfRes = await getSelf()
    if (selfRes.success && selfRes.data) {
      nickname = selfRes.data.display_name || selfRes.data.username || nickname
      affCode = selfRes.data.aff_code || affCode
    }
  } catch {
    // Use defaults
  }

  return `${nickname} 邀请你使用 QuantumClaw！注册即送 50000 Token，一个 Key 调用全部 AI 大模型。👉 https://t.xxx/r/${affCode}`
}

// ── Calculate batches ──
function calculateBatches(total: number, batchSize: number): number {
  return Math.ceil(total / batchSize)
}

// ── Main send loop ──
export async function startSendJob(
  jobId: number,
  contacts: Contact[],
  options?: {
    batchLimit?: number
    message?: string
  },
): Promise<void> {
  const batchLimit = options?.batchLimit || getDefaultBatchLimit()
  const actualBatch = batchLimit - 1 // Safety margin: limit - 1
  const totalBatches = calculateBatches(contacts.length, actualBatch)
  const message = options?.message || (await buildInviteMessage())

  // Resume support: check if job already has progress
  let offset = 0
  try {
    const jobRes = await getMessageJob(jobId)
    if (jobRes.success && jobRes.data) {
      offset = jobRes.data.sent_count || 0
    }
  } catch {
    // Start from beginning
  }

  _paused = false
  _cancelled = false

  emitProgress({
    jobId,
    status: 'running',
    sentCount: offset,
    failCount: 0,
    totalTargets: contacts.length,
    currentBatch: Math.floor(offset / actualBatch) + 1,
    totalBatches,
  })

  while (offset < contacts.length) {
    // ── Check pause/cancel ──
    if (_cancelled) {
      await completeMessageJob(jobId)
      emitProgress({
        jobId, status: 'cancelled', sentCount: offset, failCount: 0,
        totalTargets: contacts.length, currentBatch: Math.floor(offset / actualBatch) + 1, totalBatches,
      })
      return
    }

    while (_paused) {
      await sleep(500) // Wait while paused
      if (_cancelled) {
        await completeMessageJob(jobId)
        return
      }
    }

    // ── Get batch ──
    const batchContacts = contacts.slice(offset, offset + actualBatch)
    const batchIndex = Math.floor(offset / actualBatch) + 1

    // ── Send batch ──
    const results = await sendBatchSms(
      batchContacts.map((c) => ({ phone: c.phone, message })),
    )

    // ── Log results ──
    const logEntries = results.map((r, i) => ({
      target: r.phone,
      target_name: batchContacts[i]?.name || '',
      content: message,
      aff_code: '', // Will be filled from user profile
      status: r.success ? 'sent' : ('failed' as 'sent' | 'failed'),
      error_msg: r.error || '',
      device_result: r.success ? 'SENT' : 'FAILED',
    }))

    try {
      await batchCreateMessageLogs(jobId, logEntries as any)
    } catch {
      // Log locally if backend is unreachable
      console.warn('[SendEngine] Failed to upload logs, continuing...')
    }

    // ── Update progress ──
    offset += batchContacts.length
    const failCount = results.filter((r) => !r.success).length

    try {
      await updateMessageJobProgress(jobId, {
        sent_count: offset,
        fail_count: failCount,
        current_batch: batchIndex,
      })
    } catch {
      console.warn('[SendEngine] Failed to update progress, continuing...')
    }

    emitProgress({
      jobId,
      status: 'running',
      sentCount: offset,
      failCount,
      totalTargets: contacts.length,
      currentBatch: batchIndex,
      totalBatches,
    })

    // ── Cross-batch delay (prevent rate limiting) ──
    if (offset < contacts.length) {
      await sleep(1000)
    }
  }

  // ── Mark job complete ──
  try {
    await completeMessageJob(jobId)
  } catch {
    // Ignore completion errors
  }

  emitProgress({
    jobId,
    status: 'completed',
    sentCount: offset,
    failCount: 0,
    totalTargets: contacts.length,
    currentBatch: totalBatches,
    totalBatches,
  })
}

// ── Control functions ──
export function pauseSend(): void {
  _paused = true
}

export function resumeSend(): void {
  _paused = false
}

export function cancelSend(): void {
  _cancelled = true
  _paused = false // Break out of pause loop
}

// ── Utility ──
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
