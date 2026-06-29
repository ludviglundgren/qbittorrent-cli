---
title: "qbt torrent recheck"
description: "Recheck specified torrent(s)"
editUrl: false
---

Recheck specified torrent(s)

### Synopsis

Rechecks torrents indicated by hash(es).

```
qbt torrent recheck [flags]
```

### Examples

```
  qbt torrent recheck --hashes HASH
  qbt torrent recheck --hashes HASH1,HASH2

```

### Options

```
      --hashes strings   Add hashes as comma separated list
  -h, --help             help for recheck
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent](../qbt_torrent/)	 - Torrent subcommand

