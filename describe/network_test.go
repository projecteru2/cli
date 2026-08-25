package describe

import (
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func TestNetworks(t *testing.T) {
	networks := []*corepb.Network{
		{Name: "bridge", Subnets: []string{"10.0.0.0/24"}},
		{Name: "overlay", Subnets: []string{"10.1.0.0/24", "fd00::/64"}},
	}

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{name: "table", format: "", want: `┌─────────┬───────────────────────┐
│ NAME    │ NETWORK               │
├─────────┼───────────────────────┤
│ bridge  │ 10.0.0.0/24           │
│ overlay │ 10.1.0.0/24,fd00::/64 │
└─────────┴───────────────────────┘
`},
		{name: "json", format: "json", want: `[
  {
    "name": "bridge",
    "subnets": [
      "10.0.0.0/24"
    ]
  },
  {
    "name": "overlay",
    "subnets": [
      "10.1.0.0/24",
      "fd00::/64"
    ]
  }
]
`},
		{name: "yaml", format: "yaml", want: `- name: bridge
  subnets:
  - 10.0.0.0/24
- name: overlay
  subnets:
  - 10.1.0.0/24
  - fd00::/64

`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Format = tt.format
			t.Cleanup(func() { Format = "" })

			if got := captureStdout(t, func() { Networks(networks...) }); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}
