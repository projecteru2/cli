package describe

import (
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func TestWorkloads(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{name: "table", format: "", want: `┌─────────────────────────────┬────────────────────┬──────────────────────┐
│ NAME/ID/POD/NODE/PRIVILEGED │ NETWORKS           │ CPUMEM               │
├─────────────────────────────┼────────────────────┼──────────────────────┤
│ app_web_1                   │ bridge: 172.17.0.2 │ cpu_request: 1.5     │
│ cid1                        │ host: 10.0.0.1:80  │ memory_request: 1024 │
│ dev                         │      10.0.0.1:443  │                      │
│ node1                       │                    │                      │
│ Privileged: true            │                    │                      │
└─────────────────────────────┴────────────────────┴──────────────────────┘
`},
		{name: "json", format: "json", want: `[
  {
    "id": "cid1",
    "podname": "dev",
    "nodename": "node1",
    "name": "app_web_1",
    "privileged": true,
    "publish": {
      "host": "10.0.0.1:80,10.0.0.1:443"
    },
    "status": {
      "networks": {
        "bridge": "172.17.0.2",
        "host": "10.0.0.1"
      }
    },
    "resources": "{\"cpumem\":{\"cpu_request\":1.5,\"memory_request\":1024}}"
  }
]
`},
		{name: "yaml", format: "yaml", want: `- id: cid1
  name: app_web_1
  nodename: node1
  podname: dev
  privileged: true
  publish:
    host: 10.0.0.1:80,10.0.0.1:443
  resources: '{"cpumem":{"cpu_request":1.5,"memory_request":1024}}'
  status:
    networks:
      bridge: 172.17.0.2
      host: 10.0.0.1

`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Format = tt.format
			t.Cleanup(func() { Format = "" })

			if got := captureStdout(t, func() { Workloads(testWorkloads()...) }); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestWorkloadsStatistics(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		workloads []*corepb.Workload
		want      string
	}{
		{
			name:   "cpumem and storage",
			format: "json",
			workloads: []*corepb.Workload{
				{Resources: `{"cpumem":{"cpu_request":1.5,"memory_request":1024},"storage":{"storage_request":2048}}`},
				{Resources: `{"cpumem":{"cpu_request":0.5,"memory_request":512},"storage":{"storage_request":512}}`},
			},
			want: `{
  "cpus": 2,
  "memory": 1536,
  "storage": 2560
}
`,
		},
		{
			name:   "workload without a storage plugin",
			format: "json",
			workloads: []*corepb.Workload{
				{Resources: `{"cpumem":{"cpu_request":1,"memory_request":64}}`},
			},
			want: `{
  "cpus": 1,
  "memory": 64,
  "storage": 0
}
`,
		},
		{
			name:      "workload without any resources",
			format:    "json",
			workloads: []*corepb.Workload{{Resources: `{}`}},
			want: `{
  "cpus": 0,
  "memory": 0,
  "storage": 0
}
`,
		},
		{
			name:      "workload with unparsable resources",
			format:    "json",
			workloads: []*corepb.Workload{{Resources: `{`}},
			want: `{
  "cpus": 0,
  "memory": 0,
  "storage": 0
}
`,
		},
		{
			name:   "yaml",
			format: "yaml",
			workloads: []*corepb.Workload{
				{Resources: `{"cpumem":{"cpu_request":1.5,"memory_request":1024},"storage":{"storage_request":2048}}`},
			},
			want: `cpus: 1.5
memory: 1024
storage: 2048

`,
		},
		{
			name:   "table",
			format: "",
			workloads: []*corepb.Workload{
				{Resources: `{"cpumem":{"cpu_request":1.5,"memory_request":1024},"storage":{"storage_request":2048}}`},
			},
			want: `┌──────────┬────────┬─────────┐
│ CPUS     │ MEMORY │ STORAGE │
├──────────┼────────┼─────────┤
│ 1.500000 │ 1024   │ 2048    │
└──────────┴────────┴─────────┘
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Format = tt.format
			t.Cleanup(func() { Format = "" })

			if got := captureStdout(t, func() { WorkloadsStatistics(tt.workloads...) }); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestWorkloadStatuses(t *testing.T) {
	statuses := []*corepb.WorkloadStatus{
		{
			Id:        "cid1",
			Running:   true,
			Healthy:   true,
			Networks:  map[string]string{"host": "10.0.0.1", "bridge": "172.17.0.2"},
			Extension: []byte(`{"zone":"a","az":"b"}`),
		},
	}
	want := `┌──────┬───────────────┬────────────────────┬────────────┐
│ ID   │ STATUS        │ NETWORKS           │ EXTENSIONS │
├──────┼───────────────┼────────────────────┼────────────┤
│ cid1 │ Running: true │ bridge: 172.17.0.2 │ az: b      │
│      │ Healthy: true │ host: 10.0.0.1     │ zone: a    │
└──────┴───────────────┴────────────────────┴────────────┘
`

	if got := captureStdout(t, func() { WorkloadStatuses(statuses...) }); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestWorkloadsPluginColumns(t *testing.T) {
	workloads := []*corepb.Workload{
		{Name: "w1", Id: "c1", Resources: `{"cpumem":{"cpu_request":1}}`},
		{Name: "w2", Id: "c2", Resources: `{"cpumem":{"cpu_request":2},"storage":{"storage_request":10}}`},
	}
	want := `┌─────────────────────────────┬──────────┬────────────────┬─────────────────────┐
│ NAME/ID/POD/NODE/PRIVILEGED │ NETWORKS │ CPUMEM         │ STORAGE             │
├─────────────────────────────┼──────────┼────────────────┼─────────────────────┤
│ w1                          │          │ cpu_request: 1 │                     │
│ c1                          │          │                │                     │
│                             │          │                │                     │
│                             │          │                │                     │
│ Privileged: false           │          │                │                     │
├─────────────────────────────┼──────────┼────────────────┼─────────────────────┤
│ w2                          │          │ cpu_request: 2 │ storage_request: 10 │
│ c2                          │          │                │                     │
│                             │          │                │                     │
│                             │          │                │                     │
│ Privileged: false           │          │                │                     │
└─────────────────────────────┴──────────┴────────────────┴─────────────────────┘
`

	if got := captureStdout(t, func() { Workloads(workloads...) }); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestWorkloadsNetworks(t *testing.T) {
	tests := []struct {
		name     string
		workload *corepb.Workload
		want     string
	}{
		{
			name: "published network missing from the status",
			workload: &corepb.Workload{
				Name: "w1", Id: "c1",
				Publish: map[string]string{"host": "10.0.0.1:80"},
				Status:  &corepb.WorkloadStatus{Networks: map[string]string{"bridge": "172.17.0.2"}},
			},
			want: `┌─────────────────────────────┬────────────────────┐
│ NAME/ID/POD/NODE/PRIVILEGED │ NETWORKS           │
├─────────────────────────────┼────────────────────┤
│ w1                          │ bridge: 172.17.0.2 │
│ c1                          │ host: 10.0.0.1:80  │
│                             │                    │
│                             │                    │
│ Privileged: false           │                    │
└─────────────────────────────┴────────────────────┘
`,
		},
		{
			name: "published network without any status",
			workload: &corepb.Workload{
				Name: "w1", Id: "c1",
				Publish: map[string]string{"host": "10.0.0.1:80"},
			},
			want: `┌─────────────────────────────┬───────────────────┐
│ NAME/ID/POD/NODE/PRIVILEGED │ NETWORKS          │
├─────────────────────────────┼───────────────────┤
│ w1                          │ host: 10.0.0.1:80 │
│ c1                          │                   │
│                             │                   │
│                             │                   │
│ Privileged: false           │                   │
└─────────────────────────────┴───────────────────┘
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := captureStdout(t, func() { Workloads(tt.workload) }); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func testWorkloads() []*corepb.Workload {
	return []*corepb.Workload{
		{
			Name:       "app_web_1",
			Id:         "cid1",
			Podname:    "dev",
			Nodename:   "node1",
			Privileged: true,
			Publish:    map[string]string{"host": "10.0.0.1:80,10.0.0.1:443"},
			Status: &corepb.WorkloadStatus{
				Networks: map[string]string{"host": "10.0.0.1", "bridge": "172.17.0.2"},
			},
			Resources: `{"cpumem":{"cpu_request":1.5,"memory_request":1024}}`,
		},
	}
}
