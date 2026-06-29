---
title: "qbt torrent share-limit set"
description: "Set torrent share limits"
editUrl: false
---

Set torrent share limits

### Synopsis

Set share limits (ratio, seeding time, inactive seeding time) for torrents.

Limit values use qBittorrent's special semantics:
  -2   use the global limit (default)
  -1   no limit (unlimited)
  >=0  a specific value (ratio, or minutes for the seeding time limits)

qBittorrent applies all three limits in a single request, so any limit you do not
set is reset to the global limit (-2). The applied values are printed when the
command runs.

Note: on qBittorrent 5.x (Web API 2.12+) this also resets each torrent's share
limit action and mode to their defaults.

```
qbt torrent share-limit set [flags]
```

### Examples

```
  qbt torrent share-limit set --hashes hash1,hash2 --ratio 2.0
  qbt torrent share-limit set --all --seeding-time 1440
  qbt torrent share-limit set --include-category movies --ratio 1.5 --seeding-time 10080
  qbt torrent share-limit set --hashes hash1 --ratio -1
```

### Options

```
      --all                         Set share limits for all torrents
      --dry-run                     Run without doing anything
      --exclude-tags strings        Exclude torrents with any of these tags. Comma separated
      --hashes strings              Torrent hashes, as comma separated list
  -h, --help                        help for set
      --inactive-seeding-time int   Inactive seeding time limit in MINUTES. -2 = global, -1 = unlimited, >=0 = minutes (default -2)
  -c, --include-category strings    Set share limits for torrents in these categories. Comma separated
      --include-tags strings        Include torrents with any of these tags. Comma separated
      --ratio float                 Ratio limit. -2 = global, -1 = unlimited, >=0 = ratio (default -2)
      --seeding-time int            Seeding time limit in MINUTES. -2 = global, -1 = unlimited, >=0 = minutes (default -2)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent share-limit](../qbt_torrent_share-limit/)	 - Torrent share limit subcommand

