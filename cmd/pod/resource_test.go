package pod

import (
	"context"
	"slices"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
)

func TestResourceCommandDefaultsToNoFilter(t *testing.T) {
	c := Command()
	idx := slices.IndexFunc(c.Commands, func(cmd *cli.Command) bool { return cmd.Name == "resource" })
	if idx < 0 {
		t.Fatal("resource command not found")
	}
	var filter string
	c.Commands[idx].Action = func(_ context.Context, cmd *cli.Command) error {
		filter = cmd.String("filter")
		return nil
	}
	if err := c.Run(t.Context(), []string{"pod", "resource", "dev"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if filter != "" {
		t.Errorf("got %q, want no filter", filter)
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		wantName  string
		wantOp    string
		wantValue string
	}{
		{name: "percent", expr: "cpu > 40%", wantName: "cpu", wantOp: ">", wantValue: "40%"},
		{name: "fraction", expr: "memory<=0.4", wantName: "memory", wantOp: "<=", wantValue: "0.4"},
		{name: "equality", expr: "volume==1", wantName: "volume", wantOp: "==", wantValue: "1"},
		{name: "unsupported word", expr: "all"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := match(tt.expr)
			if tt.wantName == "" {
				if len(got) != 0 {
					t.Fatalf("got %v, want no match", got)
				}
				return
			}
			if got["name"] != tt.wantName || got["op"] != tt.wantOp || got["value"] != tt.wantValue {
				t.Errorf("got %v, want name=%s op=%s value=%s", got, tt.wantName, tt.wantOp, tt.wantValue)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name string
		op   string
		want bool
	}{
		{name: "greater", op: ">", want: true},
		{name: "greater or equal", op: ">=", want: true},
		{name: "less", op: "<", want: false},
		{name: "less or equal", op: "<=", want: false},
		{name: "equal", op: "==", want: false},
		{name: "unknown", op: "!=", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compare(tt.op, 0.5, 0.4); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAttr(t *testing.T) {
	nr := &corepb.NodeResource{
		ResourceUsage:    `{"cpumem":{"cpu":2,"memory":512},"storage":{"storage":50,"volumes":{"/data":10}}}`,
		ResourceCapacity: `{"cpumem":{"cpu":8,"memory":2048},"storage":{"storage":200,"volumes":{"/data":40}}}`,
	}

	tests := []struct {
		name string
		attr string
		want float64
	}{
		{name: "cpu", attr: flagCPU, want: 0.25},
		{name: "memory", attr: flagMemory, want: 0.25},
		{name: "storage", attr: flagStorage, want: 0.25},
		{name: "volume", attr: "volume", want: 0.25},
		{name: "unknown", attr: "gpu", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := attr(nr, tt.attr)
			if err != nil {
				t.Fatalf("attr: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
