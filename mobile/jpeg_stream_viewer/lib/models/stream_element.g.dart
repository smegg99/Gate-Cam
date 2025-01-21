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
      order: fields[1] as int,
      url: fields[2] as String,
      username: fields[3] as String,
      password: fields[4] as String,
      actions: (fields[5] as List?)?.cast<ActionElement>(),
    );
  }

  @override
  void write(BinaryWriter writer, StreamElement obj) {
    writer
      ..writeByte(6)
      ..writeByte(0)
      ..write(obj.name)
      ..writeByte(1)
      ..write(obj.order)
      ..writeByte(2)
      ..write(obj.url)
      ..writeByte(3)
      ..write(obj.username)
      ..writeByte(4)
      ..write(obj.password)
      ..writeByte(5)
      ..write(obj.actions);
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

class ActionElementAdapter extends TypeAdapter<ActionElement> {
  @override
  final int typeId = 1;

  @override
  ActionElement read(BinaryReader reader) {
    final numOfFields = reader.readByte();
    final fields = <int, dynamic>{
      for (int i = 0; i < numOfFields; i++) reader.readByte(): reader.read(),
    };
    return ActionElement(
      name: fields[0] as String,
      order: fields[1] as int,
      endpoint: fields[2] as String,
      method: fields[3] as String,
      headers: (fields[4] as Map).cast<String, String>(),
      body: fields[5] as String?,
    );
  }

  @override
  void write(BinaryWriter writer, ActionElement obj) {
    writer
      ..writeByte(6)
      ..writeByte(0)
      ..write(obj.name)
      ..writeByte(1)
      ..write(obj.order)
      ..writeByte(2)
      ..write(obj.endpoint)
      ..writeByte(3)
      ..write(obj.method)
      ..writeByte(4)
      ..write(obj.headers)
      ..writeByte(5)
      ..write(obj.body);
  }

  @override
  int get hashCode => typeId.hashCode;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is ActionElementAdapter &&
          runtimeType == other.runtimeType &&
          typeId == other.typeId;
}
