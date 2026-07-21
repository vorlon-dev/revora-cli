import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:revora_flutter/revora_flutter.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // ⚠️ Replace with your actual GitHub repo details and Base64‑encoded public key
  await Revora.initialize(
    owner: 'vorlon-dev', // your GitHub username or org
    repo: 'loveking', // the repository containing releases
    publicKey: 'BASE64_PUBLIC_KEY', // run the command below to get this
  );

  runApp(MyApp());
}

class MyApp extends StatefulWidget {
  @override
  _MyAppState createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  bool _updateAvailable = false;

  @override
  void initState() {
    super.initState();
    checkForUpdates();
  }

  Future<void> checkForUpdates() async {
    bool available = await Revora.checkForUpdate();
    setState(() {
      _updateAvailable = available;
    });
  }

  Future<void> applyUpdate() async {
    bool success = await Revora.applyUpdate();
    if (success) {
      if (!mounted) return;
      showDialog(
        context: context,
        builder: (_) => AlertDialog(
          title: Text('Update Ready'),
          content: Text(
            'The update has been downloaded and verified.\n'
            'Please restart the app to apply it.',
          ),
          actions: [
            TextButton(
              onPressed: () {
                // Exit the app gracefully (the update will be applied on next launch)
                SystemNavigator.pop();
              },
              child: Text('Exit App'),
            ),
          ],
        ),
      );
    } else {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Update failed. Check your connection.')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        appBar: AppBar(title: Text('Revora OTA Demo')),
        body: Center(
          child: _updateAvailable
              ? ElevatedButton(
                  onPressed: applyUpdate,
                  child: Text('Apply Update'),
                )
              : Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text('No updates available'),
                    SizedBox(height: 12),
                    FutureBuilder<String>(
                      future: Revora.getCurrentVersion(),
                      builder: (context, snapshot) {
                        final version = snapshot.data ?? '...';
                        return Text('Current version: $version');
                      },
                    ),
                  ],
                ),
        ),
      ),
    );
  }
}
