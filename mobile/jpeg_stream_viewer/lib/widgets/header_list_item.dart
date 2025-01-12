import 'package:flutter/material.dart';
import 'package:jpegsv/localization/localization.dart';

class HeaderListItem extends StatefulWidget {
  final TextEditingController keyController;
  final TextEditingController valueController;
  final VoidCallback onDelete;
  final VoidCallback onChanged;

  const HeaderListItem({
    super.key,
    required this.keyController,
    required this.valueController,
    required this.onDelete,
    required this.onChanged,
  });

  @override
  State<HeaderListItem> createState() => _HeaderListItemState();
}

class _HeaderListItemState extends State<HeaderListItem> {
  AppLocalizations get localizations => AppLocalizations.of(context);

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Container(
      padding: const EdgeInsets.all(8.0),
      child: Row(
        children: [
          Expanded(
            flex: 1,
            child: TextField(
              controller: widget.keyController,
              decoration: InputDecoration(
                labelText: localizations
                    .translate('screens.action_edit.labels.header_key'),
              ),
              onChanged: (_) => widget.onChanged(),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            flex: 1,
            child: TextField(
              controller: widget.valueController,
              decoration: InputDecoration(
                labelText: localizations
                    .translate('screens.action_edit.labels.header_value'),
              ),
              onChanged: (_) => widget.onChanged(),
            ),
          ),
          const SizedBox(width: 8),
          IconButton(
            icon: Icon(Icons.close, color: theme.colorScheme.error),
            onPressed: widget.onDelete,
          ),
        ],
      ),
    );
  }
}
