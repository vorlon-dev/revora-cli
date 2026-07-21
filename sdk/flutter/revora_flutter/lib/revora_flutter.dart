library revora_flutter;

import 'package:flutter/services.dart';

class Revora {
  static const MethodChannel _channel = MethodChannel('revora_flutter');

  /// Initialize with the GitHub repository that hosts the releases.
  /// Example: owner = "vorlon-dev", repo = "loveking"
  static Future<void> initialize({
    required String owner,
    required String repo,
    required String publicKey,
    String platform = 'flutter',
  }) async {
    await _channel.invokeMethod('initialize', {
      'owner': owner,
      'repo': repo,
      'publicKey': publicKey,
      'platform': platform,
    });
  }

  /// Check if an update is available by comparing current app version
  /// with the latest GitHub release tag.
  static Future<bool> checkForUpdate() async {
    return await _channel.invokeMethod<bool>('checkForUpdate') ?? false;
  }

  /// Download and apply the update. Returns true if restart is needed.
  static Future<bool> applyUpdate() async {
    return await _channel.invokeMethod<bool>('applyUpdate') ?? false;
  }

  /// Get the current installed app version.
  static Future<String> getCurrentVersion() async {
    return await _channel.invokeMethod<String>('getCurrentVersion') ?? '';
  }
}
