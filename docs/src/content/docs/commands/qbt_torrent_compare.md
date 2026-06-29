---
title: "qbt torrent compare"
description: "Compare torrents"
editUrl: false
---

Compare torrents

### Synopsis

Compare torrents between clients

```
qbt torrent compare [flags]
```

### Examples

```
  qbt torrent compare --addr http://localhost:10000 --user u --pass p --compare-addr http://url.com:10000 --compare-user u --compare-pass p
```

### Options

```
      --api-key string              Source api key
      --basic-pass string           Source basic auth pass
      --basic-user string           Source basic auth user
      --compare-api-key string      Secondary api key
      --compare-basic-pass string   Secondary basic auth pass
      --compare-basic-user string   Secondary basic auth user
      --compare-host string         Secondary host
      --compare-pass string         Secondary pass
      --compare-user string         Secondary user
      --dry-run                     dry run
  -h, --help                        help for compare
      --host string                 Source host
      --pass string                 Source pass
      --tag string                  set a custom tag for duplicates on compare. default: compare-dupe (default "compare-dupe")
      --tag-duplicates              tag duplicates on compare
      --user string                 Source user
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent](../qbt_torrent/)	 - Torrent subcommand

