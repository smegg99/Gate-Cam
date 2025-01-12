// lib/models/stream_element.dart
// dart run build_runner build
import 'package:hive/hive.dart';

part 'stream_element.g.dart';

@HiveType(typeId: 0)
class StreamElement extends HiveObject {
  @HiveField(0)
  String name;

  @HiveField(1)
  String url;

  @HiveField(2)
  String username;

  @HiveField(3)
  String password;

  @HiveField(4)
  List<ActionElement> actions;

  StreamElement({
    required this.name,
    required this.url,
    required this.username,
    required this.password,
    List<ActionElement>? actions,
  }) : actions = actions ?? [];
}

@HiveType(typeId: 1)
class ActionElement {
  @HiveField(0)
  String name;

  @HiveField(1)
  String icon;

  @HiveField(2)
  String endpoint;

  @HiveField(3)
  String method;

  @HiveField(4)
  Map<String, String> headers;

  @HiveField(5)
  String? body;

  ActionElement({
    required this.name,
    required this.icon,
    required this.endpoint,
    required this.method,
    this.headers = const {},
  });
}
