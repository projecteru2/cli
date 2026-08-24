package describe

import (
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
			t.Cleanup(func() { Format = "" })

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
