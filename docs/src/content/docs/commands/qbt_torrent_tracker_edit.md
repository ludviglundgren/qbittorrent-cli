---
title: "qbt torrent tracker edit"
description: "Edit torrent tracker"
editUrl: false
---

Edit torrent tracker

### Synopsis

Edit tracker for torrents via hashes

```
qbt torrent tracker edit [flags]
```

### Examples

```
  qbt torrent tracker edit --old url.old/test --new url.com/test
```

### Options

```
      --dry-run      Run without doing anything
  -h, --help         help for edit
      --new string   New tracker URL
      --old string   Old tracker URL to replace
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent tracker](../qbt_torrent_tracker/)	 - Torrent tracker subcommand

