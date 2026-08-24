---
title: Installation
---

# Installation

## Release binaries

Every tag publishes stripped `linux` and `darwin` archives for `amd64` and `arm64`, plus an
unstripped `linux` debug build:

```shell
VERSION=$(curl -fsSL https://api.github.com/repos/projecteru2/cli/releases/latest | jq -r .tag_name)
curl -fsSL -o eru-cli.tar.gz \
  "https://github.com/projecteru2/cli/releases/download/${VERSION}/eru-cli_${VERSION}_Linux_x86_64.tar.gz"
tar xzf eru-cli.tar.gz eru-cli
install -m 0755 eru-cli /usr/local/bin/eru-cli
eru-cli --version
```

The archive name follows `eru-cli_<version>_<Os>_<Arch>.tar.gz`, where `Os` is `Linux` or `Darwin`
and `Arch` is `x86_64` or `arm64`. An unversioned alias, `eru-cli_<Os>_<Arch>.tar.gz`, always points
at the latest release.

## Container image

Multi-arch images (`linux/amd64`, `linux/arm64`) are published to Docker Hub and ghcr on every tag
and every push to `master`:

```shell
docker run -it --rm --net host ghcr.io/projecteru2/cli eru-cli --eru <core-host>:5001 pod list
```

`--net host` is the simplest way to reach a core instance that listens on the host network.

## From source

Go 1.27 or newer is required; the toolchain in `go.mod` is downloaded automatically.

```shell
git clone https://github.com/projecteru2/cli
cd cli
make build
./eru-cli --version
```

`make build` writes `./eru-cli` with the version, revision and build time linked in. Set
`KEEP_SYMBOL=1` to keep the symbol table for debugging.

## Pointing at a core

The client needs one address and, if core has authentication enabled, a username and password:

```shell
export ERU=127.0.0.1:5001
export ERU_USERNAME=eru
export ERU_PASSWORD=secret
eru-cli core info
```

The address may also be given per invocation with `--eru`. When several core instances register
themselves in the same cluster, the address is only the entry point: core's own service discovery
decides which instance ultimately serves the call.
