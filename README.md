CLI
=====
![](https://github.com/projecteru2/cli/workflows/goreleaser/badge.svg)
![](https://github.com/projecteru2/cli/workflows/golangci-lint/badge.svg)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/e9a4c445afb549fea950b4353197e859)](https://www.codacy.com/gh/projecteru2/cli?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=projecteru2/cli&amp;utm_campaign=Badge_Grade)

CLI for Eru.

Modify resources for eru pods / nodes, manipulate workloads and images.

### Develop

We use `go mod` to manage requirements.  
Start developing:

```
$ git clone github.com/projecteru2/cli
$ make test binary
```

Commands' source code in `commands` dir, you can define your own commands inside. Use `make test` to test and `make build` to build. If you want to modify and build in local, you can use `make deps` to generate vendor dirs.

### Dockerized cli

Image: [projecteru2/cli](https://hub.docker.com/r/projecteru2/cli/)

```shell
docker run -it --rm \
  --net host \
  --name eru-cli \
  projecteru2/cli \
  /usr/bin/eru-cli <PARAMS>
```

### User Manual

Refer to [this document](https://github.com/projecteru2/cli/blob/master/manual.md)
