package image

import (
	"maps"
	"slices"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"gopkg.in/yaml.v3"
)

const buildSample = `
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
`

func TestBuildsUnmarshal(t *testing.T) {
	specs := &corepb.Builds{}
	if err := yaml.Unmarshal([]byte(buildSample), specs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !slices.Equal(specs.Stages, []string{"test", "build"}) {
		t.Fatalf("stages: got %v, want [test build]", specs.Stages)
	}

	test, ok := specs.Builds["test"]
	if !ok {
		t.Fatalf("builds: got %v, want a test stage", slices.Sorted(maps.Keys(specs.Builds)))
	}
	if test.Base != "alpine:3.20" || test.Repo != "git@github.com:projecteru2/cli.git" || test.Version != "HEAD" || test.Dir != "/var" {
		t.Errorf("test stage: got base=%q repo=%q version=%q dir=%q", test.Base, test.Repo, test.Version, test.Dir)
	}
	if !slices.Equal(test.Commands, []string{"make test"}) {
		t.Errorf("test commands: got %v", test.Commands)
	}
	if !maps.Equal(test.Envs, map[string]string{"TEST": "abc"}) {
		t.Errorf("test envs: got %v", test.Envs)
	}
	if !maps.Equal(test.Cache, map[string]string{"/tmp/testbuild": "/testbuild"}) {
		t.Errorf("test cache: got %v", test.Cache)
	}

	build := specs.Builds["build"]
	if !maps.Equal(build.Args, map[string]string{"WTF": "123"}) {
		t.Errorf("build args: got %v", build.Args)
	}
	if !maps.Equal(build.Labels, map[string]string{"ERU": "1"}) {
		t.Errorf("build labels: got %v", build.Labels)
	}
	if !maps.Equal(build.Artifacts, map[string]string{"/go/src/app/eru-cli": "/usr/bin/eru-cli"}) {
		t.Errorf("build artifacts: got %v", build.Artifacts)
	}
}
