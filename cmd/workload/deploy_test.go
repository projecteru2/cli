package workload

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

const specSample = `
appname: "elb"
entrypoints:
  release:
    cmd: "/usr/local/openresty/bin/openresty -p /elb/server"
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
    hook:
      after_start:
        - "ls -al /tmp"
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
`

func TestGenerateDeployOptions(t *testing.T) {
	spec := writeSpec(t)

	opts := runDeployCommand(t, []string{
		"workload", "deploy",
		"--pod", "dev", "--entry", "release", "--image", "elb:latest",
		"--cpu", "2", "--memory", "1G", "--storage", "10G",
		"--extra-resources", `{"gpu":{"count":1}}`,
		spec,
	})

	if opts.Name != "elb" || opts.Podname != "dev" || opts.Image != "elb:latest" {
		t.Errorf("got name=%q podname=%q image=%q", opts.Name, opts.Podname, opts.Image)
	}
	if !slices.Equal(opts.Dns, []string{"8.8.8.8"}) || !slices.Equal(opts.ExtraHosts, []string{"eru:127.0.0.1"}) {
		t.Errorf("got dns=%v extra_hosts=%v", opts.Dns, opts.ExtraHosts)
	}
	if opts.Labels["role"] != "edge" {
		t.Errorf("labels: got %v", opts.Labels)
	}

	entry := opts.Entrypoint
	if entry.Name != "release" || entry.Dir != "/elb" || entry.Restart != "always" || !entry.Privileged {
		t.Errorf("entrypoint: got %+v", entry)
	}
	if !slices.Equal(entry.Commands, []string{"/usr/local/openresty/bin/openresty", "-p", "/elb/server"}) {
		t.Errorf("commands: got %v", entry.Commands)
	}
	if entry.Healthcheck == nil || entry.Healthcheck.Code != 200 || entry.Healthcheck.HttpPort != "90" {
		t.Errorf("healthcheck: got %+v", entry.Healthcheck)
	}
	if entry.Hook == nil || !entry.Hook.Force || !slices.Equal(entry.Hook.AfterStart, []string{"ls -al /tmp"}) {
		t.Errorf("hook: got %+v", entry.Hook)
	}
	if entry.Sysctls["net.core.somaxconn"] != "1024" {
		t.Errorf("sysctls: got %v", entry.Sysctls)
	}

	cpumem := decodeParams(t, opts.Resources[utils.ResourceCPUMem])
	if got := cpumem.Float64(flagCPURequest); got != 2 {
		t.Errorf("cpu-request: got %v, want 2", got)
	}
	if got := cpumem.Int64(flagMemoryLimit); got != 1<<30 {
		t.Errorf("memory-limit: got %d, want %d", got, 1<<30)
	}

	storage := decodeParams(t, opts.Resources[utils.ResourceStorage])
	if got := storage.Int64(flagStorageLimit); got != 10<<30 {
		t.Errorf("storage-limit: got %d, want %d", got, 10<<30)
	}
	if got := storage.StringSlice(flagVolumesLimit); !slices.Equal(got, []string{"/tmp:/tmp/host"}) {
		t.Errorf("volumes-limit: got %v", got)
	}
	if got := storage.StringSlice(flagVolumesRequest); !slices.Equal(got, []string{"AUTO:/data:rw:1G"}) {
		t.Errorf("volumes-request: got %v", got)
	}

	gpu := decodeParams(t, opts.Resources["gpu"])
	if got := gpu.Int64("count"); got != 1 {
		t.Errorf("extra resource gpu count: got %d, want 1", got)
	}
}

func TestGenerateDeployOptionsErrors(t *testing.T) {
	spec := writeSpec(t)

	tests := []struct {
		name string
		args []string
	}{
		{name: "no spec", args: []string{"workload", "deploy", "--entry", "release"}},
		{name: "unknown entry", args: []string{"workload", "deploy", "--entry", "nope", spec}},
		{name: "bad memory", args: []string{"workload", "deploy", "--entry", "release", "--memory", "abc", spec}},
		{name: "bad storage", args: []string{"workload", "deploy", "--entry", "release", "--storage", "abc", spec}},
		{name: "bad extra resources", args: []string{"workload", "deploy", "--entry", "release", "--extra-resources", "{", spec}},
		{name: "unknown deploy strategy", args: []string{"workload", "deploy", "--entry", "release", "--deploy-strategy", "nope", spec}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Command()
			lookupSubcommand(t, c, "deploy").Action = func(ctx context.Context, cmd *cli.Command) error {
				if _, err := generateDeployOptions(ctx, cmd); err == nil {
					t.Error("got nil, want an error")
				}
				return nil
			}
			if err := c.Run(t.Context(), tt.args); err != nil {
				t.Fatalf("run: %v", err)
			}
		})
	}
}

func writeSpec(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "specs.yaml")
	if err := os.WriteFile(path, []byte(specSample), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return path
}

func runDeployCommand(t *testing.T, args []string) *corepb.DeployOptions {
	t.Helper()
	var opts *corepb.DeployOptions
	c := Command()
	lookupSubcommand(t, c, "deploy").Action = func(ctx context.Context, cmd *cli.Command) error {
		var err error
		opts, err = generateDeployOptions(ctx, cmd)
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

func lookupSubcommand(t *testing.T, cmd *cli.Command, name string) *cli.Command {
	t.Helper()
	idx := slices.IndexFunc(cmd.Commands, func(c *cli.Command) bool { return c.Name == name })
	if idx < 0 {
		t.Fatalf("subcommand %q not found", name)
	}
	return cmd.Commands[idx]
}

func decodeParams(t *testing.T, raw []byte) resourcetypes.RawParams {
	t.Helper()
	params := resourcetypes.RawParams{}
	if err := json.Unmarshal(raw, &params); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return params
}
