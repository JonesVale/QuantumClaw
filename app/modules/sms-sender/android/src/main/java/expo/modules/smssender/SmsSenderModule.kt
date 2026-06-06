package expo.modules.smssender

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.telephony.SmsManager
import android.util.Log
import androidx.core.content.ContextCompat
import expo.modules.kotlin.modules.Module
import expo.modules.kotlin.modules.ModuleDefinition
import java.util.concurrent.Executors

/**
 * SmsSenderModule - Expo Module for Android auto-SMS
 *
 * Uses SmsManager.sendTextMessage() to send SMS programmatically.
 * Requires SEND_SMS permission (user grants during consent phase).
 *
 * Batch limit detection by device manufacturer:
 *   Huawei: 30/hour
 *   Xiaomi: 100/hour
 *   OPPO: 50/hour
 *   Samsung: 100/hour
 *   Default: 30/hour
 */

private const val TAG = "SmsSenderModule"

class SmsSenderModule : Module() {
  private val executor = Executors.newSingleThreadExecutor()

  override fun definition() = ModuleDefinition {
    Name("SmsSender")

    // ── Send single SMS ──
    AsyncFunction("sendSms") { phone: String, message: String ->
      val hasPermission = ContextCompat.checkSelfPermission(
        appContext.reactContext ?: return@AsyncFunction false,
        Manifest.permission.SEND_SMS
      ) == PackageManager.PERMISSION_GRANTED

      if (!hasPermission) {
        Log.w(TAG, "SEND_SMS permission not granted")
        return@AsyncFunction false
      }

      try {
        val smsManager = SmsManager.getDefault()
        smsManager.sendTextMessage(phone, null, message, null, null)
        Log.d(TAG, "SMS sent to $phone")
        true
      } catch (e: Exception) {
        Log.e(TAG, "SMS send failed to $phone: ${e.message}")
        false
      }
    }

    // ── Send batch SMS (auto-loop) ──
    AsyncFunction("sendBatchSms") { phones: List<String>, message: String ->
      val hasPermission = ContextCompat.checkSelfPermission(
        appContext.reactContext ?: return@AsyncFunction buildFailedResults(phones.size),
        Manifest.permission.SEND_SMS
      ) == PackageManager.PERMISSION_GRANTED

      if (!hasPermission) {
        Log.w(TAG, "SEND_SMS permission not granted for batch")
        return@AsyncFunction buildFailedResults(phones.size)
      }

      val results = mutableListOf<Boolean>()
      try {
        val smsManager = SmsManager.getDefault()
        for (phone in phones) {
          try {
            smsManager.sendTextMessage(phone, null, message, null, null)
            results.add(true)
          } catch (e: Exception) {
            Log.e(TAG, "Batch SMS failed for $phone: ${e.message}")
            results.add(false)
          }
        }
      } catch (e: Exception) {
        Log.e(TAG, "Batch SMS error: ${e.message}")
        // Fill remaining
        while (results.size < phones.size) results.add(false)
      }
      results
    }

    // ── Get device batch limit ──
    AsyncFunction("getBatchLimit") {
      getDeviceLimit()
    }
  }

  private fun getDeviceLimit(): Int {
    val manufacturer = Build.MANUFACTURER.lowercase()
    return when {
      manufacturer.contains("huawei") -> 30
      manufacturer.contains("xiaomi") -> 100
      manufacturer.contains("oppo") || manufacturer.contains("oneplus") -> 50
      manufacturer.contains("samsung") -> 100
      manufacturer.contains("vivo") -> 50
      else -> 30
    }
  }

  private fun buildFailedResults(count: Int): List<Boolean> {
    return List(count) { false }
  }
}
