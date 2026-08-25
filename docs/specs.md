---
title: Spec formats
---

# Spec formats

Two commands read a YAML file: `workload deploy` / `workload replace` take a **deploy spec**, and
`image build` takes a **build spec**. Both accept a local path or an `http://` / `https://` URL as
their single positional argument.

The build spec — and only the build spec — is rendered as a Go text template with the process
environment as its data before parsing, so `{{.CI_COMMIT_SHA}}` expands to that environment
variable. A name that is not set renders empty rather than failing.

## Deploy spec

The deploy spec describes an application and the entrypoints it can run. It is usually committed
next to the application as `specs.yaml`.

```yaml
appname: "elb"
entrypoints:
  release:
    cmd: "/usr/local/openresty/bin/openresty -p /elb/server -c /elb/conf/release.conf"
    commands:
      - /usr/local/openresty/bin/openresty
      - -p
      - /elb/server
      - -c
      - /elb/conf/release.conf
    restart: always
    dir: /elb
    publish:
      - "80/http"
    healthcheck:
      tcp_ports:
        - "80"
      http_port: "90"
      url: "/"
      code: 200
    privileged: true
    sysctls:
      net.core.somaxconn: "1024"
    log:
      type: journald
      config:
        tag: elb
    hook:
      after_start:
        - "ls -al /tmp"
      before_stop:
        - "abcd"
      force: true
volumes:
  - "/tmp:/tmp/host"
volumes_request:
  - "AUTO:/data:rw:1G"
labels:
  role: edge
dns:
  - 8.8.8.8
extra_hosts:
  - "eru:127.0.0.1"
```

### Top level

| Key | Type | Meaning |
|---|---|---|
| `appname` | string | Application name; workloads are named after it. |
| `entrypoints` | map | Named entrypoints; `--entry` selects one. |
| `volumes` | list | Volume limits bound into every workload. |
| `volumes_request` | list | Volume requests, in the same syntax. |
| `labels` | map | Labels attached to the workload. |
| `dns` | list | Extra nameservers. |
| `extra_hosts` | list | Extra `/etc/hosts` entries, `name:address`. |

### Entrypoint

| Key | Type | Meaning |
|---|---|---|
| `cmd` | string | Legacy command line, split on spaces. |
| `commands` | list | Command line as a list; wins over `cmd` when present. |
| `dir` | string | Working directory. |
| `privileged` | bool | Extended privileges. |
| `restart` | string | Restart policy passed to the engine, e.g. `always`. |
| `publish` | list | Ports to publish, `<port>/<proto>`. |
| `sysctls` | map | Sysctls to set inside the workload. |
| `log` | map | `type` (`journald`, `json-file`, `none`) and a `config` map. |
| `healthcheck` | map | `tcp_ports`, `http_port`, `url`, `code`. |
| `hook` | map | `after_start`, `before_stop` command lists and `force`. |

`force: true` under `hook` makes a failing hook fail the operation; `--ignore-hook` on the command
line overrides it for a single run.

Resources are **not** part of the deploy spec except for volumes: CPU, memory and storage come from
the command line, so the same spec can be deployed at different sizes.

```shell
eru-cli workload deploy \
  --pod dev --entry release --image projecteru2/elb:latest \
  --cpu 2 --memory 2G --count 3 \
  ./specs.yaml
```

## Build spec

The build spec drives `image build` in its default (SCM) mode: core clones the repository itself and
runs the stages in order, each stage producing a layer for the next.

```yaml
stages:
  - test
  - build
builds:
  test:
    base: alpine:3.20
    repo: git@github.com:projecteru2/cli.git
    version: HEAD
    dir: /var
    commands:
      - make test
    envs:
      TEST: "abc"
    cache:
      /tmp/testbuild: /testbuild
  build:
    base: alpine:3.20
    commands:
      - make build
    args:
      WTF: "123"
    labels:
      ERU: "1"
    artifacts:
      /go/src/app/eru-cli: /usr/bin/eru-cli
```

| Key | Type | Meaning |
|---|---|---|
| `stages` | list | Build stage names, executed in this order. |
| `builds` | map | One entry per stage name. |

Each stage takes:

| Key | Type | Meaning |
|---|---|---|
| `base` | string | Base image for the stage. |
| `repo` | string | Repository to clone; only the first stage usually needs it. |
| `version` | string | Ref to check out. |
| `submodule` | bool | Clone submodules as well. |
| `dir` | string | Working directory for the commands. |
| `commands` | list | Shell commands run in order. |
| `envs` | map | Environment for the commands. |
| `args` | map | Build arguments. |
| `labels` | map | Labels set on the produced image. |
| `cache` | map | Paths carried from this stage into the next, `source: destination`. |
| `artifacts` | map | Paths copied into the final image, `source: destination`. |
| `security` | bool | Run the stage with extended privileges. |

Keys are the lowercased protobuf field names, so they are single words: use `stopsignal`, not
`stop_signal`. `--stop-signal` on the command line overrides every stage; without it the value from
the spec is kept.

```shell
eru-cli image build --name projecteru2/cli --tag latest ./build.yaml
```

The image name and tags always come from the command line; `--tag` may be repeated and defaults to
`latest`. Two other modes bypass the spec entirely:

```shell
eru-cli image build --raw   --name myimage ./context-dir   # tar a local directory
eru-cli image build --exist --name myimage <workload id>   # commit a running workload
```
