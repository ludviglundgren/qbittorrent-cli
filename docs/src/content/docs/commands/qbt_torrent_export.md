---
title: "qbt torrent export"
description: "Export torrents"
editUrl: false
---

Export torrents

### Synopsis

Export torrents and fastresume by category

```
qbt torrent export [flags]
```

### Examples

```
  qbt torrent export --source ~/.local/share/data/qBittorrent/BT_backup --export-dir ~/qbt-backup --include-category=movies,tv
```

### Options

```
  -a, --archive                    archive export dir to .tar.gz
      --dry-run                    dry run
      --exclude-category strings   Exclude categories. Comma separated
      --exclude-tag strings        Exclude tags. Comma separated
      --export-dir string          Dir to export files to (required)
  -h, --help                       help for export
      --include-category strings   Export torrents from these categories. Comma separated
      --include-tag strings        Include tags. Comma separated
      --skip-manifest              Do not export all used tags and categories into manifest
      --source string              Dir with torrent and fast-resume files (required)
  -v, --verbose                    verbose output
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent](../qbt_torrent/)	 - Torrent subcommand

