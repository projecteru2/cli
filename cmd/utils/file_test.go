package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAllFiles(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "payload")
	if err := os.WriteFile(src, []byte("hello"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name     string
		files    []string
		wantDst  string
		wantMode int64
		wantUID  int
		wantGID  int
		wantSkip bool
	}{
		{name: "src and dst only", files: []string{src + ":/tmp/payload"}, wantDst: "/tmp/payload"},
		{name: "with mode", files: []string{src + ":/tmp/payload:755"}, wantDst: "/tmp/payload", wantMode: 0o755},
		{name: "with owner", files: []string{src + ":/tmp/payload:644:12:34"}, wantDst: "/tmp/payload", wantMode: 0o644, wantUID: 12, wantGID: 34},
		{name: "no dst", files: []string{src}, wantSkip: true},
		{name: "missing source", files: []string{filepath.Join(dir, "nope") + ":/tmp/nope"}, wantSkip: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadAllFiles(tt.files)
			if tt.wantSkip {
				if len(got) != 0 {
					t.Fatalf("got %v, want no files", got)
				}
				return
			}
			f, ok := got[tt.wantDst]
			if !ok {
				t.Fatalf("got %v, want key %s", got, tt.wantDst)
			}
			if string(f.Content) != "hello" {
				t.Errorf("content: got %q, want %q", f.Content, "hello")
			}
			if f.Mode != tt.wantMode {
				t.Errorf("mode: got %o, want %o", f.Mode, tt.wantMode)
			}
			if f.UID != tt.wantUID {
				t.Errorf("uid: got %d, want %d", f.UID, tt.wantUID)
			}
			if f.GID != tt.wantGID {
				t.Errorf("gid: got %d, want %d", f.GID, tt.wantGID)
			}
		})
	}
}

func TestSplitFiles(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  map[string]string
	}{
		{name: "empty", files: nil, want: map[string]string{}},
		{name: "pairs", files: []string{"a:b", "c:d"}, want: map[string]string{"a": "b", "c": "d"}},
		{name: "drops unpaired", files: []string{"a", "c:d"}, want: map[string]string{"c": "d"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitFiles(tt.files)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("key %s: got %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
