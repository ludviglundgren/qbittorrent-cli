---
title: "qbt torrent remove"
description: "Removes specified torrent(s)"
editUrl: false
---

Removes specified torrent(s)

### Synopsis

Removes torrents indicated by hash, name or a prefix of either. Whitespace indicates next prefix unless argument is surrounded by quotes

```
qbt torrent remove [flags]
```

### Options

```
      --all                        Removes all torrents
      --delete-files               Also delete downloaded files from torrent(s)
      --dry-run                    Display what would be done without actually doing it
      --exclude-tags strings       Exclude torrents with provided tags
  -f, --filter string              Filter by state: all, active, paused, completed, stalled, errored
      --hashes strings             Add hashes as comma separated list
  -h, --help                       help for remove
  -c, --include-category strings   Remove torrents from these categories. Comma separated
      --include-tags strings       Include torrents with provided tags
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent](../qbt_torrent/)	 - Torrent subcommand

