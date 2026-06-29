---
title: "qbt bencode edit"
description: "edit bencode data"
editUrl: false
---

edit bencode data

### Synopsis

Edit bencode files like .fastresume. Shut down client and make a backup of data before.

```
qbt bencode edit [flags]
```

### Examples

```
  qbt bencode edit --dir /home/user/.local/share/qBittorrent/BT_backup --pattern '/home/user01/torrents' --replace '/home/test/torrents'
```

### Options

```
      --dir string       Dir with fast-resume files (required)
      --dry-run          Dry run, don't write changes
  -h, --help             help for edit
      --pattern string   Pattern to change (required)
      --replace string   Text to replace pattern with (required)
  -v, --verbose          Verbose output
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt bencode](../qbt_bencode/)	 - Bencode subcommand

