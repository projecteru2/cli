package describe

import (
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func TestCore(t *testing.T) {
	info := &corepb.CoreInfo{
		Version:       "v1",
		Revison:       "abc",
		BuildAt:       "now",
		GolangVersion: "go1.27",
		OsArch:        "linux/amd64",
		Identifier:    "id",
	}

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{name: "table", format: "", want: `┌────────────────┬─────────────┐
│ NAME           │ DESCRIPTION │
├────────────────┼─────────────┤
│ Version        │ v1          │
│ Git hash       │ abc         │
│ Built          │ now         │
│ Golang version │ go1.27      │
│ OS/Arch        │ linux/amd64 │
│ Identifier     │ id          │
└────────────────┴─────────────┘
`},
		{name: "json", format: "json", want: `{
  "version": "v1",
  "revison": "abc",
  "build_at": "now",
  "golang_version": "go1.27",
  "os_arch": "linux/amd64",
  "identifier": "id"
}
`},
		{name: "yaml", format: "yaml", want: `build_at: now
golang_version: go1.27
identifier: id
os_arch: linux/amd64
revison: abc
version: v1

`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Format = tt.format
			t.Cleanup(func() { Format = "" })

			if got := captureStdout(t, func() { Core(info) }); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}
