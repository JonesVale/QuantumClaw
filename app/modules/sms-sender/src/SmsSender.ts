/**
 * SmsSender - Native SMS Module
 *
 * Provides cross-platform API for sending SMS messages.
 *
 * Android: SmsManager.sendTextMessage() — auto-send with SEND_SMS permission
 * iOS: MFMessageComposeViewController — user taps send per batch
 *
 * Usage:
 *   const result = await SmsSender.send({ phone: '138xxxx', message: 'Hello' })
 *   const results = await SmsSender.sendBatch([{ phone, message }, ...])
 *   const limit = await SmsSender.getBatchLimit()
 */

import { Platform } from 'react-native'

// The native module is auto-linked by Expo
// In development, we provide a JS fallback for Expo Go
let NativeSmsSender: any = null

try {
  NativeSmsSender = require('expo-modules-core').requireNativeModule('SmsSender')
} catch {
  // Native module not available (e.g., Expo Go)
  // We'll provide a console-based fallback
  console.warn('[SmsSender] Native module not found, using JS fallback')
}

export interface SmsResult {
  phone: string
  success: boolean
  error?: string
}

export interface SendOptions {
  phone: string
  message: string
}

export interface BatchSendOptions {
  batch: SendOptions[]
  batchIndex: number
}

// ── Platform-specific batch limits ──
const PLATFORM_LIMITS: Record<string, number> = {
  huawei: 30,
  xiaomi: 100,
  oppo: 50,
  samsung: 100,
  default: 30,
}

/**
 * Detect device batch limit (per hour).
 * Returns a conservative value for the current device.
 */
export function getDefaultBatchLimit(): number {
  // Conservative default: safe for all devices
  return 20
}

/**
 * Send a single SMS message.
 * Android: auto-sends via SmsManager
 * iOS: opens MFMessageComposeViewController (user taps send)
 */
export async function sendSms(options: SendOptions): Promise<SmsResult> {
  if (NativeSmsSender) {
    try {
      const result = await NativeSmsSender.sendSms(options.phone, options.message)
      return { phone: options.phone, success: result }
    } catch (e: any) {
      return { phone: options.phone, success: false, error: e.message }
    }
  }

  // Fallback for Expo Go / dev
  console.log(`[SmsSender] Would send SMS to ${options.phone}: ${options.message.substring(0, 30)}...`)

  // Try using Linking.openURL as best-effort fallback
  if (Platform.OS === 'android') {
    try {
      const Linking = require('react-native').Linking
      await Linking.openURL(`sms:${options.phone}?body=${encodeURIComponent(options.message)}`)
      return { phone: options.phone, success: true }
    } catch (e: any) {
      return { phone: options.phone, success: false, error: e.message }
    }
  } else {
    // iOS fallback - just log, can't auto-send
    return { phone: options.phone, success: false, error: 'iOS requires user interaction per SMS' }
  }
}

/**
 * Send a batch of SMS messages.
 * Android: auto-sends each via SmsManager
 * iOS: opens MFMessageComposeViewController with batch recipients
 */
export async function sendBatchSms(
  batch: SendOptions[],
): Promise<SmsResult[]> {
  if (NativeSmsSender && Platform.OS === 'android') {
    try {
      const phones = batch.map((b) => b.phone)
      const message = batch[0]?.message || ''
      const results: boolean[] = await NativeSmsSender.sendBatchSms(phones, message)
      return batch.map((b, i) => ({
        phone: b.phone,
        success: results[i] ?? false,
      }))
    } catch (e: any) {
      return batch.map((b) => ({
        phone: b.phone,
        success: false,
        error: e.message,
      }))
    }
  }

  // Fallback: send one by one
  const results: SmsResult[] = []
  for (const item of batch) {
    const result = await sendSms(item)
    results.push(result)
    // Small delay between individual sends
    if (batch.length > 1) {
      await new Promise((resolve) => setTimeout(resolve, 200))
    }
  }
  return results
}

/**
 * Get the device-specific batch limit.
 */
export async function getDeviceBatchLimit(): Promise<number> {
  if (NativeSmsSender) {
    try {
      return await NativeSmsSender.getBatchLimit()
    } catch {
      return getDefaultBatchLimit()
    }
  }
  return getDefaultBatchLimit()
}
