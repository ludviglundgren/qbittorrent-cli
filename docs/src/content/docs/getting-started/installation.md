---
title: Installation
description: Install qbittorrent-cli from a release binary, with go install, or build it from source.
---

`qbt` ships as a single static binary for Linux, macOS and Windows.

## Download a release binary

Download the [latest release](https://github.com/ludviglundgren/qbittorrent-cli/releases/latest)
for your platform and put it somewhere in your `$PATH`.

Extract the archive:

```shell
tar -xzvf qbittorrent-cli_${VERSION}_linux_amd64.tar.gz
```

Move it somewhere in `$PATH` (needs `sudo` if you are not root), or into your
user `$HOME/bin`:

```shell
sudo mv qbt /usr/bin/
```

Verify that it runs - this should print the basic usage:

```shell
qbt
```

## Install with Go

If you have Go installed you can install the latest release directly:

```shell
go install github.com/ludviglundgren/qbittorrent-cli/cmd/qbt@latest
```

## Build from source

Clone the repository and build with `make`:

```shell
make build
```

Or with the Go toolchain directly:

```shell
go build -o bin/qbt cmd/qbt/main.go
```

### Multi-platform builds with goreleaser

Builds made with [goreleaser](https://goreleaser.com/) also embed version info:

```shell
goreleaser release --snapshot --skip=publish --clean
```

## Next steps

Once installed, head to the [Configuration](/qbittorrent-cli/getting-started/configuration/)
guide to connect `qbt` to your qBittorrent instance.
