import 'package:flutter/material.dart';
import 'package:flutter_speed_dial/flutter_speed_dial.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:jpegsv/localization/localization.dart';
import 'package:jpegsv/models/stream_element.dart';
import 'package:jpegsv/widgets/stream_list_item.dart';

class MainScreen extends StatefulWidget {
  const MainScreen({super.key});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> {
  bool _isSelectionMode = false;

  AppLocalizations get localizations => AppLocalizations.of(context);
  ThemeData get theme => Theme.of(context);

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

  void _handleSettings() {
    Navigator.pushNamed(context, '/settings');
  }

  void _handleAbout() {
    Navigator.pushNamed(context, '/about');
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
        title: Text(
            localizations.translate('screens.home.labels.delete_selected')),
        content: Text(localizations.translateWithParams(
            'screens.home.labels.delete_selected_confirmation',
            {'amount': _selectedIndices.length.toString()})),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(localizations.translate('screens.home.labels.cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(localizations.translate('screens.home.labels.delete')),
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
            ? localizations.translateWithParams('screens.home.labels.selected',
                {'amount': _selectedIndices.length.toString()})
            : localizations.translate('screens.home.title')),
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
          elements.sort((a, b) => a.order.compareTo(b.order));

          if (elements.isEmpty) {
            return Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.image_not_supported, size: 64, color: Colors.grey),
                  const SizedBox(height: 16),
                  Text(localizations
                      .translate('screens.home.labels.no_connections')),
                ],
              ),
            );
          }

          return ReorderableListView(
            buildDefaultDragHandles: false,
            onReorder: (oldIndex, newIndex) async {
              setState(() {
                if (newIndex > oldIndex) newIndex--;
                final element = elements.removeAt(oldIndex);
                elements.insert(newIndex, element);

                // Update order in memory
                for (int i = 0; i < elements.length; i++) {
                  elements[i].order = i;
                }
              });

              // Persist updated order to Hive after reordering
              Future.delayed(const Duration(milliseconds: 300), () async {
                for (final element in elements) {
                  await element.save();
                }
              });
            },
            children: List.generate(
              elements.length,
              (index) {
                final element = elements[index];
                final isSelected = _selectedIndices.contains(index);

                return StreamListItem(
                  key: ValueKey(
                      element.key ?? element.name ?? index), // Unique key
                  element: element,
                  isSelected: isSelected,
                  isSelectionMode: _isSelectionMode,
                  onTap: () => {if (_isSelectionMode) _toggleSelectItem(index)},
                  onEdit: () => _handleEdit(element),
                  onConnect: () => _handleConnect(element),
                  onLongPress: () => _toggleSelectionMode(index),
                  onCheckboxChanged: (value) => _toggleSelectItem(index),
                );
              },
            ),
          );
        },
      ),
      floatingActionButton: _isSelectionMode
          ? null
          : SpeedDial(
              gradient: null,
              icon: Icons.menu,
              activeIcon: Icons.close,
              childPadding: const EdgeInsets.symmetric(vertical: 6),
              children: [
                SpeedDialChild(
                  labelBackgroundColor: Colors.transparent,
                  labelShadow: null,
                  shape: const CircleBorder(),
                  child: const Icon(Icons.add),
                  onTap: _handleCreate,
                ),
                SpeedDialChild(
                  labelBackgroundColor: Colors.transparent,
                  labelShadow: null,
                  shape: const CircleBorder(),
                  child: const Icon(Icons.settings),
                  onTap: _handleSettings,
                ),
                SpeedDialChild(
                  labelBackgroundColor: Colors.transparent,
                  labelShadow: null,
                  shape: const CircleBorder(),
                  child: const Icon(Icons.info),
                  onTap: _handleAbout,
                ),
              ],
            ),
    );
  }
}
