import 'package:flutter/material.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:jpegsv/models/stream_element.dart';
import 'package:jpegsv/widgets/stream_list_item.dart';

class MainScreen extends StatefulWidget {
  const MainScreen({super.key});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> {
  bool _isSelectionMode = false;
  final Set<int> _selectedIndices = {};

  void _handleEdit(StreamElement element) {
    Navigator.pushNamed(context, '/edit', arguments: element);
  }

  void _handleCreate() {
    Navigator.pushNamed(context, '/edit', arguments: null);
  }

  void _handleConnect(StreamElement element) {
    Navigator.pushNamed(context, '/connect', arguments: element);
  }

  void _toggleSelectionMode(int index) {
    setState(() {
      _isSelectionMode = true;
      _selectedIndices.add(index);
    });
  }

  void _toggleSelectItem(int index) {
    setState(() {
      if (_selectedIndices.contains(index)) {
        _selectedIndices.remove(index);
        if (_selectedIndices.isEmpty) {
          _isSelectionMode = false;
        }
      } else {
        _selectedIndices.add(index);
      }
    });
  }

  Future<void> _deleteSelectedItems(Box<StreamElement> box) async {
    final elementsToDelete =
        _selectedIndices.map((index) => box.getAt(index)!).toList();

    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Selected Items'),
        content: Text(
            'Are you sure you want to delete ${_selectedIndices.length} items?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirm == true) {
      for (var element in elementsToDelete) {
        await element.delete();
      }

      setState(() {
        _isSelectionMode = false;
        _selectedIndices.clear();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_isSelectionMode
            ? '${_selectedIndices.length} Selected'
            : 'Stream Manager'),
        actions: _isSelectionMode
            ? [
                IconButton(
                  icon: const Icon(Icons.delete),
                  onPressed: () async {
                    final box = Hive.box<StreamElement>('stream_elements');
                    await _deleteSelectedItems(box);
                  },
                ),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () {
                    setState(() {
                      _isSelectionMode = false;
                      _selectedIndices.clear();
                    });
                  },
                ),
              ]
            : null,
      ),
      body: ValueListenableBuilder<Box<StreamElement>>(
        valueListenable:
            Hive.box<StreamElement>('stream_elements').listenable(),
        builder: (context, box, _) {
          final elements = box.values.toList().cast<StreamElement>();

          if (elements.isEmpty) {
            return const Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.image_not_supported, size: 64, color: Colors.grey),
                  SizedBox(height: 16),
                  Text('No streams available'),
                ],
              ),
            );
          }

          return ListView.builder(
            itemCount: elements.length,
            itemBuilder: (context, index) {
              final element = elements[index];
              final isSelected = _selectedIndices.contains(index);

              return StreamListItem(
                element: element,
                isSelected: isSelected,
                isSelectionMode: _isSelectionMode,
                onEdit: () => _handleEdit(element),
                onConnect: () => _handleConnect(element),
                onLongPress: () => _toggleSelectionMode(index),
                onCheckboxChanged: (value) => _toggleSelectItem(index),
              );
            },
          );
        },
      ),
      floatingActionButton: _isSelectionMode
          ? null
          : FloatingActionButton(
              onPressed: _handleCreate,
              child: const Icon(Icons.add),
            ),
    );
  }
}
