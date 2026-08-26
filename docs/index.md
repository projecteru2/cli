---
title: eru-cli
---

# eru-cli

`eru-cli` is the command-line client of [Eru](https://github.com/projecteru2/core). It is a single
static binary that speaks gRPC to eru core and exposes every cluster operation as a shell command:
managing pods and nodes, deploying and replacing workloads, building and caching images, wiring
networks, watching status streams and running one-shot lambdas.

```
   you                    eru-cli                    eru core                  node
    |                        |                          |                        |
    |  eru-cli pod nodes ... |                          |                        |
    |----------------------->|  gRPC (ERU=host:5001)    |                        |
    |                        |------------------------->|  engine api            |
    |                        |                          |----------------------->|
    |                        |<-------------------------|                        |
    |   table / json / yaml  |     stream of messages   |                        |
    |<-----------------------|                          |                        |
```

The client is stateless. It holds no configuration file: the target core, the credentials and the
output format come from flags or environment variables on every invocation.

## Guides

- [Installation](installation.md) — binaries, container image and building from source.
- [Command reference](cli.md) — global options and every command group.
- [Spec formats](specs.md) — the deploy spec and the build spec, with worked examples.

## Related projects

- [core](https://github.com/projecteru2/core) — the scheduler this client drives.
- [agent](https://github.com/projecteru2/agent) — per-node agent reporting workload health.
