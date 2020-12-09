package image

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	dockerengine "github.com/projecteru2/core/engine/docker"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sethgrid/curse"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

type buildImageOptions struct {
	client corepb.CoreRPCClient
	opts   *corepb.BuildImageOptions
}

func (o *buildImageOptions) run(ctx context.Context) error {
	resp, err := o.client.BuildImage(ctx, o.opts)
	if err != nil {
		return err
	}

	progess := map[string]int{}
	p := 0
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if msg.Error != "" { // nolint
			return cli.Exit(msg.ErrorDetail.Message, int(msg.ErrorDetail.Code))
		} else if msg.Stream != "" {
			fmt.Print(msg.Stream)
			if msg.Status == "finished" {
				progess = map[string]int{}
				p = 0
			}
		} else if msg.Status != "" {
			if msg.Id == "" {
				fmt.Println(msg.Status)
			} else {
				data := fmt.Sprintf("%s: %s %s", msg.Id, msg.Status, msg.Progress)
				if pos, ok := progess[msg.Id]; !ok {
					progess[msg.Id] = p
					fmt.Println(data)
					p++
				} else {
					cursor, err := curse.New()
					if err != nil {
						fmt.Print(data)
						continue
					}
					cursor.MoveUp(p - pos).EraseCurrentLine()
					fmt.Print(data)
					cursor.Reset()
				}
			}
		}
	}
	return nil
}

func cmdImageBuild(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	opts, err := generateBuildOptions(c)
	if err != nil {
		return err
	}

	o := &buildImageOptions{
		client: client,
		opts:   opts,
	}
	return o.run(c.Context)
}

func generateBuildOptions(c *cli.Context) (*corepb.BuildImageOptions, error) {
	if c.NArg() != 1 {
		return nil, errors.New("[Build] no spec")
	}

	raw := c.Bool("raw")
	exist := c.Bool("exist")
	if raw && exist {
		return nil, errors.New("[Build] mutually exclusive flag: raw or exist")
	}
	stopSignal := c.String("stop-signal")

	var (
		specs       *corepb.Builds
		tar         []byte
		existID     string
		buildMethod corepb.BuildImageOptions_BuildMethod
	)
	switch {
	case exist:
		buildMethod = corepb.BuildImageOptions_EXIST
		existID = c.Args().First()
	case !raw:
		buildMethod = corepb.BuildImageOptions_SCM
		specURI := c.Args().First()
		logrus.Debugf("[Build] Deploy %s", specURI)

		var (
			data []byte
			err  error
		)
		if strings.HasPrefix(specURI, "http") {
			data, err = utils.GetSpecFromRemote(specURI)
		} else {
			data, err = ioutil.ReadFile(specURI)
		}
		if err != nil {
			return nil, fmt.Errorf("[Build] read spec failed %v", err)
		}
		data, err = utils.EnvParser(data)
		if err != nil {
			return nil, fmt.Errorf("[Build] parse env failed %v", err)
		}
		specs = &corepb.Builds{}
		if err = yaml.Unmarshal(data, specs); err != nil {
			return nil, fmt.Errorf("[Build] unmarshal specs failed %v", err)
		}
		for s := range specs.Builds {
			b := specs.Builds[s]
			b.StopSignal = stopSignal
		}
	default:
		buildMethod = corepb.BuildImageOptions_RAW
		path := c.Args().First()
		data, err := dockerengine.CreateTarStream(path)
		if err != nil {
			return nil, errors.New("[Build] no path")
		}
		tar, err = ioutil.ReadAll(data)
		if err != nil {
			return nil, errors.New("[Build] create tar stream failed")
		}
	}

	name := c.String("name")
	if name == "" {
		return nil, errors.New("[Build] need name")
	}
	user := c.String("user")
	uid := int32(c.Int("uid"))
	tags := c.StringSlice("tag")
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
	}, nil
}
