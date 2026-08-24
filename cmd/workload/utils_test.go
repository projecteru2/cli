package workload

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestLoadSpecsReadsALocalPathNamedLikeAURL(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "http.yaml"), []byte(specSample), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	c := Command()
	lookupSubcommand(t, c, "deploy").Action = func(ctx context.Context, cmd *cli.Command) error {
		specs, err := loadSpecs(ctx, cmd)
		if err != nil {
			t.Errorf("loadSpecs: %v", err)
			return nil
		}
		if specs.Appname != "elb" {
			t.Errorf("appname: got %q, want %q", specs.Appname, "elb")
		}
		return nil
	}
	if err := c.Run(t.Context(), []string{"workload", "deploy", "--entry", "release", "http.yaml"}); err != nil {
		t.Fatalf("run: %v", err)
	}
}
