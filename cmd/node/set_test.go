package node

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
)

func TestGenerateSetNodeOptionsKeepsTLSByDefault(t *testing.T) {
	opts := runSetNodeCommand(t, []string{"node", "set", "--cpu", "64", "node1"})
	if opts.UpdateTls || opts.Ca != "" || opts.Cert != "" || opts.Key != "" {
		t.Errorf("got update_tls=%v ca=%q cert=%q key=%q", opts.UpdateTls, opts.Ca, opts.Cert, opts.Key)
	}
}

func TestGenerateSetNodeOptionsUpdatesExplicitTLS(t *testing.T) {
	dir := t.TempDir()
	ca := filepath.Join(dir, "ca")
	cert := filepath.Join(dir, "cert")
	key := filepath.Join(dir, "key")
	for path, content := range map[string]string{ca: "ca-data", cert: "cert-data", key: "key-data"} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write tls file: %v", err)
		}
	}

	opts := runSetNodeCommand(t, []string{"node", "set", "--ca", ca, "--cert", cert, "--key", key, "node1"})
	if !opts.UpdateTls || opts.Ca != "ca-data" || opts.Cert != "cert-data" || opts.Key != "key-data" {
		t.Errorf("got update_tls=%v ca=%q cert=%q key=%q", opts.UpdateTls, opts.Ca, opts.Cert, opts.Key)
	}
}

func runSetNodeCommand(t *testing.T, args []string) *corepb.SetNodeOptions {
	t.Helper()
	var opts *corepb.SetNodeOptions
	c := Command()
	lookupSubcommand(t, c, "set").Action = func(_ context.Context, cmd *cli.Command) error {
		var err error
		opts, err = generateSetNodeOptions(cmd)
		return err
	}
	if err := c.Run(t.Context(), args); err != nil {
		t.Fatalf("run: %v", err)
	}
	if opts == nil {
		t.Fatal("got nil options")
	}
	return opts
}
