# qbittorrent-cli

A cli to manage qBittorrent.

## Install

Download binary and put somewhere in $PATH.

Extract binary

    tar -xzvf qbt.tar.gz

Move to somewhere in `$PATH`. Need sudo if not already root. Or put it in your user `$HOME/bin` or similar.

    sudo mv qbt /usr/bin/

Verify that it runs

    qbt

This should print out basic usage.

## Configuration

Create a new configuration file `.qbt.toml` in `$HOME/.config/qbt/`.

    mkdir -p ~/.config/qbt && touch ~/.config/qbt/.qbt.toml

A bare minimum config.

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

### rutorrent-autodl-irssi setup

In rutrorrent, go to autodl-irssi `Preferences`, and then the `Action` tab. Put in the following for the global action. This can be set in a specific filter as well.

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

Global flags:
  * `--config` -override config file
  
Use "qbt [command] --help" for more information about a command.

### Add

Flags:
  * `--dry-run` - Run without doing anything
  * `--magnet` - Add magnet link instead of torrent file
  * `--paused` - Add torrent in paused state
  * `--skip-hash-check` - Skip hash check
  * `--save-path` - Add torrent to the specified path
  * `--category` - Add torrent to the specified category
  * `--tags` - Add tags to the torrent. Use multiple or comma-separate tags e.g. --tags linux,iso. Supported in 4.3.2+ 
  * `--ingore-rules` - Ignore rules set in config
  * `--limit-ul` - Set torrent upload speed limit. Unit in bytes/second
  * `--limit-dl` - Set torrent download speed limit. Unit in bytes/second

Add a new torrent to qBittorrent.

    qbt add my-torrent-file.torrent
