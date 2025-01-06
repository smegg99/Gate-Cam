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

  StreamElement({
    required this.name,
    required this.url,
    required this.username,
    required this.password,
  });
}
