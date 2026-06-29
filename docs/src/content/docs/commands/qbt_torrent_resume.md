---
title: "qbt torrent resume"
description: "Resume specified torrent(s)"
editUrl: false
---

Resume specified torrent(s)

### Synopsis

Resume the torrent(s) indicated by the supplied hash(es), or resume every torrent with --all.

```
qbt torrent resume [flags]
```

### Examples

```
  qbt torrent resume --all
  qbt torrent resume HASH1 HASH2
  qbt torrent resume --hashes HASH1,HASH2
```

### Options

```
      --all              resumes all torrents
      --hashes strings   Add hashes as comma separated list
  -h, --help             help for resume
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent](../qbt_torrent/)	 - Torrent subcommand

