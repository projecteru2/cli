package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/net/context"
)

// PublishCommand for publish containers
func PublishCommand() *cli.Command {
	return &cli.Command{
		Name:  "publish",
		Usage: "publish commands",
		Subcommands: []*cli.Command{
			{
				Name:      "dump",
				Usage:     "dump elb things",
				ArgsUsage: "elb url",
				Action:    dumpELB,
			},
			{
				Name:      "update",
				Usage:     "update publish things",
				ArgsUsage: specFileURI,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "app",
						Usage: "app name",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "entry",
						Usage: "entry name",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "node",
						Usage: "which node to run",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "elb",
						Usage: "elb url",
						Value: "",
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "label filter can set multiple times",
					},
					&cli.StringFlag{
						Name:  "upstream-name",
						Usage: "custom upstream name, if not set, will use app_verison_entry",
						Value: "",
					},
				},
				Action: publishContainers,
			},
		},
	}
}

func dumpELB(c *cli.Context) error {
	elb := c.Args().First()
	if elb == "" {
		log.Fatal("need elb url")
	}
	url := fmt.Sprintf("%s/__erulb__/dump", elb)
	return doPut(url, []byte{})
}

func publishContainers(c *cli.Context) error {
	client := setupAndGetGRPCConnection().GetRPCClient()

	app := c.String("app")
	elb := c.String("elb")
	entry := c.String("entry")
	node := c.String("node")
	labels := makeLabels(c.StringSlice("label"))

	if app == "" || elb == "" {
		log.Fatal("[Publish] need appname or elb url")
	}

	specURI := c.Args().First()
	if specURI != "" {
		log.Debugf("[Publish] Publish %s", specURI)

		var data []byte
		var err error
		var domain []byte
		if strings.HasPrefix(specURI, "http") {
			data, err = utils.GetSpecFromRemote(specURI)
		} else {
			data, err = ioutil.ReadFile(specURI)
		}
		if err != nil {
			return cli.Exit(err, -1)
		}
		if domain, err = yaml.YAMLToJSON(data); err != nil {
			log.Fatalf("[Publish] wrong spec file %v", err)
		}

		if err = doUpdateDomain(elb, domain); err != nil {
			log.Fatalf("[Publish] update domain failed %v", err)
		}
	}

	upstreamName := c.String("upstream-name")
	if upstreamName == "" {
		upstreamName = fmt.Sprintf("%s_%s", app, entry)
	}

	lsOpts := &pb.ListContainersOptions{
		Appname:    app,
		Entrypoint: entry,
		Nodename:   node,
		Labels:     labels,
	}

	resp, err := client.ListContainers(context.Background(), lsOpts)
	if err != nil {
		log.Fatalf("[Publish] check container failed %v", err)
	}

	upstreams := map[string]map[string]string{}
	for {
		container, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.Exit(err, -1)
		}
		upstream, ok := upstreams[upstreamName]
		if !ok {
			upstreams[upstreamName] = map[string]string{}
			upstream = upstreams[upstreamName]
		}
		pubinfo := coreutils.DecodePublishInfo(container.Publish)
		for _, pub := range pubinfo {
			for _, addr := range pub {
				upstream[addr] = ""
			}
		}
	}

	return doUpdateUpstream(elb, upstreams)
}

func doUpdateUpstream(elb string, upstreams map[string]map[string]string) error {
	url := fmt.Sprintf("%s/__erulb__/upstream", elb)
	data, err := json.Marshal(upstreams)
	if err != nil {
		log.Fatal(err)
	}
	return doPut(url, data)
}

func doUpdateDomain(elb string, domain []byte) error {
	url := fmt.Sprintf("%s/__erulb__/domain", elb)
	return doPut(url, domain)
}

func doPut(url string, data []byte) error {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode == 200 {
		log.Infof("update %s success %s", url, string(body))
	} else {
		log.Error(string(body))
	}
	return nil
}
