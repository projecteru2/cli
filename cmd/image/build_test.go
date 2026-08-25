package image

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
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
    submodule: true
    security: true
    stop_signal: SIGUSR1
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

func TestGenerateBuildOptions(t *testing.T) {
	specs := runBuildCommand(t, []string{"image", "build", "--name", "app", writeBuildSpec(t)}).Builds

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
	if test.StopSignal != "SIGUSR1" {
		t.Errorf("test stop_signal: got %q, want %q", test.StopSignal, "SIGUSR1")
	}
	if !test.Submodule || !test.Security {
		t.Errorf("test stage: got submodule=%v security=%v, want both true", test.Submodule, test.Security)
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

func TestGenerateBuildOptionsStopSignalFlagWins(t *testing.T) {
	specs := runBuildCommand(t, []string{"image", "build", "--name", "app", "--stop-signal", "SIGTERM", writeBuildSpec(t)}).Builds

	for _, name := range slices.Sorted(maps.Keys(specs.Builds)) {
		if got := specs.Builds[name].StopSignal; got != "SIGTERM" {
			t.Errorf("%s stop_signal: got %q, want %q", name, got, "SIGTERM")
		}
	}
}

func writeBuildSpec(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "builds.yaml")
	if err := os.WriteFile(path, []byte(buildSample), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return path
}

func runBuildCommand(t *testing.T, args []string) *corepb.BuildImageOptions {
	t.Helper()
	c := Command()
	idx := slices.IndexFunc(c.Commands, func(sub *cli.Command) bool { return sub.Name == "build" })
	if idx < 0 {
		t.Fatal("subcommand build not found")
	}

	var opts *corepb.BuildImageOptions
	c.Commands[idx].Action = func(ctx context.Context, cmd *cli.Command) error {
		var err error
		opts, err = generateBuildOptions(ctx, cmd)
		return err
	}
	if err := c.Run(t.Context(), args); err != nil {
		t.Fatalf("run: %v", err)
	}
	if opts == nil {
		t.Fatal("got nil options")
	}
	return opts
}
