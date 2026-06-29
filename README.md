# qbittorrent-cli

[![goreleaser](https://github.com/ludviglundgren/qbittorrent-cli/actions/workflows/release.yml/badge.svg)](https://github.com/ludviglundgren/qbittorrent-cli/actions/workflows/release.yml)

A CLI to manage qBittorrent. Add torrents, manage categories and tags, reannounce, and import from other clients.

> [!NOTE]
> **📖 Full documentation lives at [ludviglundgren.github.io/qbittorrent-cli](https://ludviglundgren.github.io/qbittorrent-cli/).**
> The complete, always up-to-date command and flag reference is generated directly from the source - see [Command reference](#command-reference) below.

## Features

* qBittorrent v5 compatible (since v2.1.0)
* Add torrents from file, glob pattern, URL or magnet - handy with autodl-irssi
* Reannounce torrents for troublesome trackers
* Limit how many torrents download simultaneously
* Import torrents with state from Deluge and rTorrent
* Manage categories and tags
* Compare torrents between instances
* Self updater

## Install

`qbt` is a single static binary for Linux, macOS and Windows.

Download the [latest release](https://github.com/ludviglundgren/qbittorrent-cli/releases/latest), extract it, and move it somewhere in your `$PATH`:

```shell
tar -xzvf qbittorrent-cli_${VERSION}_linux_amd64.tar.gz
sudo mv qbt /usr/bin/
qbt   # prints usage
```

Or install with Go:

```shell
go install github.com/ludviglundgren/qbittorrent-cli/cmd/qbt@latest
```

See the [installation guide](https://ludviglundgren.github.io/qbittorrent-cli/getting-started/installation/) for building from source and other options.

## Configuration

Create `.qbt.toml` in `~/.config/qbt/`:

```shell
mkdir -p ~/.config/qbt && touch ~/.config/qbt/.qbt.toml
```

A bare-minimum config - see the [full example](.qbt.toml.example):

```toml
[qbittorrent]
addr     = "http://127.0.0.1:6776" # qbittorrent webui-api hostname/ip
apikey   = "APIKEY"                # qbittorrent webui-api api key, qbit 5.2.x+ (optional)
login    = "user"                  # qbittorrent webui-api user                 (optional)
password = "password"              # qbittorrent webui-api password             (optional)

[rules]
enabled              = true # enable or disable rules
max_active_downloads = 2    # set max active downloads
```

The [configuration guide](https://ludviglundgren.github.io/qbittorrent-cli/getting-started/configuration/) documents every option, including `[reannounce]`, `[add]`, `[[compare]]` instances and autodl-irssi / ruTorrent setup.

## Usage

Run `qbt help` or `qbt [command] --help` for details on any command.

```shell
# Add a torrent with a category and tags
qbt torrent add my-file.torrent --category linux-iso --tags iso

# Add everything matching a glob, paused
qbt torrent add ./files/*.torrent --paused --skip-hash-check

# List downloading torrents in a category
qbt torrent list --filter downloading --category linux-iso

# Reannounce a torrent that is stuck on the tracker
qbt torrent reannounce --hash <hash>

# Move torrents from NVMe to HDD after 45 minutes of seeding (handy in cron)
qbt torrent category move --from nvme --to hdd --min-seed-time 45

# Set a share ratio limit on all torrents
qbt torrent share-limit set --all --ratio 2.0

# Import torrents (and their state) from Deluge - always dry-run first
qbt torrent import deluge \
  --source-dir ~/.config/deluge/state \
  --qbit-dir ~/.local/share/qBittorrent/BT_backup \
  --dry-run
```

## Command reference

The full reference for every command and flag is generated from the source and published at
[ludviglundgren.github.io/qbittorrent-cli](https://ludviglundgren.github.io/qbittorrent-cli/commands/qbt/).

| Command                                                                                  | Description                                              |
| ---------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| [`torrent`](https://ludviglundgren.github.io/qbittorrent-cli/commands/qbt_torrent/)      | Add, list, export, import and manage torrents           |
| [`category`](https://ludviglundgren.github.io/qbittorrent-cli/commands/qbt_category/)    | Add, edit, delete and list categories                   |
| [`tag`](https://ludviglundgren.github.io/qbittorrent-cli/commands/qbt_tag/)              | Add, delete and list tags                               |
| [`transfer`](https://ludviglundgren.github.io/qbittorrent-cli/commands/qbt_transfer/)    | Show transfer / session status and speeds               |
| [`app`](https://ludviglundgren.github.io/qbittorrent-cli/commands/qbt_app/)              | Show qBittorrent application and Web API versions        |
| [`bencode`](https://ludviglundgren.github.io/qbittorrent-cli/commands/qbt_bencode/)      | Edit bencode files such as `.fastresume`                |
| [`version`](https://ludviglundgren.github.io/qbittorrent-cli/commands/qbt_version/)      | Print `qbt` version info                                 |
| [`update`](https://ludviglundgren.github.io/qbittorrent-cli/commands/qbt_update/)        | Update `qbt` to the latest version                      |

## Contributing

The command reference under `docs/src/content/docs/commands/` is generated - do not edit it by hand. After changing a command or flag, regenerate it:

```shell
make docs
```

CI fails if the committed docs are out of date. The documentation site is built with [Starlight](https://starlight.astro.build/) and lives in [`docs/`](docs/).

## License

[MIT](LICENSE)
