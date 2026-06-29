---
title: "qbt torrent tag issues"
description: "tag torrents with issues"
editUrl: false
---

tag torrents with issues

### Synopsis

Tag torrents that may have broken trackers or be unregistered

```
qbt torrent tag issues [flags]
```

### Examples

```
  qbt torrent tag issues --unregistered --not-working
```

### Options

```
      --dry-run        Dry run, do not tag torrents
  -h, --help           help for issues
      --not-working    tag not working torrents
      --unregistered   tag unregistered
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent tag](../qbt_torrent_tag/)	 - Torrent tag subcommand

