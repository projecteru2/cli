package describe

import (
	"io"
	"os"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func TestPods(t *testing.T) {
	pods := []*corepb.Pod{
		{Name: "dev", Desc: "development"},
		{Name: "prod", Desc: "production"},
	}

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "table",
			format: "",
			want: `┌──────┬─────────────┐
│ NAME │ DESCRIPTION │
├──────┼─────────────┤
│ dev  │ development │
│ prod │ production  │
└──────┴─────────────┘
`,
		},
		{
			name:   "json",
			format: "json",
			want: `[
  {
    "name": "dev",
    "desc": "development"
  },
  {
    "name": "prod",
    "desc": "production"
  }
]
`,
		},
		{
			name:   "yaml",
			format: "yaml",
			want: `- desc: development
  name: dev
- desc: production
  name: prod

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Format = tt.format
			defer func() { Format = "" }()

			if got := captureStdout(t, func() { Pods(pods...) }); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestPodCapacity(t *testing.T) {
	want := `Total: 7
┌──────┬──────────┐
│ NODE │ CAPACITY │
├──────┼──────────┤
│ b    │ 5        │
│ a    │ 2        │
└──────┴──────────┘
`
	got := captureStdout(t, func() {
		PodCapacity(7, map[string]int64{"a": 2, "b": 5})
	})
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestNetworks(t *testing.T) {
	want := `┌─────────┬─────────────┐
│ NAME    │ NETWORK     │
├─────────┼─────────────┤
│ bridge  │ 10.0.0.0/24 │
│ overlay │ 10.1.0.0/24 │
└─────────┴─────────────┘
`
	got := captureStdout(t, func() {
		Networks(
			&corepb.Network{Name: "bridge", Subnets: []string{"10.0.0.0/24"}},
			&corepb.Network{Name: "overlay", Subnets: []string{"10.1.0.0/24"}},
		)
	})
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestWorkloadsStatistics(t *testing.T) {
	tests := []struct {
		name      string
		workloads []*corepb.Workload
		want      string
	}{
		{
			name: "cpumem and storage",
			workloads: []*corepb.Workload{
				{Resources: `{"cpumem":{"cpu_request":1.5,"memory_request":1024},"storage":{"storage_request":2048}}`},
				{Resources: `{"cpumem":{"cpu_request":0.5,"memory_request":512},"storage":{"storage_request":512}}`},
			},
			want: `{
  "CPUs": 2,
  "Memory": 1536,
  "Storage": 2560
}
`,
		},
		{
			name: "workload without a storage plugin",
			workloads: []*corepb.Workload{
				{Resources: `{"cpumem":{"cpu_request":1,"memory_request":64}}`},
			},
			want: `{
  "CPUs": 1,
  "Memory": 64,
  "Storage": 0
}
`,
		},
		{
			name:      "workload without any resources",
			workloads: []*corepb.Workload{{Resources: `{}`}},
			want: `{
  "CPUs": 0,
  "Memory": 0,
  "Storage": 0
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Format = "json"
			t.Cleanup(func() { Format = "" })

			if got := captureStdout(t, func() { WorkloadsStatistics(tt.workloads...) }); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	out := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		out <- string(b)
	}()

	f()

	os.Stdout = orig
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return <-out
}
