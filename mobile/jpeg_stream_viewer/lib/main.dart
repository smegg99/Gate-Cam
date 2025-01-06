import 'dart:io';
import 'package:flutter/material.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:path_provider/path_provider.dart';
import 'package:jpegsv/models/stream_element.dart';
import 'package:jpegsv/screens/main_screen.dart';
import 'package:jpegsv/screens/edit_screen.dart';
import 'package:jpegsv/screens/connect_screen.dart';
import 'package:jpegsv/theme/theme.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final directory = await _getAppDirectory();

  await Hive.initFlutter(directory.path);
  Hive.registerAdapter(StreamElementAdapter());
  await Hive.openBox<StreamElement>('stream_elements');

  runApp(const MyApp());
}

Future<Directory> _createAppDirectory(Directory appDir) async {
  try {
    if (!await appDir.exists()) {
      await appDir.create(recursive: true);
    }
  } catch (e) {
    print('Error creating app directory: $e');
    throw Exception('Failed to initialize app directory.');
  }

  return appDir;
}

Future<Directory> _getAppDirectory() async {
  if (Platform.isLinux || Platform.isMacOS) {
    final homeDir = Directory(Platform.environment['HOME'] ?? '');
    final appDir = Directory('${homeDir.path}/.jpegsv');
    return _createAppDirectory(appDir);
  } else if (Platform.isWindows) {
    final appData = Directory(Platform.environment['APPDATA'] ?? '');
    final appDir = Directory('${appData.path}\\jpegsv');
    return _createAppDirectory(appDir);
  } else {
    return await getApplicationDocumentsDirectory();
  }
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'JPEG Stream Viewer',
      debugShowCheckedModeBanner: false,
      theme: lightTheme,
      darkTheme: lightTheme,
      themeMode: ThemeMode.system,
      routes: {
        '/': (context) => const MainScreen(),
        '/edit': (context) => EditScreen(
              element:
                  ModalRoute.of(context)!.settings.arguments as StreamElement?,
            ),
        '/connect': (context) => ConnectScreen(
              element:
                  ModalRoute.of(context)!.settings.arguments as StreamElement,
            ),
      },
    );
  }
}
