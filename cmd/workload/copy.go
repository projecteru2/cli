package workload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	corepb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type copyWorkloadsOptions struct {
	client  corepb.CoreRPCClient
	dir     string
	sources map[string][]string // workload ID -> paths to copy
}

func (o *copyWorkloadsOptions) run(ctx context.Context) error {
	targets := map[string]*corepb.CopyPaths{}
	for id, paths := range o.sources {
		targets[id] = &corepb.CopyPaths{Paths: paths}
	}

	resp, err := o.client.Copy(ctx, &corepb.CopyOptions{Targets: targets})
	if err != nil {
		return err
	}

	now := time.Now().Format("2006.01.02.15.04.05")
	if err := os.MkdirAll(o.dir, os.FileMode(0o700)); err != nil {
		return err
	}

	var errs error
	files := make(map[string][]byte)

	for {
		msg, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if msg.Error != "" {
			errs = errors.Join(errs, fmt.Errorf("copy from %s: %s", coreutils.ShortID(msg.Id), msg.Error))
			continue
		}

		filename := fmt.Sprintf("%s-%s-%s.tar", coreutils.ShortID(msg.Id), msg.Name, now)
		files[filename] = append(files[filename], msg.Data...)
	}

	for filename, content := range files {
		storePath := filepath.Join(o.dir, filename)
		if _, err := os.Stat(storePath); err == nil {
			errs = errors.Join(errs, fmt.Errorf("%s already exists", storePath))
			continue
		}
		if err := os.WriteFile(storePath, content, 0o600); err != nil {
			errs = errors.Join(errs, fmt.Errorf("write %s: %w", storePath, err))
		}
	}
	return errs
}

func cmdWorkloadCopy(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	sources := map[string][]string{}
	for _, source := range cmd.Args().Slice() {
		ps := strings.Split(source, ":")
		if len(ps) != 2 {
			continue
		}

		sources[ps[0]] = strings.Split(ps[1], ",")
	}

	if len(sources) == 0 {
		return errors.New("source files should not be empty")
	}

	o := &copyWorkloadsOptions{
		client:  client,
		sources: sources,
		dir:     cmd.String("dir"),
	}
	return o.run(ctx)
}
