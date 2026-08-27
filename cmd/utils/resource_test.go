package utils

import (
	"context"
	"encoding/json"
	"maps"
	"testing"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/urfave/cli/v3"
)

func TestEncodeResources(t *testing.T) {
	tests := []struct {
		name      string
		extra     string
		resources resourcetypes.Resources
		want      map[string]map[string]any
		wantErr   bool
	}{
		{
			name:      "plugins only",
			resources: resourcetypes.Resources{"cpumem": {"cpu-limit": 2.0}},
			want:      map[string]map[string]any{"cpumem": {"cpu-limit": 2.0}},
		},
		{
			name:      "extra resource is added",
			extra:     `{"gpu":{"count":1}}`,
			resources: resourcetypes.Resources{"cpumem": {"cpu-limit": 2.0}},
			want: map[string]map[string]any{
				"cpumem": {"cpu-limit": 2.0},
				"gpu":    {"count": 1.0},
			},
		},
		{
			name:      "extra resource never replaces an encoded plugin",
			extra:     `{"cpumem":{"cpu-limit":99}}`,
			resources: resourcetypes.Resources{"cpumem": {"cpu-limit": 2.0}},
			want:      map[string]map[string]any{"cpumem": {"cpu-limit": 2.0}},
		},
		{
			name:      "untouched plugins are omitted",
			resources: resourcetypes.Resources{"cpumem": {}, "storage": {"storage-limit": 1.0}},
			want:      map[string]map[string]any{"storage": {"storage-limit": 1.0}},
		},
		{
			name:      "extra resources fill an untouched plugin",
			extra:     `{"cpumem":{"cpu-limit":3}}`,
			resources: resourcetypes.Resources{"cpumem": {}},
			want:      map[string]map[string]any{"cpumem": {"cpu-limit": 3.0}},
		},
		{
			name:      "malformed extra resources",
			extra:     "{",
			resources: resourcetypes.Resources{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeResources(commandWithExtraResources(t, tt.extra), tt.resources)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("got %v, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("EncodeResources: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for plugin, want := range tt.want {
				params := map[string]any{}
				if err := json.Unmarshal(got[plugin], &params); err != nil {
					t.Fatalf("decode %s: %v", plugin, err)
				}
				if !maps.Equal(params, want) {
					t.Errorf("%s: got %v, want %v", plugin, params, want)
				}
			}
		})
	}
}

func TestCompactParamsLetsExtraResourcesFillStorage(t *testing.T) {
	storage := resourcetypes.RawParams{
		"storage-request": int64(0),
		"storage-limit":   int64(0),
		"volumes-request": []string(nil),
		"volumes-limit":   []string{},
	}
	if got := CompactParams(storage); len(got) != 0 {
		t.Errorf("got %v, want every zero value dropped", got)
	}

	kept := resourcetypes.RawParams{
		"storage-request": int64(0),
		"storage-limit":   int64(1073741824),
		"volumes-limit":   []string{"/data0:1G"},
	}
	want := resourcetypes.RawParams{
		"storage-limit": int64(1073741824),
		"volumes-limit": []string{"/data0:1G"},
	}
	if got := CompactParams(kept); !maps.EqualFunc(got, want, func(a, b any) bool {
		aj, _ := json.Marshal(a)
		bj, _ := json.Marshal(b)
		return string(aj) == string(bj)
	}) {
		t.Errorf("got %v, want %v", got, want)
	}

	cmd := commandWithExtraResources(t, `{"resource-storage":{"storage":2147483648}}`)
	encoded, err := EncodeResources(cmd, resourcetypes.Resources{"resource-storage": CompactParams(storage)})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded["resource-storage"]) != `{"storage":2147483648}` {
		t.Errorf("got %s, want the extra request once no storage flag is set", encoded["resource-storage"])
	}
}

func commandWithExtraResources(t *testing.T, extra string) *cli.Command {
	t.Helper()
	var parsed *cli.Command
	cmd := &cli.Command{
		Name:  "test",
		Flags: []cli.Flag{&cli.StringFlag{Name: "extra-resources"}},
		Action: func(_ context.Context, c *cli.Command) error {
			parsed = c
			return nil
		},
	}
	args := []string{"test"}
	if extra != "" {
		args = append(args, "--extra-resources", extra)
	}
	if err := cmd.Run(t.Context(), args); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return parsed
}
