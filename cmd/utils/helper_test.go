package utils

import (
	"maps"
	"testing"
)

func TestGetNetworks(t *testing.T) {
	tests := []struct {
		name    string
		network string
		want    map[string]string
	}{
		{name: "empty", network: "", want: map[string]string{}},
		{name: "name only", network: "bridge", want: map[string]string{"bridge": ""}},
		{name: "name and ip", network: "bridge=10.0.0.2", want: map[string]string{"bridge": "10.0.0.2"}},
		{name: "too many parts", network: "a=b=c", want: map[string]string{"a=b=c": ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNetworks(tt.network); !maps.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRAMInHuman(t *testing.T) {
	tests := []struct {
		name    string
		ram     string
		want    int64
		wantErr bool
	}{
		{name: "empty", ram: "", want: 0},
		{name: "bytes", ram: "1024", want: 1024},
		{name: "kilobytes", ram: "100KB", want: 102400},
		{name: "gigabytes", ram: "1G", want: 1 << 30},
		{name: "negative", ram: "-10G", want: -(10 << 30)},
		{name: "garbage", ram: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRAMInHuman(tt.ram)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("got %d, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRAMInHuman: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSplitEquality(t *testing.T) {
	tests := []struct {
		name     string
		elements []string
		want     map[string]string
	}{
		{name: "empty", elements: nil, want: map[string]string{}},
		{name: "pairs", elements: []string{"a=1", "b=2"}, want: map[string]string{"a": "1", "b": "2"}},
		{name: "keeps the rest of the value", elements: []string{"a=1=2"}, want: map[string]string{"a": "1=2"}},
		{name: "drops keys without a value", elements: []string{"a", "b=2"}, want: map[string]string{"b": "2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitEquality(tt.elements); !maps.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvParser(t *testing.T) {
	t.Setenv("ERU_TEST_SPEC_NAME", "elb")

	got, err := EnvParser([]byte(`appname: "{{.ERU_TEST_SPEC_NAME}}"`))
	if err != nil {
		t.Fatalf("EnvParser: %v", err)
	}
	if string(got) != `appname: "elb"` {
		t.Errorf("got %q, want %q", got, `appname: "elb"`)
	}
}

func TestEnvParserBadTemplate(t *testing.T) {
	if _, err := EnvParser([]byte("{{")); err == nil {
		t.Fatal("got nil, want a template error")
	}
}
