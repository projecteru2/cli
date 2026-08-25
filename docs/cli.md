---
title: Command reference
---

# Command reference

```
eru-cli [global options] <command> [command options] [arguments...]
```

`command` is one of eight groups: `core`, `image`, `lambda`, `network`, `node`, `pod`, `status` and
`workload`. `--help` works at every level and prints the same information this page derives from.

## Global options

Global options come before the command name and apply to all of it.

| Option | Environment | Default | Meaning |
|---|---|---|---|
| `--eru`, `-e` | `ERU` | `127.0.0.1:5001` | Address of the eru core to call. |
| `--username`, `-u` | `ERU_USERNAME` | empty | Username when core requires authentication. |
| `--password`, `-p` | `ERU_PASSWORD` | empty | Password when core requires authentication. |
| `--output`, `-o` | `ERU_OUTPUT_FORMAT` | empty | `json`, `yaml`, or empty for a table. |
| `--debug`, `-d` | | off | Log at debug level instead of info. |
| `--version`, `-v` | | | Print version, revision, build time and Go toolchain. |

The table format prints a readable summary; `json` and `yaml` print the full message as core
returned it and are the right choice for scripting.

## Exit status

`0` means every item succeeded. Commands that act on several workloads, images or nodes in one call
— `deploy`, `replace`, `send`, `sendlarge`, `copy`, `start`, `stop`, `restart`, `remove`,
`dissociate`, `image cache`, `image remove`, `network connect`, `network disconnect` — report each
failure and exit non-zero if any of them failed, so a script does not need to parse the log. Any
failing command exits `255`, except `workload exec` and `lambda`, which exit with the remote
command's own code.

## core

Inspect the core instance itself.

| Command | Arguments | Purpose |
|---|---|---|
| `core info` | | Version, revision, build time and identifier of the core being called. |
| `core watch` | | Follow the service discovery stream and print core addresses as they change. |

## pod

A pod is a named group of nodes.

| Command | Arguments | Notable options |
|---|---|---|
| `pod list` | | |
| `pod add` | `<pod name>` | `--desc` |
| `pod remove` | `<pod name>` | |
| `pod nodes` | `<pod name>` | `--filter up\|down\|all` (default `all`), `--label a=1`, `--timeout 10`, `--show-info`, `--stream` |
| `pod networks` | `<pod name>` | `--driver` |
| `pod resource` | `<pod name>` | `--filter`, `--stream` |
| `pod capacity` | `<pod name>` | `--cpu`, `--memory`, `--storage`, `--cpu-bind`, `--node`, `--extra-resources` |

`pod resource --filter` takes an expression over the usage percentages, for example
`--filter "cpu > 40%"` or `--filter "memory <= 0.4"`. The attribute is one of `cpu`, `memory`,
`storage` or `volume`, the operator one of `>`, `>=`, `<`, `<=`, `==`.

`pod capacity` asks core how many workloads of a given size would still fit, without deploying
anything:

```shell
eru-cli pod capacity --cpu 2 --memory 1G --storage 10G <pod>
```

## node

| Command | Arguments | Notable options |
|---|---|---|
| `node add` | `<pod name>` | `--nodename` (`$HOSTNAME`), `--endpoint`, `--ca`/`--cert`/`--key`, `--cpu`, `--share`, `--memory`, `--storage`, `--volume`, `--disk`, `--numa-cpu`, `--numa-memory`, `--label`, `--test`, `--extra-resources` |
| `node get` | `<node name>` | |
| `node remove` | `<node name>` | |
| `node set`, `node update` | `<node name>` | `--cpu`, `--memory`, `--storage`, `--volume`, `--disk`, `--rm-disk`, `--numa-cpu`, `--numa-memory`, `--label`, `--delta`, `--endpoint`, `--ca`/`--cert`/`--key`, `--mark-workloads-down`, `--extra-resources` |
| `node up` | `<node name>` | |
| `node down` | `<node name>` | `--check`, `--check-timeout` (default `20`) |
| `node workloads`, `node containers` | `<node name>` | `--label` |
| `node resource` | `<node name>` | `--fix` |
| `node set-status` | `<node name>` | `--ttl 180`, `--interval` |
| `node watch-status` | | |

When `--endpoint` is omitted, `node add` derives `tcp://<local ipv4>:2376` — or port `2375` when no
CA certificate is found — and looks for TLS material under `/etc/docker/tls/`.

`--delta` on `node set` makes every value relative to the current one, so
`--memory -10G --delta` subtracts ten gigabytes instead of setting ten.

`node set-status --interval N` turns the command into a heartbeat loop that reports the node alive
every `N` seconds; without it the status is set once.

## workload

| Command | Arguments | Notable options |
|---|---|---|
| `workload deploy` | `<spec file uri>` | `--pod`, `--entry`, `--image` (all required), `--node`, `--count`, `--network`, `--cpu`/`--cpu-request`/`--cpu-limit`, `--memory*`, `--storage*`, `--env`, `--nodelabel`, `--deploy-strategy`, `--user`, `--file`, `--after-create`, `--nodes-limit`, `--auto-replace`, `--cpu-bind`, `--ignore-hook`, `--raw-args`, `--extra-resources`, `--dry-run` |
| `workload replace` | `<spec file uri>` | `--entry`, `--image` (required), `--pod`, `--node`, `--count`, `--network`, `--network-inherit`, `--env`, `--user`, `--label`, `--file`, `--copy`, `--after-create`, `--ignore-hook`, `--debug` |
| `workload get` | `<workload id>...` | |
| `workload list` | `[appname]` | `--entry`, `--node`, `--pod`, `--label`, `--limit`, `--match-ip`, `--skip-ip`, `--statistics` |
| `workload start`/`stop`/`restart` | `<workload id>...` | `--force` |
| `workload remove` | `<workload id>...` | `--force` |
| `workload realloc` | `<workload id>` | `--cpu*`, `--memory*`, `--storage*`, `--volumes-request`, `--volumes-limit`, `--cpu-bind`, `--cpu-unbind`, `--extra-resources` |
| `workload dissociate` | `<workload id>...` | `--node` to take every workload on a node; returns the resources to eru without removing the workload. |
| `workload exec` | `<workload id> -- cmd...` | `--interactive`, `--env`, `--workdir` |
| `workload logs` | `<workload id>` | `--tail`, `--since`, `--until`, `--follow` |
| `workload get-status` | `<workload id>...` | |
| `workload set-status` | `<workload id>...` | `--running`, `--healthy`, `--ttl`, `--network name=ip`, `--extension` |
| `workload copy` | `<workload id>:path1,path2` | `--dir` (default `/tmp`) |
| `workload send` | `<workload id>...` | `--file src:dst[:mode[:uid:gid]]` |
| `workload sendlarge` | `<workload id>...` | `--file src:dst[:mode[:uid:gid]]`, one file per call, streamed in chunks |

`workload` also answers to the alias `container`.

`--cpu`, `--memory` and `--storage` are shortcuts that set the matching request and limit to the
same value; give `--cpu-request`/`--cpu-limit` explicitly when they must differ.

`--dry-run` on `deploy` prints the capacity core calculated and deploys nothing.
`--auto-replace` deploys when no workload of that application and entrypoint exists yet, and
replaces otherwise; the replacement keeps the old workload's network unless `--network` is given.
`deploy` defaults `--network` to `host`.

Everything after the workload id in `workload exec` is the remote command, so flags of the remote
program are passed through untouched:

```shell
eru-cli workload exec -i <workload id> -- ls -al /tmp
```

`workload copy` writes one tar per workload into `--dir`, named
`<short id>-<workload name>-<timestamp>.tar`; an existing file of that name is left alone. Each
argument must read `<workload id>:<path>[,<path>...]` — anything else is rejected rather than
skipped. `workload sendlarge` rejects an empty source file.

## image

| Command | Arguments | Notable options |
|---|---|---|
| `image build` | `<spec file uri>` | `--name` (required), `--tag`, `--raw`, `--exist`, `--user`, `--uid`, `--stop-signal`, `--platform` |
| `image list`, `image ls` | | `--pod` or `--node` (one is required), `--filter` |
| `image cache` | `<image>...` | `--pod`, `--node` |
| `image remove` | `<image>...` | `--pod`, `--node`, `--prune` |

`image build` has three modes. By default the argument is a build spec (see
[Spec formats](specs.md)) and core clones the repository itself. With `--raw` the argument is a
local directory that is streamed to core as a tar. With `--exist` the argument is the id of an
existing workload that is committed into an image. `--raw` and `--exist` are mutually exclusive.

## network

| Command | Arguments | Notable options |
|---|---|---|
| `network connect` | `<workload id>...` | `--network` (required), `--ipv4`, `--ipv6` |
| `network disconnect` | `<workload id>...` | `--network` (required) |

## status

```
eru-cli status [appname] [--entry ...] [--node ...] [--label a=1]
```

Follows the workload status stream and prints a line whenever a workload starts, stops, becomes
unhealthy or has its status expire. It runs until interrupted; `SIGINT` and `SIGTERM` end it
cleanly.

## lambda

```
eru-cli lambda [options] -- cmd1 cmd2 cmd3
```

Runs a command inside a freshly created workload, streams its output back and exits with the
command's exit code. Everything after the first positional argument is the remote command line.

| Option | Default | Meaning |
|---|---|---|
| `--pod`, `--node` | | Where to run. |
| `--image` | `alpine:latest` | Base image. |
| `--name` | | Entrypoint name of the lambda. |
| `--count` | `1` | How many copies to run; the cli waits for all of them. |
| `--cpu`, `--cpu-request` | `1`, `0` | CPU limit and request. |
| `--memory`, `--memory-request` | `512M` | Memory limit and request. |
| `--storage`, `--storage-request` | | Storage limit and request. |
| `--volume`, `--volume-request` | | Volume limit and request, repeatable. |
| `--env` | | `KEY=value`, repeatable. |
| `--file` | | `src:dst`, repeatable. |
| `--working-dir` | `/` | Working directory. |
| `--user` | `root` | User inside the workload. |
| `--privileged`, `-p` | off | Extended privileges. |
| `--stdin`, `-s` | off | Attach stdin and put the terminal in raw mode. |
| `--async`, `--async-timeout` | off, `30` | Return immediately and let core reap the workload. |
| `--deploy-strategy` | `auto` | `auto`, `fill`, `each`, `global`, `drained` or `dummy`. |
| `--workload-id` | off | Prefix every output line with the workload id. |

```shell
eru-cli lambda --pod dev --image golang:1.27 --cpu 2 --memory 2G -- go version
```
