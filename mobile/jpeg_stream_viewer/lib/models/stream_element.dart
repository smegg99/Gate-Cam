// lib/models/stream_element.dart
// dart run build_runner build
import 'package:hive/hive.dart';

part 'stream_element.g.dart';

@HiveType(typeId: 0)
class StreamElement extends HiveObject {
  @HiveField(0)
  String name;

  @HiveField(1)
  int order;

  @HiveField(2)
  String url;

  @HiveField(3)
  String username;

  @HiveField(4)
  String password;

  @HiveField(5)
  List<ActionElement> actions;

  StreamElement({
    required this.name,
    required this.order,
    required this.url,
    required this.username,
    required this.password,
    List<ActionElement>? actions,
  }) : actions = actions ?? [];

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'order': order,
      'url': url,
      'username': username,
      'password': password,
      'actions': actions.map((action) => action.toJson()).toList(),
    };
  }

  factory StreamElement.fromJson(Map<String, dynamic> json) {
    return StreamElement(
      name: json['name'] ?? '',
      order: json['order'] ?? 0,
      url: json['url'] ?? '',
      username: json['username'] ?? '',
      password: json['password'] ?? '',
      actions: (json['actions'] as List<dynamic>?)
              ?.map((action) => ActionElement.fromJson(action))
              .toList() ??
          [],
    );
  }
}

@HiveType(typeId: 1)
class ActionElement {
  @HiveField(0)
  String name;

  @HiveField(1)
  int order;

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
    required this.order,
    required this.endpoint,
    required this.method,
    this.headers = const {},
    this.body,
  });

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'order': order,
      'endpoint': endpoint,
      'method': method,
      'headers': headers,
      'body': body,
    };
  }

  factory ActionElement.fromJson(Map<String, dynamic> json) {
    return ActionElement(
      name: json['name'],
      order: json['order'],
      endpoint: json['endpoint'],
      method: json['method'],
      headers: Map<String, String>.from(json['headers']),
      body: json['body'],
    );
  }
}
