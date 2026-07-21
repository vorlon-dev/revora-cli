import Flutter
import UIKit

public class RevoraFlutterPlugin: NSObject, FlutterPlugin {
    var owner: String = ""
    var repo: String = ""
    var publicKey: SecKey?

    public static func register(with registrar: FlutterPluginRegistrar) {
        let channel = FlutterMethodChannel(name: "revora_flutter", binaryMessenger: registrar.messenger())
        let instance = RevoraFlutterPlugin()
        registrar.addMethodCallDelegate(instance, channel: channel)
    }

    public func handle(_ call: FlutterMethodCall, result: @escaping FlutterResult) {
        switch call.method {
        case "initialize":
            if let args = call.arguments as? [String: Any] {
                owner = args["owner"] as? String ?? ""
                repo = args["repo"] as? String ?? ""
                if let keyStr = args["publicKey"] as? String {
                    publicKey = loadPublicKey(keyStr)
                }
            }
            result(nil)
        case "checkForUpdate":
            checkForUpdate { available in result(available) }
        case "applyUpdate":
            applyUpdate { applied in result(applied) }
        case "getCurrentVersion":
            result(getCurrentVersion())
        default:
            result(FlutterMethodNotImplemented)
        }
    }

    private func loadPublicKey(_ keyStr: String) -> SecKey? {
        guard let data = Data(base64Encoded: keyStr) else { return nil }
        let attributes: [String: Any] = [
            kSecAttrKeyType as String: kSecAttrKeyTypeRSA,
            kSecAttrKeyClass as String: kSecAttrKeyClassPublic
        ]
        return SecKeyCreateWithData(data as CFData, attributes as CFDictionary, nil)
    }

    private func getCurrentVersion() -> String {
        return Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "0.0.0"
    }

    private func checkForUpdate(completion: @escaping (Bool) -> Void) {
        let currentVersion = getCurrentVersion()
        let url = URL(string: "https://api.github.com/repos/\(owner)/\(repo)/releases/latest")!
        URLSession.shared.dataTask(with: url) { data, _, error in
            guard let data = data, error == nil,
                  let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
                  let tagName = json["tag_name"] as? String,
                  tagName != currentVersion else {
                completion(false)
                return
            }
            if let assets = json["assets"] as? [[String: Any]] {
                for asset in assets {
                    if let name = asset["name"] as? String,
                       let downloadUrl = asset["browser_download_url"] as? String {
                        switch name {
                        case "update.patch":
                            UserDefaults.standard.set(downloadUrl, forKey: "revora_patch_url")
                        case "manifest.json":
                            UserDefaults.standard.set(downloadUrl, forKey: "revora_manifest_url")
                        case "manifest.json.sig":
                            UserDefaults.standard.set(downloadUrl, forKey: "revora_signature_url")
                        default: break
                        }
                    }
                }
                completion(true)
            } else {
                completion(false)
            }
        }.resume()
    }

    private func applyUpdate(completion: @escaping (Bool) -> Void) {
        guard let manifestUrl = UserDefaults.standard.string(forKey: "revora_manifest_url"),
              let patchUrl = UserDefaults.standard.string(forKey: "revora_patch_url"),
              let signatureUrl = UserDefaults.standard.string(forKey: "revora_signature_url"),
              let publicKey = publicKey else {
            completion(false)
            return
        }

        // Download manifest and signature
        downloadText(url: URL(string: manifestUrl)!) { manifestData in
            guard let manifestData = manifestData else { completion(false); return }
            self.downloadText(url: URL(string: signatureUrl)!) { signatureB64 in
                guard let signatureB64 = signatureB64,
                      let signatureData = Data(base64Encoded: signatureB64) else { completion(false); return }

                var error: Unmanaged<CFError>?
                let valid = SecKeyVerifySignature(publicKey, .rsaSignatureMessagePKCS1v15SHA256, manifestData as CFData, signatureData as CFData, &error)
                if !valid { completion(false); return }

                // Download patch
                self.downloadFile(url: URL(string: patchUrl)!) { localURL in
                    guard let localURL = localURL else { completion(false); return }
                    let pendingDir = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first!
                        .appendingPathComponent("revora_pending")
                    try? FileManager.default.createDirectory(at: pendingDir, withIntermediateDirectories: true)
                    let dest = pendingDir.appendingPathComponent("update.patch")
                    try? FileManager.default.copyItem(at: localURL, to: dest)
                    UserDefaults.standard.set(true, forKey: "revora_pending_restart")
                    completion(true)
                }
            }
        }
    }

    private func downloadText(url: URL, completion: @escaping (Data?) -> Void) {
        URLSession.shared.dataTask(with: url) { data, _, _ in completion(data) }.resume()
    }

    private func downloadFile(url: URL, completion: @escaping (URL?) -> Void) {
        URLSession.shared.downloadTask(with: url) { localURL, _, _ in completion(localURL) }.resume()
    }
}