import 'dart:convert';
import 'dart:typed_data';
import 'dart:io';
import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_mjpeg/flutter_mjpeg.dart';
import 'package:path_provider/path_provider.dart';
import 'package:jpegsv/models/stream_element.dart';

class ConnectScreen extends StatefulWidget {
  final StreamElement element;

  const ConnectScreen({super.key, required this.element});

  @override
  State<ConnectScreen> createState() => _ConnectScreenState();
}

class _ConnectScreenState extends State<ConnectScreen> {
  late String streamUrl;
  late Map<String, String> headers;
  bool isPaused = false;
  bool hasError = false;
  bool isLoading = false;
  Uint8List? lastFrame;
  final GlobalKey _streamKey = GlobalKey();
  bool _isMounted = true;

  @override
  void initState() {
    super.initState();
    streamUrl = widget.element.url;
    headers = {
      'Authorization':
          'Basic ${base64Encode(utf8.encode('${widget.element.username}:${widget.element.password}'))}',
    };
  }

  @override
  void dispose() {
    _isMounted = false;
    super.dispose();
  }

  Future<void> _captureLastFrame() async {
    try {
      RenderRepaintBoundary boundary = _streamKey.currentContext
          ?.findRenderObject() as RenderRepaintBoundary;

      ui.Image image = await boundary.toImage();

      ByteData? byteData =
          await image.toByteData(format: ui.ImageByteFormat.png);

      if (byteData != null && _isMounted) {
        setState(() {
          lastFrame = byteData.buffer.asUint8List();
        });
      }
    } catch (e) {
      if (_isMounted) {
        print('Error capturing last frame: $e');
      }
    }
  }

  Future<void> _captureAndSaveSnapshot() async {
    try {
      RenderRepaintBoundary boundary = _streamKey.currentContext
          ?.findRenderObject() as RenderRepaintBoundary;

      ui.Image image = await boundary.toImage();

      ByteData? byteData =
          await image.toByteData(format: ui.ImageByteFormat.png);

      if (byteData != null) {
        Uint8List pngBytes = byteData.buffer.asUint8List();

        String filePath = await _saveToLocalStorage(pngBytes);

        if (_isMounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Snapshot saved to $filePath')),
          );
        }
      }
    } catch (e) {
      if (_isMounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error capturing snapshot: $e')),
        );
      }
    }
  }

  Future<String> _saveToLocalStorage(Uint8List bytes) async {
    final directory = await getApplicationDocumentsDirectory();
    final fileName = 'snapshot_${DateTime.now().millisecondsSinceEpoch}.png';
    final file = File('${directory.path}/$fileName');

    await file.writeAsBytes(bytes);
    return file.path;
  }

  Future<void> _reconnectStream() async {
    if (_isMounted) {
      setState(() {
        hasError = false;
        isLoading = true;
      });

      await Future.delayed(const Duration(seconds: 1));

      if (_isMounted) {
        setState(() {
          isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Connecting to ${widget.element.name}'),
        actions: [
          if (isLoading)
            const Padding(
              padding: EdgeInsets.all(8.0),
              child: CircularProgressIndicator(
                valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
              ),
            ),
        ],
      ),
      body: Center(
        child: RepaintBoundary(
          key: _streamKey,
          child: Stack(
            children: [
              if (!hasError)
                InteractiveViewer(
                  panEnabled: !isPaused,
                  scaleEnabled: !isPaused,
                  minScale: 1.0,
                  maxScale: 5.0,
                  child: Mjpeg(
                    stream: streamUrl,
                    headers: headers,
                    isLive: true,
                    error: (context, error, stackTrace) {
                      if (_isMounted) {
                        Future.microtask(() {
                          if (_isMounted) {
                            setState(() {
                              hasError = true;
                            });
                          }
                        });
                      }
                      return const SizedBox.shrink();
                    },
                  ),
                ),
              if (hasError)
                Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: const [
                      Icon(Icons.error_outline, size: 64, color: Colors.red),
                      SizedBox(height: 8),
                      Text(
                        'Failed to load stream',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 18,
                        ),
                      ),
                      SizedBox(height: 16),
                      Text(
                        'Retrying...',
                        style: TextStyle(fontSize: 14, color: Colors.grey),
                      ),
                    ],
                  ),
                ),
              if (isPaused && lastFrame != null)
                Positioned.fill(
                  child: Image.memory(
                    lastFrame!,
                    fit: BoxFit.cover,
                  ),
                ),
            ],
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _captureAndSaveSnapshot,
        child: const Icon(Icons.camera),
        tooltip: 'Capture Snapshot',
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
      bottomNavigationBar: BottomAppBar(
        shape: const CircularNotchedRectangle(),
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.start,
            children: [
              IconButton(
                icon: Icon(isPaused ? Icons.play_arrow : Icons.pause),
                onPressed: () async {
                  if (!isPaused) {
                    await _captureLastFrame();
                  }
                  if (_isMounted) {
                    setState(() {
                      isPaused = !isPaused;
                    });
                  }
                },
                tooltip: isPaused ? 'Resume Stream' : 'Pause Stream',
              ),
            ],
          ),
        ),
      ),
    );
  }
}
