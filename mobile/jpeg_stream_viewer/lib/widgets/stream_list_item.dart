import 'package:flutter/material.dart';
import 'package:jpegsv/models/stream_element.dart';

class StreamListItem extends StatelessWidget {
  final StreamElement element;
  final bool isSelected;
  final bool isSelectionMode;
  final VoidCallback onEdit;
  final VoidCallback onConnect;
  final VoidCallback onLongPress;
  final ValueChanged<bool?> onCheckboxChanged;

  const StreamListItem({
    super.key,
    required this.element,
    required this.isSelected,
    required this.isSelectionMode,
    required this.onEdit,
    required this.onConnect,
    required this.onLongPress,
    required this.onCheckboxChanged,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onLongPress: onLongPress,
      onDoubleTap: onConnect,
      child: Container(
        color: isSelected ? Theme.of(context).colorScheme.secondary.withAlpha((0.2 * 255).toInt()) : null,
        padding: const EdgeInsets.symmetric(vertical: 8.0, horizontal: 16.0),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(element.name,
                      style: Theme.of(context).textTheme.titleMedium),
                  Text(element.url,
                      style: Theme.of(context).textTheme.labelSmall),
                ],
              ),
            ),
            if (isSelectionMode)
              Checkbox(value: isSelected, onChanged: onCheckboxChanged)
            else
              Row(
                children: [
                  IconButton(
                    icon: Icon(Icons.edit, color: Theme.of(context).colorScheme.secondary),
                    onPressed: onEdit,
                  ),
                  IconButton(
                    icon: Icon(Icons.link, color: Theme.of(context).primaryColor),
                    onPressed: onConnect,
                  ),
                ],
              ),
          ],
        ),
      ),
    );
  }
}
