---
title: Configuration
description: Configure qbittorrent-cli with a .qbt.toml file - connection, reannounce, rules, add defaults and compare instances.
---

`qbt` is configured with a TOML file. Most commands need at least the
`[qbittorrent]` connection block.

## Config file location

Create `.qbt.toml` in `~/.config/qbt/`:

```shell
mkdir -p ~/.config/qbt && touch ~/.config/qbt/.qbt.toml
```

`qbt` looks for a file named `.qbt.toml` in the following locations, in order:

1. the current working directory
2. your home directory (`$HOME`)
3. `$HOME/.config/qbt`

You can also point at a specific file with the global `--config` flag:

```shell
qbt --config /path/to/.qbt.toml torrent list
```

## Minimal configuration

```toml
[qbittorrent]
addr     = "http://127.0.0.1:6776" # qbittorrent webui-api hostname/ip
login    = "user"                  # qbittorrent webui-api user     (optional)
password = "password"              # qbittorrent webui-api password (optional)
```

## Connection - `[qbittorrent]`

| Key         | Description                                                        |
| ----------- | ------------------------------------------------------------------ |
| `addr`      | qBittorrent Web API URL, including scheme and port                 |
| `login`     | Web API username (optional)                                        |
| `password`  | Web API password (optional)                                        |
| `apikey`    | Web API key - qBittorrent 5.2.x and newer only (optional)          |
| `basicUser` | HTTP basic auth username, if your Web API is behind basic auth     |
| `basicPass` | HTTP basic auth password, if your Web API is behind basic auth     |

## Reannounce - `[reannounce]`

Some trackers are buggy and need a reannounce before a torrent can start.
These values are used by [`qbt torrent reannounce`](/qbittorrent-cli/commands/qbt_torrent_reannounce/).

```toml
[reannounce]
enabled  = true # true or false
attempts = 10   # number of attempts, typically 10-30
interval = 7000 # interval between attempts, in milliseconds
```

## Rules - `[rules]`

Limit how many torrents download simultaneously. Applied when adding torrents
unless you pass `--ignore-rules`.

```toml
[rules]
enabled              = true # enable or disable rules
max_active_downloads = 2    # max active downloads
```

* On HDDs with 1 Gbit, `max_active_downloads = 2` is a good value to avoid
  overloading the disks while giving each torrent as much bandwidth as possible.
* On SSDs and 1 Gbit+ you can increase this value.

## Add defaults - `[add]`

Defaults applied by [`qbt torrent add`](/qbittorrent-cli/commands/qbt_torrent_add/).

```toml
[add]
sequential       = false # download pieces in sequential order (useful for streaming)
first_last_piece = false # prioritize first and last pieces for all new torrents
```

## Compare instances - `[[compare]]`

[`qbt torrent compare`](/qbittorrent-cli/commands/qbt_torrent_compare/) can
compare torrents against other qBittorrent instances. Add one `[[compare]]`
block per instance.

```toml
[[compare]]
addr     = "http://100.100.100.100:6776"
login    = "user"
password = "password"
#basicUser = "user"
#basicPass = "password"

[[compare]] # you can specify multiple compare blocks
addr     = "http://100.100.100.101:6776"
login    = "user"
password = "password"
```

## autodl-irssi

`qbt torrent add` works well as an autodl-irssi upload action.

Edit `autodl.cfg`.

Global action:

```ini
[options]
...
upload-type = exec
upload-command = /usr/bin/qbt
upload-args = add "$(TorrentPathName)"
```

Per filter:

```ini
[filter example_filter_action]
...
upload-type = exec
upload-command = /usr/bin/qbt
upload-args = add "$(TorrentPathName)" --category cat1 --tags tag1
```

### ruTorrent

In ruTorrent, go to autodl-irssi `Preferences`, then the `Action` tab and set
the following for the global action. This can be set per filter as well, where
you can also add a category, tags and so on.

```text
Choose .torrent action: Run Program
Command: /usr/bin/qbt
Arguments: add "$(TorrentPathName)"
```
