package types

import (
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
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
    hook:
      after_start:
        - "ls -al /tmp"
      before_stop:
        - "abcd"
      force: true
  worker:
    commands:
      - /bin/worker
      - --loop
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

func TestSpecsUnmarshal(t *testing.T) {
	specs := &Specs{}
	if err := yaml.Unmarshal([]byte(specSample), specs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if specs.Appname != "elb" {
		t.Errorf("appname: got %q, want %q", specs.Appname, "elb")
	}
	if !slices.Equal(specs.Volumes, []string{"/tmp:/tmp/host"}) {
		t.Errorf("volumes: got %v", specs.Volumes)
	}
	if !slices.Equal(specs.VolumesRequest, []string{"AUTO:/data:rw:1G"}) {
		t.Errorf("volumes_request: got %v", specs.VolumesRequest)
	}
	if specs.Labels["role"] != "edge" {
		t.Errorf("labels: got %v", specs.Labels)
	}
	if !slices.Equal(specs.DNS, []string{"8.8.8.8"}) {
		t.Errorf("dns: got %v", specs.DNS)
	}
	if !slices.Equal(specs.ExtraHosts, []string{"eru:127.0.0.1"}) {
		t.Errorf("extra_hosts: got %v", specs.ExtraHosts)
	}

	release, ok := specs.Entrypoints["release"]
	if !ok {
		t.Fatalf("entrypoints: got %v, want a release entry", specs.Entrypoints)
	}
	if release.Dir != "/elb" || release.Restart != "always" || !release.Privileged {
		t.Errorf("release: got dir=%q restart=%q privileged=%v", release.Dir, release.Restart, release.Privileged)
	}
	if release.HealthCheck == nil || release.HealthCheck.HTTPCode != 200 || release.HealthCheck.HTTPPort != "90" {
		t.Errorf("healthcheck: got %+v", release.HealthCheck)
	}
	if release.Hook == nil || !release.Hook.Force || !slices.Equal(release.Hook.AfterStart, []string{"ls -al /tmp"}) {
		t.Errorf("hook: got %+v", release.Hook)
	}
	if !slices.Equal(release.Publish, []string{"80/http"}) {
		t.Errorf("publish: got %v", release.Publish)
	}
}

func TestEntrypointGetCommands(t *testing.T) {
	specs := &Specs{}
	if err := yaml.Unmarshal([]byte(specSample), specs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	tests := []struct {
		name  string
		entry string
		want  []string
	}{
		{
			name:  "legacy cmd is split on spaces",
			entry: "release",
			want:  []string{"/usr/local/openresty/bin/openresty", "-p", "/elb/server"},
		},
		{
			name:  "commands win over cmd",
			entry: "worker",
			want:  []string{"/bin/worker", "--loop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := specs.Entrypoints[tt.entry].GetCommands(); !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
