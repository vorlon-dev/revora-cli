import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import 'revora_flutter_platform_interface.dart';

/// An implementation of [RevoraFlutterPlatform] that uses method channels.
class MethodChannelRevoraFlutter extends RevoraFlutterPlatform {
  /// The method channel used to interact with the native platform.
  @visibleForTesting
  final methodChannel = const MethodChannel('revora_flutter');

  @override
  Future<String?> getPlatformVersion() async {
    final version = await methodChannel.invokeMethod<String>(
      'getPlatformVersion',
    );
    return version;
  }
}
