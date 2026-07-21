import 'package:flutter_test/flutter_test.dart';
import 'package:revora_flutter/revora_flutter.dart';
import 'package:revora_flutter/revora_flutter_platform_interface.dart';
import 'package:revora_flutter/revora_flutter_method_channel.dart';
import 'package:plugin_platform_interface/plugin_platform_interface.dart';

class MockRevoraFlutterPlatform
    with MockPlatformInterfaceMixin
    implements RevoraFlutterPlatform {
  @override
  Future<String?> getPlatformVersion() => Future.value('42');
}

void main() {
  final RevoraFlutterPlatform initialPlatform = RevoraFlutterPlatform.instance;

  test('$MethodChannelRevoraFlutter is the default instance', () {
    expect(initialPlatform, isInstanceOf<MethodChannelRevoraFlutter>());
  });

  test('getPlatformVersion', () async {
    RevoraFlutter revoraFlutterPlugin = RevoraFlutter();
    MockRevoraFlutterPlatform fakePlatform = MockRevoraFlutterPlatform();
    RevoraFlutterPlatform.instance = fakePlatform;

    expect(await revoraFlutterPlugin.getPlatformVersion(), '42');
  });
}
