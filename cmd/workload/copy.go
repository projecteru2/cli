package workload

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type copyWorkloadsOptions struct {
	client corepb.CoreRPCClient
	// where to store copied content
	dir string
	// map workloadID -> list of path of files
	sources map[string][]string
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
	baseDir := filepath.Join(o.dir)
	if err := os.MkdirAll(baseDir, os.FileMode(0700)); err != nil {
		return err
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if msg.Error != "" {
			logrus.Errorf("[Copy] Failed %s %s", coreutils.ShortID(msg.Id), msg.Error)
			continue
		}

		filename := fmt.Sprintf("%s-%s-%s.tar.gz", coreutils.ShortID(msg.Id), msg.Name, now)
		storePath := filepath.Join(baseDir, filename)
		if _, err := os.Stat(storePath); err != nil {
			f, err := os.Create(storePath)
			if err != nil {
				logrus.Errorf("[Copy] Error during create backup file %s: %v", storePath, err)
				continue
			}
			if _, err := f.Write(msg.Data); err != nil {
				logrus.Errorf("[Copy] Write file error %v", err)
			}
			f.Close()
		}
	}
	return nil
}

func cmdWorkloadCopy(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	sources := map[string][]string{}
	for _, source := range c.Args().Slice() {
		ps := strings.Split(source, ":")
		if len(ps) != 2 {
			continue
		}

		fs := strings.Split(ps[1], ",")
		if len(fs) == 0 {
			continue
		}

		sources[ps[0]] = fs
	}

	if len(sources) == 0 {
		return fmt.Errorf("source files should not be empty")
	}

	o := &copyWorkloadsOptions{
		client:  client,
		sources: sources,
		dir:     c.String("dir"),
	}
	return o.run(c.Context)
}
