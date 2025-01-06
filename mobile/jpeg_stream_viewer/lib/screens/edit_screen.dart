import 'package:flutter/material.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:jpegsv/models/stream_element.dart';

class EditScreen extends StatefulWidget {
  final StreamElement? element;

  const EditScreen({super.key, this.element});

  @override
  State<EditScreen> createState() => _EditScreenState();
}

class _EditScreenState extends State<EditScreen> {
  late TextEditingController _nameController;
  late TextEditingController _urlController;
  late TextEditingController _usernameController;
  late TextEditingController _passwordController;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.element?.name ?? '');
    _urlController = TextEditingController(text: widget.element?.url ?? '');
    _usernameController =
        TextEditingController(text: widget.element?.username ?? '');
    _passwordController =
        TextEditingController(text: widget.element?.password ?? '');
  }

  @override
  void dispose() {
    _nameController.dispose();
    _urlController.dispose();
    _usernameController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _saveChanges() async {
    if (_nameController.text.isEmpty ||
        _urlController.text.isEmpty ||
        _usernameController.text.isEmpty ||
        _passwordController.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please fill all fields')),
      );
      return;
    }

    final box = Hive.box<StreamElement>('stream_elements');

    if (widget.element == null) {
      // Create new element
      final newElement = StreamElement(
        name: _nameController.text,
        url: _urlController.text,
        username: _usernameController.text,
        password: _passwordController.text,
      );
      await box.add(newElement);
    } else {
      // Update existing element
      widget.element!.name = _nameController.text;
      widget.element!.url = _urlController.text;
      widget.element!.username = _usernameController.text;
      widget.element!.password = _passwordController.text;
      await widget.element!.save();
    }

    if (mounted) {
      Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
            widget.element == null ? 'Create Connection' : 'Edit Connection'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            TextField(
              controller: _nameController,
              decoration: const InputDecoration(labelText: 'Connection Name'),
            ),
            TextField(
              controller: _urlController,
              decoration: const InputDecoration(labelText: 'URL'),
            ),
            TextField(
              controller: _usernameController,
              decoration: const InputDecoration(labelText: 'Username'),
            ),
            TextField(
              controller: _passwordController,
              decoration: const InputDecoration(labelText: 'Password'),
              obscureText: true,
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _saveChanges,
              child: Text(widget.element == null ? 'Create' : 'Save Changes'),
            ),
          ],
        ),
      ),
    );
  }
}
