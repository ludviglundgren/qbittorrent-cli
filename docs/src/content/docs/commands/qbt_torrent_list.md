---
title: "qbt torrent list"
description: "List torrents"
editUrl: false
---

List torrents

### Synopsis

List all torrents, or torrents with a specific filters. Get by filter, category, tag and hashes. Can be combined

```
qbt torrent list [flags]
```

### Examples

```
qbt torrent list --filter=downloading --category=linux-iso
```

### Options

```
  -c, --category string   Filter by category. All categories by default.
  -f, --filter string     Filter by state. Available filters: all, downloading, seeding, completed, paused, active, inactive, resumed, 
                          stalled, stalled_uploading, stalled_downloading, errored (default "all")
      --hashes strings    Filter by hashes. Separated by comma: "hash1,hash2".
  -h, --help              help for list
      --output string     Print as [formatted text (default), json]
  -t, --tag string        Filter by tag. Single tag: tag1
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent](../qbt_torrent/)	 - Torrent subcommand

