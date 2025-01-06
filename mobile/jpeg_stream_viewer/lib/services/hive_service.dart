import 'package:hive/hive.dart';
import 'package:jpegsv/models/stream_element.dart';

class HiveService {
  static const _boxName = 'stream_elements';

  Future<Box<StreamElement>> openBox() async {
    return await Hive.openBox<StreamElement>(_boxName);
  }

  Future<void> addElement(StreamElement element) async {
    final box = await openBox();
    await box.add(element);
  }

  Future<void> deleteElement(StreamElement element) async {
    final box = await openBox();
    await box.delete(element.key);
  }

  Future<void> updateElement(StreamElement element) async {
    await element.save();
  }

  Future<List<StreamElement>> getAllElements() async {
    final box = await openBox();
    return box.values.toList();
  }
}
