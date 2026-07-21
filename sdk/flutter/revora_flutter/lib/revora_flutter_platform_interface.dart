import 'package:plugin_platform_interface/plugin_platform_interface.dart';

import 'revora_flutter_method_channel.dart';

abstract class RevoraFlutterPlatform extends PlatformInterface {
  /// Constructs a RevoraFlutterPlatform.
  RevoraFlutterPlatform() : super(token: _token);

  static final Object _token = Object();

  static RevoraFlutterPlatform _instance = MethodChannelRevoraFlutter();

  /// The default instance of [RevoraFlutterPlatform] to use.
  ///
  /// Defaults to [MethodChannelRevoraFlutter].
  static RevoraFlutterPlatform get instance => _instance;

  /// Platform-specific implementations should set this with their own
  /// platform-specific class that extends [RevoraFlutterPlatform] when
  /// they register themselves.
  static set instance(RevoraFlutterPlatform instance) {
    PlatformInterface.verifyToken(instance, _token);
    _instance = instance;
  }

  Future<String?> getPlatformVersion() {
    throw UnimplementedError('platformVersion() has not been implemented.');
  }
}
