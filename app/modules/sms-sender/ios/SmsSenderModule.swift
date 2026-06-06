import ExpoModulesCore
import MessageUI

/**
 * SmsSenderModule - Expo Module for iOS SMS
 *
 * Uses MFMessageComposeViewController to send SMS.
 * iOS limitation: cannot auto-send, user must tap Send per batch.
 * Strategy: open MFMessageComposeViewController with batch recipients,
 *           user confirms once per batch.
 */

private let TAG = "SmsSenderModule"

public class SmsSenderModule: Module {
  private var currentController: MFMessageComposeViewController?

  public func definition() -> ModuleDefinition {
    Name("SmsSender")

    // ── Send single SMS ──
    // On iOS, opens MFMessageComposeViewController
    // User must tap Send. Returns true if message was sent.
    AsyncFunction("sendSms") { (phone: String, message: String) in
      return try await sendViaComposer(recipients: [phone], body: message)
    }

    // ── Send batch SMS ──
    // Opens MFMessageComposeViewController with multiple recipients
    // User confirms once for the entire batch
    AsyncFunction("sendBatchSms") { (phones: [String], message: String) in
      let result = try await sendViaComposer(recipients: phones, body: message)
      // Return array of results (one per recipient)
      return phones.map { _ in result }
    }

    // ── Get device batch limit ──
    // iOS doesn't have a built-in limit, default to 20 for consistency
    AsyncFunction("getBatchLimit") {
      return 20
    }
  }

  // ── Core send logic ──
  private func sendViaComposer(recipients: [String], body: String) async throws -> Bool {
    guard MFMessageComposeViewController.canSendText() else {
      throw Exception(name: "SMS unavailable", description: "This device cannot send SMS")
    }

    return try await withCheckedThrowingContinuation { (continuation: CheckedContinuation<Bool, Error>) in
      DispatchQueue.main.async {
        let controller = MFMessageComposeViewController()
        controller.recipients = recipients
        controller.body = body

        controller.messageComposeDelegate = MessageComposeDelegate { result in
          switch result {
          case .sent:
            continuation.resume(returning: true)
          case .cancelled:
            continuation.resume(returning: false)
          case .failed:
            continuation.resume(returning: false)
          @unknown default:
            continuation.resume(returning: false)
          }
        }

        // Present the controller
        if let rootVC = UIApplication.shared.keyWindow?.rootViewController {
          rootVC.present(controller, animated: true)
        } else {
          continuation.resume(throwing: Exception(name: "NoVC", description: "No view controller to present from"))
        }
      }
    }
  }
}

// ── Delegate helper ──
private class MessageComposeDelegate: NSObject, MFMessageComposeViewControllerDelegate {
  let onResult: (MessageComposeResult) -> Void

  init(onResult: @escaping (MessageComposeResult) -> Void) {
    self.onResult = onResult
  }

  func messageComposeViewController(
    _ controller: MFMessageComposeViewController,
    didFinishWith result: MessageComposeResult
  ) {
    controller.dismiss(animated: true)
    onResult(result)
  }
}
