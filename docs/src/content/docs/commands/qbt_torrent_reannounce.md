---
title: "qbt torrent reannounce"
description: "Reannounce torrent(s)"
editUrl: false
---

Reannounce torrent(s)

### Synopsis

Reannounce torrents with non-OK tracker status.

```
qbt torrent reannounce [flags]
```

### Options

```
      --attempts int      Reannounce torrents X times (default 50)
      --category string   Reannounce torrents with category
      --dry-run           Run without doing anything
      --hash string       Reannounce torrent with hash
  -h, --help              help for reannounce
      --interval int      Reannounce torrents X times with interval Y. In MS (default 7000)
      --tag string        Reannounce torrents with tag
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent](../qbt_torrent/)	 - Torrent subcommand

