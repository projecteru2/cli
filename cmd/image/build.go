package image

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	dockerengine "github.com/projecteru2/core/engine/docker"
	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/projecteru2/cli/cmd/utils"
)

// progressRewrite rewrites an already printed layer line in place.
const progressRewrite = "\x1b7\x1b[%dA\r\x1b[2K%s\x1b8"

type buildImageOptions struct {
	client corepb.CoreRPCClient
	opts   *corepb.BuildImageOptions
}

func (o *buildImageOptions) run(ctx context.Context) error {
	resp, err := o.client.BuildImage(ctx, o.opts)
	if err != nil {
		return err
	}

	progress := map[string]int{}
	p := 0
	for {
		msg, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Printf("Build failed: %+v\n", err)
			return err
		}

		if msg.Error != "" { //nolint
			fmt.Printf("Build failed: %s\tErrorDetail: %s\n", msg.Error, msg.ErrorDetail)
			return cli.Exit(msg.ErrorDetail.Message, int(msg.ErrorDetail.Code))
		} else if msg.Stream != "" {
			fmt.Print(msg.Stream)
			if msg.Status == "finished" {
				progress = map[string]int{}
				p = 0
			}
		} else if msg.Status != "" {
			if msg.Id == "" {
				fmt.Println(msg.Status)
			} else {
				data := fmt.Sprintf("%s: %s %s", msg.Id, msg.Status, msg.Progress)
				if pos, ok := progress[msg.Id]; !ok {
					progress[msg.Id] = p
					fmt.Println(data)
					p++
				} else if term.IsTerminal(int(os.Stdout.Fd())) {
					fmt.Printf(progressRewrite, p-pos, data)
				} else {
					fmt.Print(data)
				}
			}
		}
	}
	fmt.Printf("build image %s complete\n", o.opts.Name)
	return nil
}

func cmdImageBuild(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	opts, err := generateBuildOptions(ctx, cmd)
	if err != nil {
		return err
	}

	o := &buildImageOptions{
		client: client,
		opts:   opts,
	}
	return o.run(ctx)
}

func generateBuildOptions(ctx context.Context, cmd *cli.Command) (*corepb.BuildImageOptions, error) {
	if cmd.NArg() != 1 {
		return nil, errors.New("no spec given")
	}

	raw := cmd.Bool("raw")
	exist := cmd.Bool("exist")
	if raw && exist {
		return nil, errors.New("--raw and --exist are mutually exclusive")
	}
	stopSignal := cmd.String("stop-signal")

	var (
		specs       *corepb.Builds
		tar         []byte
		existID     string
		buildMethod corepb.BuildImageOptions_BuildMethod
	)
	switch {
	case exist:
		buildMethod = corepb.BuildImageOptions_EXIST
		existID = cmd.Args().First()
	case !raw:
		buildMethod = corepb.BuildImageOptions_SCM
		specURI := cmd.Args().First()
		log.WithFunc("image.generateBuildOptions").Debugf(ctx, "deploy %s", specURI)

		var (
			data []byte
			err  error
		)
		if strings.HasPrefix(specURI, "http") {
			data, err = utils.GetSpecFromRemote(ctx, specURI)
		} else {
			data, err = os.ReadFile(specURI) //nolint:gosec
		}
		if err != nil {
			return nil, fmt.Errorf("read spec: %w", err)
		}
		data, err = utils.EnvParser(data)
		if err != nil {
			return nil, fmt.Errorf("parse env: %w", err)
		}
		specs = &corepb.Builds{}
		if err = yaml.Unmarshal(data, specs); err != nil {
			return nil, fmt.Errorf("unmarshal specs: %w", err)
		}
		for s := range specs.Builds {
			b := specs.Builds[s]
			b.StopSignal = stopSignal
		}
	default:
		buildMethod = corepb.BuildImageOptions_RAW
		path := cmd.Args().First()
		data, err := dockerengine.CreateTarStream(path)
		if err != nil {
			return nil, errors.New("build path must be given")
		}
		tar, err = io.ReadAll(data)
		if err != nil {
			return nil, errors.New("create tar stream failed")
		}
	}

	name := cmd.String("name")
	if name == "" {
		return nil, errors.New("image name must be given")
	}
	user := cmd.String("user")
	uid := int32(cmd.Int("uid")) //nolint:gosec
	tags := cmd.StringSlice("tag")
	if len(tags) == 0 {
		tags = append(tags, "latest")
	}

	return &corepb.BuildImageOptions{
		Name:        name,
		User:        user,
		Uid:         uid,
		Tags:        tags,
		BuildMethod: buildMethod,
		Builds:      specs,
		Tar:         tar,
		ExistId:     existID,
		Platform:    cmd.String("platform"),
	}, nil
}
