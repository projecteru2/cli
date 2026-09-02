package describe

import (
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func TestNodes(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{name: "table", format: "", want: `┌───────┬─────────────────────┬─────────────────┬──────────────┐
│ NAME  │ ENDPOINT            │ STATUS          │ CPUMEM       │
├───────┼─────────────────────┼─────────────────┼──────────────┤
│ node1 │ tcp://10.0.0.1:2376 │ UP              │ Capacity:    │
│       │                     │ bypass false    │              │
│       │                     │ available true  │              │
│       │                     │ test false      │              │
│       │                     │                 │ cpu: 8       │
│       │                     │                 │ memory: 2048 │
│       │                     │                 │ ------------ │
│       │                     │                 │ Usage:       │
│       │                     │                 │ cpu: 2       │
│       │                     │                 │ memory: 512  │
├───────┼─────────────────────┼─────────────────┼──────────────┤
│ node2 │ tcp://10.0.0.2:2376 │ DOWN            │ Capacity:    │
│       │                     │ bypass true     │              │
│       │                     │ available false │              │
│       │                     │ test true       │              │
│       │                     │                 │ cpu: 4       │
│       │                     │                 │ memory: 1024 │
│       │                     │                 │ ------------ │
│       │                     │                 │ Usage:       │
│       │                     │                 │ cpu: 4       │
│       │                     │                 │ memory: 1024 │
└───────┴─────────────────────┴─────────────────┴──────────────┘
`},
		{name: "json", format: "json", want: `[
  {
    "name": "node1",
    "endpoint": "tcp://10.0.0.1:2376",
    "available": true,
    "info": "{\"ID\":\"ABC\",\"NCPU\":8}",
    "resource_capacity": "{\"cpumem\":{\"cpu\":8,\"memory\":2048}}",
    "resource_usage": "{\"cpumem\":{\"cpu\":2,\"memory\":512}}"
  },
  {
    "name": "node2",
    "endpoint": "tcp://10.0.0.2:2376",
    "bypass": true,
    "info": "{\"ID\":\"DEF\"}",
    "resource_capacity": "{\"cpumem\":{\"cpu\":4,\"memory\":1024}}",
    "resource_usage": "{\"cpumem\":{\"cpu\":4,\"memory\":1024}}",
    "test": true
  }
]
`},
		{name: "yaml", format: "yaml", want: `- available: true
  endpoint: tcp://10.0.0.1:2376
  info: '{"ID":"ABC","NCPU":8}'
  name: node1
  resource_capacity: '{"cpumem":{"cpu":8,"memory":2048}}'
  resource_usage: '{"cpumem":{"cpu":2,"memory":512}}'
- bypass: true
  endpoint: tcp://10.0.0.2:2376
  info: '{"ID":"DEF"}'
  name: node2
  resource_capacity: '{"cpumem":{"cpu":4,"memory":1024}}'
  resource_usage: '{"cpumem":{"cpu":4,"memory":1024}}'
  test: true

`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Format = tt.format
			t.Cleanup(func() { Format = "" })

			if got := captureStdout(t, func() { Nodes(ToChan(testNodes()...), false, false) }); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestNodesWithInfo(t *testing.T) {
	want := `┌───────┬─────────────────────┬─────────────────┬──────────────┬───────────────────────┐
│ NAME  │ ENDPOINT            │ STATUS          │ CPUMEM       │ INFO                  │
├───────┼─────────────────────┼─────────────────┼──────────────┼───────────────────────┤
│ node1 │ tcp://10.0.0.1:2376 │ UP              │ Capacity:    │ {"ID":"ABC","NCPU":8} │
│       │                     │ bypass false    │              │                       │
│       │                     │ available true  │              │                       │
│       │                     │ test false      │              │                       │
│       │                     │                 │ cpu: 8       │                       │
│       │                     │                 │ memory: 2048 │                       │
│       │                     │                 │ ------------ │                       │
│       │                     │                 │ Usage:       │                       │
│       │                     │                 │ cpu: 2       │                       │
│       │                     │                 │ memory: 512  │                       │
├───────┼─────────────────────┼─────────────────┼──────────────┼───────────────────────┤
│ node2 │ tcp://10.0.0.2:2376 │ DOWN            │ Capacity:    │ {"ID":"DEF"}          │
│       │                     │ bypass true     │              │                       │
│       │                     │ available false │              │                       │
│       │                     │ test true       │              │                       │
│       │                     │                 │ cpu: 4       │                       │
│       │                     │                 │ memory: 1024 │                       │
│       │                     │                 │ ------------ │                       │
│       │                     │                 │ Usage:       │                       │
│       │                     │                 │ cpu: 4       │                       │
│       │                     │                 │ memory: 1024 │                       │
└───────┴─────────────────────┴─────────────────┴──────────────┴───────────────────────┘
`
	got := captureStdout(t, func() { Nodes(ToChan(testNodes()...), true, false) })
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestNodeResources(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{name: "table", format: "", want: `┌───────┬────────┬────────┬─────────┬────────┬──────────┐
│ NAME  │ CPU    │ MEMORY │ STORAGE │ VOLUME │ DIFFS    │
├───────┼────────┼────────┼─────────┼────────┼──────────┤
│ node1 │ 25.00% │ 25.00% │ 25.00%  │ 25.00% │ cpu diff │
└───────┴────────┴────────┴─────────┴────────┴──────────┘
`},
		{name: "json", format: "json", want: `[
  {
    "name": "node1",
    "diffs": [
      "cpu diff"
    ],
    "resource_capacity": "{\"cpumem\":{\"cpu\":8,\"memory\":2048},\"resource-storage\":{\"storage\":200,\"volumes\":{\"/data\":40}}}",
    "resource_usage": "{\"cpumem\":{\"cpu\":2,\"memory\":512},\"resource-storage\":{\"storage\":50,\"volumes\":{\"/data\":10}}}"
  }
]
`},
		{name: "yaml", format: "yaml", want: `- diffs:
  - cpu diff
  name: node1
  resource_capacity: '{"cpumem":{"cpu":8,"memory":2048},"resource-storage":{"storage":200,"volumes":{"/data":40}}}'
  resource_usage: '{"cpumem":{"cpu":2,"memory":512},"resource-storage":{"storage":50,"volumes":{"/data":10}}}'

`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Format = tt.format
			t.Cleanup(func() { Format = "" })

			got := captureStdout(t, func() { NodeResources(t.Context(), ToChan(testNodeResources()...), false) })
			if got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestNodesStream(t *testing.T) {
	want := `┌───────┬─────────────────────┬────────────────┬──────────────┐
│ NAME  │ ENDPOINT            │ STATUS         │ CPUMEM       │
├───────┼─────────────────────┼────────────────┼──────────────┤
│ node1 │ tcp://10.0.0.1:2376 │ UP             │ Capacity:    │
│       │                     │ bypass false   │              │
│       │                     │ available true │              │
│       │                     │ test false     │              │
│       │                     │                │ cpu: 8       │
│       │                     │                │ memory: 2048 │
│       │                     │                │ ------------ │
│       │                     │                │ Usage:       │
│       │                     │                │ cpu: 2       │
│       │                     │                │ memory: 512  │
└───────┴─────────────────────┴────────────────┴──────────────┘
┌───────┬─────────────────────┬─────────────────┬──────────────┐
│ NAME  │ ENDPOINT            │ STATUS          │ CPUMEM       │
├───────┼─────────────────────┼─────────────────┼──────────────┤
│ node2 │ tcp://10.0.0.2:2376 │ DOWN            │ Capacity:    │
│       │                     │ bypass true     │              │
│       │                     │ available false │              │
│       │                     │ test true       │              │
│       │                     │                 │ cpu: 4       │
│       │                     │                 │ memory: 1024 │
│       │                     │                 │ ------------ │
│       │                     │                 │ Usage:       │
│       │                     │                 │ cpu: 4       │
│       │                     │                 │ memory: 1024 │
└───────┴─────────────────────┴─────────────────┴──────────────┘
`
	got := captureStdout(t, func() { Nodes(ToChan(testNodes()...), false, true) })
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestNodeStatusMessage(t *testing.T) {
	ms := []*corepb.NodeStatusStreamMessage{
		{Nodename: "node1", Podname: "dev", Alive: true},
		{Nodename: "node2", Error: "boom"},
	}

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{name: "json", format: "json", want: `[
  {
    "nodename": "node1",
    "podname": "dev",
    "alive": true
  },
  {
    "nodename": "node2",
    "error": "boom"
  }
]
`},
		{name: "yaml", format: "yaml", want: `- alive: true
  nodename: node1
  podname: dev
- error: boom
  nodename: node2

`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Format = tt.format
			t.Cleanup(func() { Format = "" })

			got := captureStdout(t, func() { NodeStatusMessage(t.Context(), ms...) })
			if got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestNodesPluginColumns(t *testing.T) {
	tests := []struct {
		name     string
		showInfo bool
		nodes    []*corepb.Node
		want     string
	}{
		{
			name:     "plugins differ between nodes",
			showInfo: true,
			nodes: []*corepb.Node{
				{
					Name: "n1", Endpoint: "e1", Available: true, Info: "INFO1",
					ResourceCapacity: `{"cpumem":{"cpu":8}}`,
					ResourceUsage:    `{"cpumem":{"cpu":2}}`,
				},
				{
					Name: "n2", Endpoint: "e2", Available: true, Info: "INFO2",
					ResourceCapacity: `{"cpumem":{"cpu":4},"gpu":{"count":2}}`,
					ResourceUsage:    `{"cpumem":{"cpu":1},"gpu":{"count":1}}`,
				},
			},
			want: `┌──────┬──────────┬────────────────┬──────────────┬──────────────┬───────┐
│ NAME │ ENDPOINT │ STATUS         │ CPUMEM       │ GPU          │ INFO  │
├──────┼──────────┼────────────────┼──────────────┼──────────────┼───────┤
│ n1   │ e1       │ UP             │ Capacity:    │              │ INFO1 │
│      │          │ bypass false   │              │              │       │
│      │          │ available true │              │              │       │
│      │          │ test false     │              │              │       │
│      │          │                │ cpu: 8       │              │       │
│      │          │                │ ------------ │              │       │
│      │          │                │ Usage:       │              │       │
│      │          │                │ cpu: 2       │              │       │
├──────┼──────────┼────────────────┼──────────────┼──────────────┼───────┤
│ n2   │ e2       │ UP             │ Capacity:    │ Capacity:    │ INFO2 │
│      │          │ bypass false   │              │              │       │
│      │          │ available true │              │              │       │
│      │          │ test false     │              │              │       │
│      │          │                │ cpu: 4       │ count: 2     │       │
│      │          │                │ ------------ │ ------------ │       │
│      │          │                │ Usage:       │ Usage:       │       │
│      │          │                │ cpu: 1       │ count: 1     │       │
└──────┴──────────┴────────────────┴──────────────┴──────────────┴───────┘
`,
		},
		{
			name: "plugin with capacity but no usage",
			nodes: []*corepb.Node{{
				Name: "n1", Endpoint: "e1", Available: true,
				ResourceCapacity: `{"cpumem":{"cpu":8},"gpu":{"count":4}}`,
				ResourceUsage:    `{"cpumem":{"cpu":2}}`,
			}},
			want: `┌──────┬──────────┬────────────────┬──────────────┬──────────────┐
│ NAME │ ENDPOINT │ STATUS         │ CPUMEM       │ GPU          │
├──────┼──────────┼────────────────┼──────────────┼──────────────┤
│ n1   │ e1       │ UP             │ Capacity:    │ Capacity:    │
│      │          │ bypass false   │              │              │
│      │          │ available true │              │              │
│      │          │ test false     │              │              │
│      │          │                │ cpu: 8       │ count: 4     │
│      │          │                │ ------------ │ ------------ │
│      │          │                │ Usage:       │ Usage:       │
│      │          │                │ cpu: 2       │              │
└──────┴──────────┴────────────────┴──────────────┴──────────────┘
`,
		},
		{
			name: "usage key without a capacity key",
			nodes: []*corepb.Node{{
				Name: "n1", Endpoint: "e1", Available: true,
				ResourceCapacity: `{"cpumem":{"cpu":8}}`,
				ResourceUsage:    `{"cpumem":{"cpu":2,"memory":512}}`,
			}},
			want: `┌──────┬──────────┬────────────────┬──────────────┐
│ NAME │ ENDPOINT │ STATUS         │ CPUMEM       │
├──────┼──────────┼────────────────┼──────────────┤
│ n1   │ e1       │ UP             │ Capacity:    │
│      │          │ bypass false   │              │
│      │          │ available true │              │
│      │          │ test false     │              │
│      │          │                │ cpu: 8       │
│      │          │                │ ------------ │
│      │          │                │ Usage:       │
│      │          │                │ cpu: 2       │
│      │          │                │ memory: 512  │
└──────┴──────────┴────────────────┴──────────────┘
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStdout(t, func() {
				if tt.showInfo {
					Nodes(ToChan(tt.nodes...), true, false)
					return
				}
				Nodes(ToChan(tt.nodes...), false, false)
			})
			if got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func testNodes() []*corepb.Node {
	return []*corepb.Node{
		{
			Name:             "node1",
			Endpoint:         "tcp://10.0.0.1:2376",
			Available:        true,
			Info:             `{"ID":"ABC","NCPU":8}`,
			ResourceCapacity: `{"cpumem":{"cpu":8,"memory":2048}}`,
			ResourceUsage:    `{"cpumem":{"cpu":2,"memory":512}}`,
		},
		{
			Name:             "node2",
			Endpoint:         "tcp://10.0.0.2:2376",
			Bypass:           true,
			Test:             true,
			Info:             `{"ID":"DEF"}`,
			ResourceCapacity: `{"cpumem":{"cpu":4,"memory":1024}}`,
			ResourceUsage:    `{"cpumem":{"cpu":4,"memory":1024}}`,
		},
	}
}

func testNodeResources() []*corepb.NodeResource {
	return []*corepb.NodeResource{
		{
			Name:             "node1",
			ResourceUsage:    `{"cpumem":{"cpu":2,"memory":512},"resource-storage":{"storage":50,"volumes":{"/data":10}}}`,
			ResourceCapacity: `{"cpumem":{"cpu":8,"memory":2048},"resource-storage":{"storage":200,"volumes":{"/data":40}}}`,
			Diffs:            []string{"cpu diff"},
		},
	}
}
