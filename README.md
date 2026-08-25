# cli

`cli` is the Eru command-line client. It builds `eru-cli`, a single static binary that talks to
[eru core](https://github.com/projecteru2/core) over gRPC and drives the whole cluster from a shell:
pods, nodes, workloads, images, networks and one-shot lambdas.

**Documentation: [projecteru2.github.io/cli](https://projecteru2.github.io/cli/)** (source in [`docs/`](docs/))

## Highlights

- Eight command groups — `core`, `pod`, `node`, `workload`, `image`, `network`, `status`, `lambda` —
  covering every RPC eru core exposes.
- Table, JSON or YAML output for every read command, selected once with `--output`.
- Interactive streams: `workload exec` and `lambda` attach a raw terminal, forward `SIGWINCH`
  and return the remote exit code as their own.
- Script-friendly exit status: a batch command that acts on many workloads exits non-zero when any
  one of them fails, instead of hiding the failure in its log.
- Deploy and build from a YAML spec that lives next to the application, read from a path or an
  HTTP URL, with `${ENV}` template expansion.

## Quick start

```shell
# talk to a core instance
export ERU=127.0.0.1:5001

eru-cli core info                       # core version and identity
eru-cli pod list                        # pods known to core
eru-cli pod nodes --filter up <pod>     # live nodes in a pod
eru-cli workload list <appname>         # workloads of an application

# run a throwaway command in the cluster
eru-cli lambda --pod <pod> --image alpine:latest -- echo hello

# deploy from a spec
eru-cli workload deploy --pod <pod> --entry web --image <image> ./specs.yaml
```

Every read command takes `--output json` or `--output yaml` when you want to pipe the result:

```shell
eru-cli --output json pod nodes <pod> | jq '.[].name'
```

See [docs/cli.md](docs/cli.md) for the full command reference and
[docs/specs.md](docs/specs.md) for the deploy and build spec formats.

## Related projects

- [core](https://github.com/projecteru2/core) — the gRPC scheduler this client drives.
- [agent](https://github.com/projecteru2/agent) — per-node agent reporting workload health.
- [yavirt](https://github.com/projecteru2/yavirt) — virtual machine engine for eru.
- [resource-storage](https://github.com/projecteru2/resource-storage),
  [resource-gpu](https://github.com/projecteru2/resource-gpu) — resource plugins whose parameters
  this client passes through with `--extra-resources`.

## Development

```shell
make build       # build ./eru-cli
make test        # go vet + go test -race
make lint        # golangci-lint on linux and darwin
make fmt         # gofumpt + goimports
make help        # every target
```

## License

MIT, see [LICENSE](LICENSE).
