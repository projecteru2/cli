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
