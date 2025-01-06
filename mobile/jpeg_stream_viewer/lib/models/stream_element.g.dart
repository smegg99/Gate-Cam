// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'stream_element.dart';

// **************************************************************************
// TypeAdapterGenerator
// **************************************************************************

class StreamElementAdapter extends TypeAdapter<StreamElement> {
  @override
  final int typeId = 0;

  @override
  StreamElement read(BinaryReader reader) {
    final numOfFields = reader.readByte();
    final fields = <int, dynamic>{
      for (int i = 0; i < numOfFields; i++) reader.readByte(): reader.read(),
    };
    return StreamElement(
      name: fields[0] as String,
      url: fields[1] as String,
      username: fields[2] as String,
      password: fields[3] as String,
    );
  }

  @override
  void write(BinaryWriter writer, StreamElement obj) {
    writer
      ..writeByte(4)
      ..writeByte(0)
      ..write(obj.name)
      ..writeByte(1)
      ..write(obj.url)
      ..writeByte(2)
      ..write(obj.username)
      ..writeByte(3)
      ..write(obj.password);
  }

  @override
  int get hashCode => typeId.hashCode;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is StreamElementAdapter &&
          runtimeType == other.runtimeType &&
          typeId == other.typeId;
}
