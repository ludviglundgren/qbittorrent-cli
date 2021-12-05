# qbittorrent-cli

[![goreleaser](https://github.com/ludviglundgren/qbittorrent-cli/actions/workflows/release.yml/badge.svg)](https://github.com/ludviglundgren/qbittorrent-cli/actions/workflows/release.yml)

A cli to manage qBittorrent. Add torrents, reannounce and import from other clients.

## Features

* Add torrents to qBittorrent from file or magnet link. Useful in combination with autodl-irssi
* Reannounce torrents for troublesome trackers
* Set limits on how many simultaneously active downloads are allowed
* Import torrents with state from Deluge and rTorrent

## Install

Download the [latest binary](https://github.com/ludviglundgren/qbittorrent-cli/releases/latest) and put somewhere in $PATH.

Extract binary

    tar -xzvf qbt_$VERSION_linux_amd64.tar.gz

Move to somewhere in `$PATH`. Need sudo if not already root. Or put it in your user `$HOME/bin` or similar.

    sudo mv qbt /usr/bin/

Verify that it runs

    qbt

This should print out basic usage.

Thanks to Go we can build binaries for multiple platforms, but currently only 32 and 64bit linux is built. Add an issue if you need it built for something else.

## Build from source

You can also build it yourself if you have Go installed locally, or with `goreleaser`.

With `make`

    make build

Or with only go

    go build -o bin/qbt cmd/qbt/main.go

### Multi-platform with `goreleaser`

Builds with `goreleaser` will also include version info.

    goreleaser --snapshot --skip-publish --rm-dist

## Configuration

Create a new configuration file `.qbt.toml` in `$HOME/.config/qbt/`.

    mkdir -p ~/.config/qbt && touch ~/.config/qbt/.qbt.toml

A bare minimum config. Check [full example config](.qbt.toml.example).

```toml
[qbittorrent]
host     = "127.0.0.1" # qbittorrent webui-api hostname/ip
port     = 6776        # qbittorrent webui-api port
login    = "user"      # qbittorrent webui-api user
password = "password"  # qbittorrent webui-api password

[rules]
enabled              = true   # enable or disable rules
max_active_downloads = 2      # set max active downloads
```

* If running on HDDs and 1Gbit - `max_active_downloads = 2` is a good setting to not overload the disks and gives as much bandwidth as possible to the torrents.
* For SSDs and 1Gbit+ you can increase this value.

### autodl-irssi setup

Edit `autodl.cfg`

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

### rutorrent

In rutrorrent, go to autodl-irssi `Preferences`, and then the `Action` tab. Put in the following for the global action. This can be set per filter as well, then you can add category, tags etc.

```
Choose .torrent action: Run Program
Command: /usr/bin/qbt
Arguments: add "$(TorrentPathName)"
```

## Usage

Use `qbt help` to find out more about how to use.

Commands:
  - add
  - list
  - version
  - help
  - move

Global flags:
  * `--config` - override config file
  
Use "qbt [command] --help" for more information about a command.

### Add

Add a new torrent to qBittorrent.

    qbt add my-torrent-file.torrent

Optional flags:
* `--dry-run` - Run without doing anything
* `--magnet <LINK>` - Add magnet link instead of torrent file
* `--paused` - Add torrent in paused state
* `--skip-hash-check` - Skip hash check
* `--save-path <PATH>` - Add torrent to the specified path
* `--category <CATEGORY>` - Add torrent to the specified category
* `--tags <TAG,TAG>` - Add tags to the torrent. Use multiple or comma-separate tags e.g. --tags linux,iso. Supported in 4.3.2+
* `--ingore-rules` - Ignore rules set in config
* `--limit-ul <SPEED>` - Set torrent upload speed limit. Unit in bytes/second
* `--limit-dl <SPEED>` - Set torrent download speed limit. Unit in bytes/second
 
### Move

Move torrents from one or multiple categories to some other category.

By using ATM (Automatic Torrent Mode) and the default behavior of categories mapped to folders this can be used to move from NVME to HDD storage after a `min-seed-time`.

    qbt move --from categroy1 --to category2 --min-seed-time 45

Optional flags:
* `--dry-run` - Run without doing anything
* `--min-seed-time` - Only move torrents with a minimum active seedtime of X minutes

Usable with cron as well. Run every 15 min.

    */15 * * * * /usr/bin/qbt move --from nvme --to hdd --min-seed-time 30

### Import

Import torrents from other client into qBittorrent, and keep state. 

Required flags:
* `--source <NAME>` - Import from client [deluge, rtorrent]
* `--source-dir <PATH>` - State/session dir for client
* `--qbit-dir <PATH>` - qBittorrent dir (~/.local/share/data/qBittorrent/BT_backup)

Supported clients:
* Deluge
* rTorrent

Optional flags:
* `--dry-run` - don't write anything to disk
* `--skip-backup` - skip backup of state dirs before importing data

> WARNING: Make sure to stop both the source client and qBittorrent before importing.

Example with Deluge.

    qbt import --source deluge --source-dir ~/.config/deluge/state/ --qbit-dir ~/.local/share/data/qBittorrent/BT_backup --dry-run

After the import you will have to manually delete the torrents from the source client, but don't check the "also delete files" as currently the import DOES NOT move the actual data.

#### Caveats

- Does not support changing paths for data, it expects data to be at same place as the source client.
- Does not support renamed files either.
- Does not import labels/categories/tags
- Use at own caution. The backups are there if something goes wrong.

#### Workflow

Torrents imported into qBittorrent does not have automatic management enabled, because it's default behavior is to move data.

1. Stop source client and qBittorrent.
2. Start with a dry run and see what it does `qbt import ..... --dry-run`
3. If it looks ok, run without `--dry-run`
4. Start clients again, go into the source client and stop everything.
5. Set categories/tags in batches. Start to add a category, then set "Automatic torrent management" for it to automatically move the files to the categories specified directory.

