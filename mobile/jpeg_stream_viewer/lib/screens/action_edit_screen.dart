import 'package:flutter/material.dart';
import 'package:jpegsv/localization/localization.dart';
import 'package:jpegsv/models/stream_element.dart';
import 'package:jpegsv/widgets/header_list_item.dart';

class ActionEditScreen extends StatefulWidget {
  final ActionElement action;

  const ActionEditScreen({super.key, required this.action});

  @override
  State<ActionEditScreen> createState() => _ActionEditScreenState();
}

class _ActionEditScreenState extends State<ActionEditScreen> {
  late TextEditingController _nameController;
  late TextEditingController _endpointController;
  late TextEditingController _bodyController;
  late String _method;
  late Map<String, TextEditingController> _keyControllers;
  late Map<String, TextEditingController> _valueControllers;
  final ScrollController _scrollController = ScrollController();
  final bool _isSelectionMode = false;

  AppLocalizations get localizations => AppLocalizations.of(context);
  ThemeData get theme => Theme.of(context);

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.action.name);
    _endpointController = TextEditingController(text: widget.action.endpoint);
    _bodyController = TextEditingController(
        text: widget.action.body ?? '');
    _method = widget.action.method;

    _keyControllers = {};
    _valueControllers = {};
    for (var entry in widget.action.headers.entries) {
      _keyControllers[entry.key] = TextEditingController(text: entry.key);
      _valueControllers[entry.key] = TextEditingController(text: entry.value);
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    _endpointController.dispose();
    _bodyController.dispose();
    for (var controller in _keyControllers.values) {
      controller.dispose();
    }
    for (var controller in _valueControllers.values) {
      controller.dispose();
    }
    _scrollController.dispose();
    super.dispose();
  }

  void _addEmptyHeader() {
    setState(() {
      final newKey = localizations.translateWithParams(
          'screens.action_edit.labels.new_header',
          {'index': (_keyControllers.length + 1).toString()});
      _keyControllers[newKey] = TextEditingController(text: newKey);
      _valueControllers[newKey] = TextEditingController();
    });
  }

  void _saveChanges() {
    if (_nameController.text.isEmpty ||
      _endpointController.text.isEmpty ||
      _method.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(localizations.translate(
          'screens.action_edit.labels.please_fill_needed_fields'))),
      );
      return;
    }

    widget.action.name = _nameController.text;
    widget.action.endpoint = _endpointController.text;
    widget.action.body = _bodyController.text;
    widget.action.method = _method;

    widget.action.headers = _keyControllers.map((key, keyController) {
      final valueController = _valueControllers[key]!;
      return MapEntry(keyController.text, valueController.text);
    });

    Navigator.pop(context, widget.action);
  }

  Widget _buildHeaderEntry(String key) {
    final keyController = _keyControllers[key]!;
    final valueController = _valueControllers[key]!;

    return HeaderListItem(
      keyController: keyController,
      valueController: valueController,
      onChanged: () {},
      onDelete: () {
        setState(() {
          _keyControllers.remove(key);
          _valueControllers.remove(key);
        });
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(localizations.translate('screens.action_edit.title')),
        actions: [
          if (!_isSelectionMode)
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: IconButton(
                icon: const Icon(Icons.save),
                tooltip:
                    localizations.translate('screens.action_edit.buttons.save'),
                onPressed: _saveChanges,
              ),
            ),
        ],
      ),
      body: ListView(
        controller: _scrollController,
        padding:
            const EdgeInsets.only(left: 16, right: 16, top: 16, bottom: 84),
        children: [
          Padding(
            padding: const EdgeInsets.only(bottom: 16.0),
            child: TextField(
              controller: _nameController,
              decoration: InputDecoration(
                labelText: '${localizations.translate('screens.action_edit.labels.action_name')} *',
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(bottom: 16.0),
            child: TextField(
              controller: _endpointController,
              decoration: InputDecoration(
                labelText: '${localizations.translate('screens.action_edit.labels.endpoint')} *',
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(bottom: 16.0),
            child: Row(
              children: [
                Expanded(
                  child: DropdownMenu<String>(
                    initialSelection: _method,
                    label: Text(
                      '${localizations.translate('screens.action_edit.labels.method')} *',
                    ),
                    dropdownMenuEntries:
                        ['GET', 'POST', 'PUT', 'DELETE'].map((method) {
                      return DropdownMenuEntry(
                        value: method,
                        label: method,
                      );
                    }).toList(),
                    onSelected: (value) {
                      if (value != null) {
                        setState(() {
                          _method = value;
                        });
                      }
                    },
                  ),
                ),
                // const SizedBox(width: 16),
                // Expanded(
                //   child: TextField(
                //     controller: _iconController,
                //     decoration: InputDecoration(
                //       labelText: localizations
                //           .translate('screens.action_edit.labels.icon'),
                //     ),
                //   ),
                // ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(bottom: 16.0),
            child: TextField(
              controller: _bodyController,
              maxLines: 5,
              decoration: InputDecoration(
                labelText:
                    localizations.translate('screens.action_edit.labels.body'),
                alignLabelWithHint: true,
                border: OutlineInputBorder(),
              ),
            ),
          ),
          const SizedBox(height: 24),
          if (_keyControllers.isEmpty)
            Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.playlist_remove, size: 64, color: Colors.grey),
                  const SizedBox(height: 16),
                  Text(localizations.translate(
                      'screens.action_edit.labels.no_headers_available')),
                ],
              ),
            )
          else ...[
            Align(
              alignment: Alignment.centerLeft,
              child: Text(
                localizations.translate('screens.action_edit.labels.headers'),
                style:
                    const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
            ),
            const Divider(),
            ListView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: _keyControllers.keys.length,
              itemBuilder: (context, index) {
                final key = _keyControllers.keys.elementAt(index);
                return _buildHeaderEntry(key);
              },
            ),
          ],
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _addEmptyHeader,
        icon: const Icon(Icons.add),
        label: Text(
            localizations.translate('screens.action_edit.buttons.add_header')),
      ),
    );
  }
}
