---
title: "qbt torrent import"
description: "Import torrents"
editUrl: false
---

Import torrents

### Synopsis

Import torrents with state from other clients [rtorrent, deluge]

```
qbt torrent import {rtorrent | deluge} --source-dir dir --qbit-dir dir2 [--skip-backup] [--dry-run] [flags]
```

### Examples

```
  qbt torrent import deluge --source-dir ~/.config/deluge/state/ --qbit-dir ~/.local/share/data/qBittorrent/BT_backup --dry-run
  qbt torrent import rtorrent --source-dir ~/.sessions --qbit-dir ~/.local/share/data/qBittorrent/BT_backup --dry-run
```

### Options

```
      --dry-run             Run without importing anything
  -h, --help                help for import
      --qbit-dir string     qBittorrent BT_backup dir. Commonly ~/.local/share/qBittorrent/BT_backup (required)
      --skip-backup         Skip backup before import
      --source-dir string   source client state dir (required)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent](../qbt_torrent/)	 - Torrent subcommand

