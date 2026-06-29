---
title: "qbt torrent pause"
description: "Pause specified torrent(s)"
editUrl: false
---

Pause specified torrent(s)

### Synopsis

Pause the torrent(s) indicated by the supplied hash(es), or pause every torrent with --all.

```
qbt torrent pause [flags]
```

### Examples

```
  qbt torrent pause --all
  qbt torrent pause HASH1 HASH2
  qbt torrent pause --hashes HASH1,HASH2
```

### Options

```
      --all              Pauses all torrents
      --hashes strings   Add hashes as comma separated list
  -h, --help             help for pause
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent](../qbt_torrent/)	 - Torrent subcommand

