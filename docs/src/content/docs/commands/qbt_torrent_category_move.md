---
title: "qbt torrent category move"
description: "move torrents between categories"
editUrl: false
---

move torrents between categories

### Synopsis

Move torrents from one category to another

```
qbt torrent category move [flags]
```

### Examples

```
  qbt torrent category move --from cat1 --to cat2
```

### Options

```
      --dry-run                Run without doing anything
      --exclude-tags strings   Exclude torrents with provided tags
      --from strings           Move from categories (required)
  -h, --help                   help for move
      --include-tags strings   Include torrents with provided tags
      --min-seed-time int      Minimum seed time in MINUTES before moving.
      --to string              Move to the specified category (required)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -q, --quiet           suppress output
```

### SEE ALSO

* [qbt torrent category](../qbt_torrent_category/)	 - Torrent category subcommand

