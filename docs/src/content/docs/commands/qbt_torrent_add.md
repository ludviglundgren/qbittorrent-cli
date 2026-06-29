---
title: "qbt torrent add"
description: "Add torrent(s)"
editUrl: false
---

Add torrent(s)

### Synopsis

Add new torrent(s) to qBittorrent from file or magnet. Supports glob pattern for files like: ./files/*.torrent

```
qbt torrent add [flags]
```

### Examples

```
  qbt torrent add my-file.torrent --category test --tags tag1
  qbt torrent add ./files/*.torrent --paused --skip-hash-check
  qbt torrent add magnet:?xt=urn:btih:5dee65101db281ac9c46344cd6b175cdcad53426&dn=download
```

### Options

```
      --category string         Add torrent to the specified category
      --dry-run                 Run without doing anything
      --first-last-piece        Prioritize first and last pieces for preview
  -h, --help                    help for add
      --ignore-rules            Ignore rules from config
      --limit-dl uint           Set torrent download speed limit. Unit in bytes/second
      --limit-ul uint           Set torrent upload speed limit. Unit in bytes/second
      --paused                  Add torrent in paused state
      --recheck                 Force recheck after adding (useful when using --paused)
      --remove-stalled          Remove stalled torrents from re-announce
      --save-path string        Add torrent to the specified path
      --sequential              Download torrent pieces in sequential order
      --skip-hash-check         Skip hash check
      --sleep duration          Set the amount of time to wait between adding torrents in seconds (default 200ms)
      --stop-condition string   Add torrent with the specified stop condition. Possible values: None, MetadataReceived, FilesChecked. Example: --stop-condition MetadataReceived
      --tags stringArray        Add tags to torrent
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent](../qbt_torrent/)	 - Torrent subcommand

