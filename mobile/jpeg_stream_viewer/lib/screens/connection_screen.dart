import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_mjpeg/flutter_mjpeg.dart';
import 'package:flutter_speed_dial/flutter_speed_dial.dart';
import 'package:jpegsv/models/stream_element.dart';
import 'package:jpegsv/localization/localization.dart';

class ConnectionScreen extends StatefulWidget {
  final StreamElement element;
  final Directory appDirectory;

  const ConnectionScreen({
    super.key,
    required this.element,
    required this.appDirectory,
  });

  @override
  State<ConnectionScreen> createState() => _ConnectionScreenState();
}

class _ConnectionScreenState extends State<ConnectionScreen>
    with TickerProviderStateMixin {
  late String streamUrl;
  late Map<String, String> headers;

  AppLocalizations get localizations => AppLocalizations.of(context);
  ThemeData get theme => Theme.of(context);

  bool isPaused = false;
  bool hasError = false;
  bool isLoading = true;
  Uint8List? lastFrame;
  final GlobalKey _mjpegKey = GlobalKey();
  Timer? _reconnectTimer;

  final TransformationController _transformationController =
      TransformationController();

  late AnimationController _resetAnimationController;
  int _reconnectAttempt = 0;

  @override
  void initState() {
    super.initState();
    streamUrl = widget.element.url;
    headers = {
      'Authorization':
          'Basic ${base64Encode(utf8.encode('${widget.element.username}:${widget.element.password}'))}',
    };
    _initializeStream();

    _resetAnimationController = AnimationController(
      duration: const Duration(milliseconds: 300),
      vsync: this,
    );
  }

  @override
  void dispose() {
    _reconnectTimer?.cancel();
    _transformationController.dispose();
    _resetAnimationController.dispose();
    super.dispose();
  }

  void _performAction(ActionElement action) {
    showModalBottomSheet(
      context: context,
      isDismissible: false,
      isScrollControlled: true,
      builder: (BuildContext context) {
        bool isLoading = true;
        String? statusMessage;

        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setModalState) {
            if (isLoading) {
              Future.delayed(Duration.zero, () async {
                try {
                  final request = await HttpClient()
                      .openUrl(action.method, Uri.parse(action.endpoint));
                  action.headers.forEach((key, value) {
                    request.headers.add(key, value);
                  });

                  final result = await request.close();

                  if (result.statusCode == 200) {
                    setModalState(() {
                      isLoading = false;
                      statusMessage = localizations.translateWithParams(
                        'screens.connection.labels.action_executed_successfully',
                        {'name': action.name},
                      );
                    });
                  } else {
                    throw Exception(
                        'Failed with status code ${result.statusCode}');
                  }
                } catch (e) {
                  setModalState(() {
                    isLoading = false;
                    statusMessage = localizations.translateWithParams(
                      'screens.connection.labels.action_error',
                      {'error': e.toString()},
                    );
                  });
                }
              });
            }

            return Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (isLoading) ...[
                    const CircularProgressIndicator(),
                    const SizedBox(height: 16),
                    Text(
                      localizations.translate(
                          'screens.connection.labels.executing_action'),
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                  ],
                  if (!isLoading && statusMessage != null) ...[
                    Icon(
                      statusMessage!.contains('successfully')
                          ? Icons.check_circle_outline
                          : Icons.error_outline,
                      color: statusMessage!.contains('successfully')
                          ? theme.colorScheme.primary
                          : theme.colorScheme.error,
                      size: 48,
                    ),
                    const SizedBox(height: 16),
                    Text(
                      statusMessage!,
                      style: const TextStyle(fontWeight: FontWeight.bold),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: () => Navigator.pop(context),
                      child: Text(localizations
                          .translate('screens.connection.buttons.close')),
                    ),
                  ],
                ],
              ),
            );
          },
        );
      },
    );
  }

  void _showActionsBottomSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) {
        return DraggableScrollableSheet(
          expand: false,
          initialChildSize: 0.5,
          minChildSize: 0.3,
          maxChildSize: 0.8,
          builder: (context, scrollController) {
            return Column(
              children: [
                Container(
                  height: 5,
                  width: 50,
                  margin: const EdgeInsets.symmetric(vertical: 8),
                  decoration: BoxDecoration(
                    color: Colors.grey[300],
                    borderRadius: BorderRadius.circular(5),
                  ),
                ),
                Text(
                  localizations.translate('screens.connection.labels.actions'),
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
                Expanded(
                  child: ListView.separated(
                    controller: scrollController,
                    itemCount: widget.element.actions.length,
                    itemBuilder: (context, index) {
                      final action = widget.element.actions[index];
                      return ListTile(
                        leading: const Icon(Icons.http),
                        title: Text(
                          action.name,
                          overflow: TextOverflow.ellipsis,
                          maxLines: 1,
                        ),
                        subtitle: Text(
                          action.endpoint,
                          overflow: TextOverflow.ellipsis,
                          maxLines: 1,
                        ),
                        trailing: ElevatedButton(
                          onPressed: () => _performAction(action),
                          child: Text(localizations
                              .translate('screens.connection.buttons.run')),
                        ),
                      );
                    },
                    separatorBuilder: (context, index) =>
                        const Divider(height: 1),
                  ),
                ),
              ],
            );
          },
        );
      },
    );
  }

  void _initializeStream() {
    setState(() {
      isLoading = true;
      hasError = false;
      _reconnectAttempt++;
    });

    _pollStreamAvailability();
  }

  void _pollStreamAvailability() async {
    const pollingInterval = Duration(seconds: 5);
    while (true) {
      if (!mounted) break;

      try {
        final httpClient = HttpClient();
        final request = await httpClient.getUrl(Uri.parse(streamUrl));
        headers.forEach(request.headers.add);
        final response = await request.close();

        if (response.statusCode == 200) {
          print('Stream is online.');
          setState(() {
            hasError = false;
            isLoading = false;
          });
          return;
        } else {
          print('Stream not available. Status code: ${response.statusCode}');
        }
      } catch (e) {
        print('Error while checking stream availability: $e');
      }

      if (!mounted) break;
      await Future.delayed(pollingInterval);
    }
  }

  Future<void> _saveImage(Uint8List imageBytes) async {
    try {
      final directory = widget.appDirectory;
      final filePath =
          '${directory.path}/snapshot_${DateTime.now().millisecondsSinceEpoch}.png';
      final file = File(filePath);
      await file.writeAsBytes(imageBytes);
      print('Snapshot saved to $filePath');
    } catch (e) {
      print('Error saving image: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text(localizations.translate(
                  'screens.connection.labels.failed_to_save_snapshot'))),
        );
      }
    }
  }

  Future<void> _captureStreamFrame() async {
    try {
      RenderRepaintBoundary? boundary = _mjpegKey.currentContext
          ?.findRenderObject() as RenderRepaintBoundary?;

      if (boundary == null) {
        print('Mjpeg Boundary is null');
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text(localizations.translate(
                  'screens.connection.labels.failed_to_capture_snapshot'))),
        );
        return;
      }

      ui.Image image =
          await boundary.toImage(pixelRatio: View.of(context).devicePixelRatio);
      ByteData? byteData =
          await image.toByteData(format: ui.ImageByteFormat.png);

      if (byteData != null && mounted) {
        Uint8List streamFrame = byteData.buffer.asUint8List();
        await _saveImage(streamFrame);

        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
                content: Text(localizations
                    .translate('screens.connection.labels.captured_snapshot'))),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        print('Error capturing stream frame: $e');
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text(localizations.translate(
                  'screens.connection.labels.failed_to_capture_snapshot'))),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    String appBarTitle;
    if (isLoading) {
      appBarTitle = localizations.translateWithParams(
          'screens.connection.labels.connecting_to',
          {'name': widget.element.name});
    } else if (hasError) {
      appBarTitle =
          localizations.translate('screens.connection.labels.failed_to_load');
    } else {
      appBarTitle = localizations.translateWithParams(
          'screens.connection.labels.connected_to',
          {'name': widget.element.name});
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(appBarTitle),
      ),
      body: hasError
          ? Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.error_outline,
                      size: 64, color: theme.colorScheme.error),
                  SizedBox(height: 8),
                  Center(
                    child: Text(
                      localizations.translate(
                          'screens.connection.labels.failed_to_load'),
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 16,
                      ),
                    ),
                  ),
                  SizedBox(height: 8),
                  Text(
                    localizations
                        .translate('screens.connection.labels.retrying'),
                    style: TextStyle(fontSize: 14, color: Colors.grey),
                  ),
                  SizedBox(height: 32),
                  if (isLoading)
                    Padding(
                      padding: EdgeInsets.all(16.0),
                      child: CircularProgressIndicator(
                        valueColor: AlwaysStoppedAnimation<Color>(
                            theme.colorScheme.onSurface),
                      ),
                    ),
                ],
              ),
            )
          : InteractiveViewer(
              transformationController: _transformationController,
              panEnabled: true,
              scaleEnabled: true,
              minScale: 1.0,
              maxScale: 3.0,
              child: Center(
                child: RepaintBoundary(
                  key: _mjpegKey,
                  child: Mjpeg(
                    stream: streamUrl,
                    headers: headers,
                    isLive: !isPaused,
                    fit: BoxFit.contain,
                    key: Key('mjpeg_$_reconnectAttempt'),
                    error: (context, error, stackTrace) {
                      print('Mjpeg Stream Error: $error');
                      WidgetsBinding.instance.addPostFrameCallback((_) {
                        if (mounted) {
                          setState(() {
                            hasError = true;
                          });
                          _pollStreamAvailability();
                        }
                      });
                      return const SizedBox.shrink();
                    },
                  ),
                ),
              ),
            ),
      floatingActionButton: hasError
          ? null
          : SpeedDial(
              icon: Icons.menu,
              activeIcon: Icons.close,
              childPadding: const EdgeInsets.symmetric(vertical: 6),
              children: [
                SpeedDialChild(
                  labelBackgroundColor: Colors.transparent,
                  shape: const CircleBorder(),
                  child: const Icon(Icons.camera_alt),
                  onTap: _captureStreamFrame,
                ),
                SpeedDialChild(
                  labelBackgroundColor: Colors.transparent,
                  shape: const CircleBorder(),
                  child: Icon(isPaused ? Icons.play_arrow : Icons.pause),
                  onTap: () => setState(() => isPaused = !isPaused),
                ),
                if (widget.element.actions.isNotEmpty)
                  SpeedDialChild(
                    labelBackgroundColor: Colors.transparent,
                    shape: const CircleBorder(),
                    child: const Icon(Icons.list),
                    onTap: _showActionsBottomSheet,
                  ),
              ],
            ),
    );
  }
}
