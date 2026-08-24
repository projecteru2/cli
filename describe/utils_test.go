package describe

import (
	"io"
	"math"
	"os"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func TestToResourcePrecent(t *testing.T) {
	tests := []struct {
		name     string
		usage    string
		capacity string
		wantCPU  map[string]float64
		wantSt   map[string]float64
	}{
		{
			name:     "cpu and memory",
			usage:    `{"cpumem":{"cpu":2,"memory":512}}`,
			capacity: `{"cpumem":{"cpu":8,"memory":2048}}`,
			wantCPU:  map[string]float64{"cpu": 0.25, "memory": 0.25},
			wantSt:   map[string]float64{},
		},
		{
			name:     "zero capacity keeps zero percent",
			usage:    `{"cpumem":{"cpu":2,"memory":512}}`,
			capacity: `{"cpumem":{"cpu":0,"memory":0}}`,
			wantCPU:  map[string]float64{"cpu": 0, "memory": 0},
			wantSt:   map[string]float64{},
		},
		{
			name:     "storage and volumes",
			usage:    `{"storage":{"storage":50,"volumes":{"/data":10,"/tmp":10}}}`,
			capacity: `{"storage":{"storage":200,"volumes":{"/data":40,"/tmp":40}}}`,
			wantCPU:  map[string]float64{},
			wantSt:   map[string]float64{"storage": 0.25, "volumes": 0.25},
		},
		{
			name:     "no volume capacity keeps zero percent",
			usage:    `{"storage":{"storage":50,"volumes":{}}}`,
			capacity: `{"storage":{"storage":200,"volumes":{}}}`,
			wantCPU:  map[string]float64{},
			wantSt:   map[string]float64{"storage": 0.25, "volumes": 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr, sr, err := ToResourcePrecent(&corepb.NodeResource{
				ResourceUsage:    tt.usage,
				ResourceCapacity: tt.capacity,
			})
			if err != nil {
				t.Fatalf("ToResourcePrecent: %v", err)
			}
			assertPercents(t, "cpumem", cr, tt.wantCPU)
			assertPercents(t, "storage", sr, tt.wantSt)
		})
	}
}

func TestToResourcePrecentInvalidJSON(t *testing.T) {
	if _, _, err := ToResourcePrecent(&corepb.NodeResource{ResourceUsage: "{"}); err == nil {
		t.Fatal("got nil, want a json error")
	}
}

func TestToChan(t *testing.T) {
	got := []int{}
	for v := range ToChan(1, 2, 3) {
		got = append(got, v)
	}
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Fatalf("got %v, want [1 2 3]", got)
	}
}

func TestToTableRows(t *testing.T) {
	rows := toTableRows([][]string{{"a", "b", "c"}, {"x"}})
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	want := [][]any{{"a", "x"}, {"b", ""}, {"c", ""}}
	for i, w := range want {
		for j := range w {
			if rows[i][j] != w[j] {
				t.Errorf("row %d col %d: got %v, want %v", i, j, rows[i][j], w[j])
			}
		}
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

func assertPercents(t *testing.T, kind string, got, want map[string]float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %v, want %v", kind, got, want)
	}
	for k, w := range want {
		v, ok := got[k]
		if !ok {
			t.Errorf("%s: missing key %s in %v", kind, k, got)
			continue
		}
		if math.IsNaN(v) || math.Abs(v-w) > 1e-9 {
			t.Errorf("%s[%s]: got %v, want %v", kind, k, v, w)
		}
	}
}
