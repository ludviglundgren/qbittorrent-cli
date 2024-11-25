# qbittorrent-cli

[![goreleaser](https://github.com/ludviglundgren/qbittorrent-cli/actions/workflows/release.yml/badge.svg)](https://github.com/ludviglundgren/qbittorrent-cli/actions/workflows/release.yml)

A cli to manage qBittorrent. Add torrents, categories, tags, reannounce and import from other clients.

## Features

* qBittorrent v5 compatible as of version v2.1.0
* Add torrents to qBittorrent from file or magnet link. Useful in combination with autodl-irssi
* Reannounce torrents for troublesome trackers
* Set limits on how many simultaneously active downloads are allowed
* Import torrents with state from Deluge and rTorrent
* Manage categories
* Manage tags
* Self updater

## Table of contents

* [Features](#features)
* [Install](#install)
* [Build from source](#build-from-source)
  * [Multi-platform with `goreleaser`](#multi-platform-with-goreleaser)
* [Configuration](#configuration)
* [Usage](#usage)
  * [App](#app)
    * [Version](#version)
  * [Bencode](#bencode)
    * [Edit](#edit)
  * [Category](#category)
    * [Add](#add)
    * [Delete](#delete)
    * [Edit](#edit-1)
    * [List](#list)
  * [Tag](#tag)
    * [Add](#add-1)
    * [Delete](#delete-1)
    * [List](#list-1)
  * [Torrent](#torrent)
    * [Add](#add-2)
    * [Category](#category-1)
      * [Move](#move)
      * [Set](#set)
      * [Unset](#unset)
    * [Compare](#compare)
    * [Export](#export)
    * [Hash](#hash)
    * [Import](#import)
      * [Caveats](#caveats)
      * [Workflow](#workflow)
    * [List](#list-2)
    * [Pause](#pause)
    * [Reannounce](#reannounce)
    * [Recheck](#recheck)
    * [Remove](#remove)
    * [Resume](#resume)
    * [Tag](#tag-1)
      * [Issues](#issues)
    * [Tracker](#tracker)
      * [Tracker edit](#tracker-edit)
  * [Version](#version-1)
  * [Update](#update)


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

    goreleaser release --snapshot --skip=publish --clean

## Configuration

Create a new configuration file `.qbt.toml` in `$HOME/.config/qbt/`.

    mkdir -p ~/.config/qbt && touch ~/.config/qbt/.qbt.toml

A bare minimum config. Check [full example config](.qbt.toml.example).

```toml
[qbittorrent]
addr       = "http://127.0.0.1:6776" # qbittorrent webui-api hostname/ip
login      = "user"                  # qbittorrent webui-api user
password   = "password"              # qbittorrent webui-api password
#basicUser = "user"                  # qbittorrent webui-api basic auth user
#basicPass = "password"              # qbittorrent webui-api basic auth password

[rules]
enabled              = true   # enable or disable rules
max_active_downloads = 2      # set max active downloads
```

* If running on HDDs and 1Gbit - `max_active_downloads = 2` is a good setting to not overload the disks and gives as much bandwidth as possible to the torrents.
* For SSDs and 1Gbit+ you can increase this value.

<details>
<summary>autodl-irssi setup</summary>

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

</details>

## Usage

Use `qbt help` to find out more about how to use.

```text
Usage:
  qbt [command]

Available Commands:
  app         app subcommand
  bencode     bencode subcommand
  category    category subcommand
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  tag         tag subcommand
  torrent     torrent subcommand
  update      Update qbittorrent-cli to latest version
  version     Print the version

Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
  -h, --help            help for qbt

Use "qbt [command] --help" for more information about a command.
```

Commands:
* [app](#app)
* [bencode](#bencode)
* [category](#category)
* [tag](#tag)
* [torrent](#torrent)
* [version](#version-1)
* [update](#update)

### App

```
Get qBittorrent info

Usage:
  qbt app [command]

Available Commands:
  version     Get application version

Flags:
  -h, --help   help for app
```

#### Version

```
Get qBittorrent version info

Usage:
  qbt app version [flags]

Flags:
  -h, --help   help for version
```

### Bencode

```
Do various bencode operations

Usage:
  qbt bencode [command]

Available Commands:
  edit        edit bencode data

Flags:
  -h, --help   help for bencode

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

#### Edit

```
Edit bencode files like .fastresume. Shut down client and make a backup of data before.

Usage:
  qbt bencode edit [flags]

Examples:
  qbt bencode edit --dir /home/user/.local/share/qBittorrent/BT_backup --pattern '/home/user01/torrents' --replace '/home/test/torrents'

Flags:
      --dir string       Dir with fast-resume files (required)
      --dry-run          Dry run, don't write changes
  -h, --help             help for edit
      --pattern string   Pattern to change (required)
      --replace string   Text to replace pattern with (required)
  -v, --verbose          Verbose output

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

Edit fastresume data like save-path. Make sure to shut down the client and backup the data before running this.

## Category

```
Do various category actions

Usage:
  qbt category [command]

Available Commands:
  add         Add category
  delete      Delete category
  edit        Edit category
  list        List categories

Flags:
  -h, --help   help for category

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)

Use "qbt category [command] --help" for more information about a command.
```

#### Add

```
Add new category

Usage:
  qbt category add [flags]

Examples:
  qbt category add test-category
  qbt category add test-category --save-path "/home/user/torrents/test-category"

Flags:
      --dry-run            Run without doing anything
  -h, --help               help for add
      --save-path string   Category default save-path. Optional. Defaults to dir in default save dir.

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

#### Delete

```
Delete category

Usage:
  qbt category delete [flags]

Examples:
  qbt category delete test-category

Flags:
      --dry-run   Run without doing anything
  -h, --help      help for delete

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

#### Edit

```
Edit category

Usage:
  qbt category edit [flags]

Examples:
  qbt category edit test-category --save-path "/home/user/new/path"
  qbt category edit test-category --save-path ""

Flags:
      --dry-run            Run without doing anything
  -h, --help               help for edit
      --save-path string   Edit category save-path

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

#### List

```
List categories

Usage:
  qbt category list [flags]

Flags:
  -h, --help            help for list
      --output string   Print as [formatted text (default), json]

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

## Tag

```
Do various tag actions

Usage:
  qbt tag [command]

Available Commands:
  add         Add tags
  delete      Delete tags
  list        List tags

Flags:
  -h, --help   help for tag

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

#### Add

```
Add new tag

Usage:
  qbt tag add [flags]

Examples:
  qbt tag add tag1

Flags:
      --dry-run   Run without doing anything
  -h, --help      help for add

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

#### Delete

```
Usage:
  qbt tag delete [flags]

Examples:
  qbt tag delete tag1

Flags:
      --dry-run   Run without doing anything
  -h, --help      help for delete

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

#### List

```
List tags.

Usage:
  qbt tag list [flags]

Flags:
  -h, --help            help for list
      --output string   Print as [formatted text (default), json]

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

## Torrent

```text
Do various torrent operations

Usage:
  qbt torrent [command]

Available Commands:
  add         Add torrent(s)
  category    Torrent category subcommand
  compare     Compare torrents
  export      Export torrents
  hash        Print the hash of a torrent file or magnet
  import      Import torrents
  list        List torrents
  pause       Pause specified torrent(s)
  reannounce  Reannounce torrent(s)
  remove      Removes specified torrent(s)
  resume      Resume specified torrent(s)
  tracker     Torrent tracker subcommand

Flags:
  -h, --help   help for torrent

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)

Use "qbt torrent [command] --help" for more information about a command.
```

### Add

```text
Add new torrent(s) to qBittorrent from file or magnet. Supports glob pattern for files like: ./files/*.torrent

Usage:
  qbt torrent add [flags]

Examples:
  qbt torrent add my-file.torrent --category test --tags tag1
  qbt torrent add ./files/*.torrent --paused --skip-hash-check
  qbt torrent add magnet:?xt=urn:btih:5dee65101db281ac9c46344cd6b175cdcad53426&dn=download

Flags:
      --category string    Add torrent to the specified category
      --dry-run            Run without doing anything
  -h, --help               help for add
      --ignore-rules       Ignore rules from config
      --limit-dl uint      Set torrent download speed limit. Unit in bytes/second
      --limit-ul uint      Set torrent upload speed limit. Unit in bytes/second
      --paused             Add torrent in paused state
      --remove-stalled     Remove stalled torrents from re-announce
      --save-path string   Add torrent to the specified path
      --skip-hash-check    Skip hash check
      --tags stringArray   Add tags to torrent

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Category

```text
Do various torrent category operations

Usage:
  qbt torrent category [command]

Available Commands:
  move        move torrents between categories
  set         Set torrent category
  unset       Unset torrent category

Flags:
  -h, --help   help for category

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)

Use "qbt torrent category [command] --help" for more information about a command.
```

#### Move

```text
Move torrents from one category to another

Usage:
  qbt torrent category move [flags]

Examples:
  qbt torrent category move --from cat1 --to cat2

Flags:
      --dry-run                Run without doing anything
      --exclude-tags strings   Exclude torrents with provided tags
      --from strings           Move from categories (required)
  -h, --help                   help for move
      --include-tags strings   Include torrents with provided tags
      --min-seed-time int      Minimum seed time in MINUTES before moving.
      --to string              Move to the specified category (required)

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

Move torrents from one or multiple categories to some other category.

By using ATM (Automatic Torrent Mode) and the default behavior of categories mapped to folders this can be used to move from NVME to HDD storage after a `min-seed-time`.

    qbt torrent category move --from categroy1 --to category2 --min-seed-time 45

Optional flags:
* `--dry-run` - Run without doing anything
* `--min-seed-time` - Only move torrents with a minimum active seedtime of X minutes
* `--include-tags` - Only move torrents with any of the tags in comma separate list (tag1,tag2)
* `--exclude-tags` - Don't move torrents with any of the tags in comma separate list (tag1,tag2)

Usable with cron as well. Run every 15 min.

```cronexp
*/15 * * * * /usr/bin/qbt move --from nvme --to hdd --min-seed-time 30
```

#### Set

```text
Set category for torrents via hashes

Usage:
  qbt torrent category set [flags]

Examples:
  qbt torrent category set test-category --hashes hash1,hash2

Flags:
      --dry-run          Run without doing anything
      --hashes strings   Torrent hashes, as comma separated list
  -h, --help             help for set

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

#### Unset

```text
Unset category for torrents via hashes

Usage:
  qbt torrent category unset [flags]

Examples:
  qbt torrent category unset --hashes hash1,hash2

Flags:
      --dry-run          Run without doing anything
      --hashes strings   Torrent hashes, as comma separated list
  -h, --help             help for unset

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Compare

```text
Compare torrents between clients

Usage:
  qbt torrent compare [flags]

Examples:
  qbt torrent compare --addr http://localhost:10000 --user u --pass p --compare-addr http://url.com:10000 --compare-user u --compare-pass p

Flags:
      --basic-pass string           Source basic auth pass
      --basic-user string           Source basic auth user
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

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

Compare torrents between two instances. Source instance and `compare` instance.

Required flags:
* `--addr` - Host url with http and port if needed
* `--user` - Host user
* `--pass` - Host pass
* `--basic-user` - Host basic auth user
* `--basic-pass` - Host basic auth pass

* `--compare-addr` - url with http and port if needed
* `--compare-user` - user
* `--compare-pass` - pass
* `--compare-basic-user` - basic auth user
* `--compare-basic-pass` - basic auth pass

Optional flags:
* `--dry-run` - Run without doing anything
* `--tag-duplicates` - Tag duplicates with `compare-dupe` tag, only on compare host
* `--tag` - Override the default tag `compare-dupe`

### Export

```text
Export torrents and fastresume by category

Usage:
  qbt torrent export [flags]

Examples:
  qbt torrent export --source ~/.local/share/data/qBittorrent/BT_backup --export-dir ~/qbt-backup --include-category=movies,tv

Flags:
      --dry-run                    dry run
      --exclude-category strings   Exclude categories. Comma separated
      --exclude-tag strings        Exclude tags. Comma separated
      --export-dir string          Dir to export files to (required)
  -h, --help                       help for export
      --include-category strings   Export torrents from these categories. Comma separated
      --include-tag strings        Include tags. Comma separated
      --skip-manifest              Do not export all used tags and categories into manifest
      --source string              Dir with torrent and fast-resume files (required)
  -v, --verbose                    verbose output

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Hash

```text
Print the hash of a torrent file or magnet

Usage:
  qbt torrent hash [flags]

Examples:
  qbt torrent hash file.torrent

Flags:
  -h, --help   help for hash

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Import

```text
Import torrents with state from other clients [rtorrent, deluge]

Usage:
  qbt torrent import {rtorrent | deluge} --source-dir dir --qbit-dir dir2 [--skip-backup] [--dry-run] [flags]

Examples:
  qbt torrent import deluge --source-dir ~/.config/deluge/state/ --qbit-dir ~/.local/share/data/qBittorrent/BT_backup --dry-run
  qbt torrent import rtorrent --source-dir ~/.sessions --qbit-dir ~/.local/share/data/qBittorrent/BT_backup --dry-run

Flags:
      --dry-run             Run without importing anything
  -h, --help                help for import
      --qbit-dir string     qBittorrent BT_backup dir. Commonly ~/.local/share/qBittorrent/BT_backup (required)
      --skip-backup         Skip backup before import
      --source-dir string   source client state dir (required)

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

Import torrents from other client into qBittorrent, and keep state. 

> WARNING: Make sure to stop both the source client and qBittorrent before importing.

After the import you will have to manually delete the torrents from the source client, but don't check the "also delete files" as currently the import DOES NOT move the actual data.

#### Caveats

- Does not support changing paths for data, it expects data to be at same place as the source client.
- Does not support renamed files either.
- Does not import labels/categories/tags
- Use at own caution. The backups are there if something goes wrong.

#### Workflow

Torrents imported into qBittorrent does not have automatic management enabled, because it's default behavior is to move data.

1. Stop source client and qBittorrent.
2. Start with a dry run and see what it does `qbt torrent import ..... --dry-run`
3. If it looks ok, run without `--dry-run`
4. Start clients again, go into the source client and stop everything.
5. Set categories/tags in batches. Start to add a category, then set "Automatic torrent management" for it to automatically move the files to the categories specified directory.

### List

```text
List all torrents, or torrents with a specific filters. Get by filter, category, tag and hashes. Can be combined

Usage:
  qbt torrent list [flags]

Examples:
qbt torrent list --filter=downloading --category=linux-iso

Flags:
  -c, --category string   Filter by category. All categories by default.
  -f, --filter string     Filter by state. Available filters: all, downloading, seeding, completed, paused, active, inactive, resumed, 
                          stalled, stalled_uploading, stalled_downloading, errored (default "all")
      --hashes strings    Filter by hashes. Separated by comma: "hash1,hash2".
  -h, --help              help for list
      --output string     Print as [formatted text (default), json]
  -t, --tag string        Filter by tag. Single tag: tag1

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Pause

```text
Pauses torrents indicated by hash, name or a prefix of either. Whitespace indicates next prefix unless argument is surrounded by quotes

Usage:
  qbt torrent pause [flags]

Flags:
      --all              Pauses all torrents
      --hashes strings   Add hashes as comma separated list
  -h, --help             help for pause
      --names            Provided arguments will be read as torrent names

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Reannounce

```text
Reannounce torrents with non-OK tracker status.

Usage:
  qbt torrent reannounce [flags]

Flags:
      --attempts int      Reannounce torrents X times (default 50)
      --category string   Reannounce torrents with category
      --dry-run           Run without doing anything
      --hash string       Reannounce torrent with hash
  -h, --help              help for reannounce
      --interval int      Reannounce torrents X times with interval Y. In MS (default 7000)
      --tag string        Reannounce torrents with tag

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Recheck

```text
Rechecks torrents indicated by hash(es).

Usage:
  qbt torrent recheck [flags]

Examples:
  qbt torrent recheck --hashes HASH
  qbt torrent recheck --hashes HASH1,HASH2


Flags:
      --hashes strings   Add hashes as comma separated list
  -h, --help             help for recheck

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Remove

```text
Removes torrents indicated by hash, name or a prefix of either. Whitespace indicates next prefix unless argument is surrounded by quotes

Usage:
  qbt torrent remove [flags]

Flags:
      --all              Removes all torrents
      --delete-files     Also delete downloaded files from torrent(s)
      --dry-run          Display what would be done without actually doing it
      --hashes strings   Add hashes as comma separated list
  -h, --help             help for remove
      --paused           Removes all paused torrents

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Resume

```text
Resumes torrents indicated by hash, name or a prefix of either. Whitespace indicates next prefix unless argument is surrounded by quotes

Usage:
  qbt torrent resume [flags]

Flags:
      --all              resumes all torrents
      --hashes strings   Add hashes as comma separated list
  -h, --help             help for resume

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Tag

Tag torrents.

```text
Do various torrent tag operations

Usage:
  qbt torrent tag [command]

Available Commands:
  issues      tag torrents with issues

Flags:
  -h, --help   help for tag

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

#### Issues

```text
Tag torrents that may have broken trackers or be unregistered

Usage:
  qbt torrent tag issues [flags]

Examples:
  qbt torrent tag issues --unregistered --not-working

Flags:
      --dry-run        Dry run, do not tag torrents
  -h, --help           help for issues
      --not-working    tag not working torrents
      --unregistered   tag unregistered

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

### Tracker

```text
Do various torrent category operations

Usage:
  qbt torrent tracker [command]

Available Commands:
  edit        Edit torrent tracker

Flags:
  -h, --help   help for tracker

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)

Use "qbt torrent tracker [command] --help" for more information about a command.
```

#### Tracker edit

```text
Edit tracker for torrents via hashes

Usage:
  qbt torrent tracker edit [flags]

Examples:
  qbt torrent tracker edit --old url.old/test --new url.com/test

Flags:
      --dry-run          Run without doing anything
  -h, --help             help for edit
      --new string       New tracker URL
      --old string       Old tracker URL to replace

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

## Version

```text
Print qbt version info

Usage:
  qbt version [flags]

Examples:
  qbt version
  qbt version --output json

Flags:
  -h, --help            help for version
      --output string   Print as [text, json] (default "text")

Global Flags:
      --config string   config file (default is $HOME/.config/qbt/.qbt.toml)
```

## Update

```text
Update qbittorrent-cli to latest version

Usage:
  qbt update [flags]

Examples:
  qbt update

Flags:
  -h, --help      help for update
      --verbose   Verbose output: Print changelog
```
