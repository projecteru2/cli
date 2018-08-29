package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	enginetypes "github.com/docker/docker/api/types"
	"github.com/ghodss/yaml"
	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
)

// PublishCommand for publish containers
func PublishCommand() *cli.Command {
	return &cli.Command{
		Name:  "publish",
		Usage: "publish commands",
		Subcommands: []*cli.Command{
			&cli.Command{
				Name:      "dump",
				Usage:     "dump elb things",
				ArgsUsage: "elb url",
				Action:    dumpELB,
			},
			&cli.Command{
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
						Name:  "elb",
						Usage: "elb url",
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
						Name:  "version",
						Usage: "verison",
						Value: "latest",
					},
					&cli.StringFlag{
						Name:  "upstream_name",
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
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	specURI := c.Args().First()
	log.Debugf("[Publish] Publish %s", specURI)

	var data []byte
	domain := []byte{}
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

	app := c.String("app")
	elb := c.String("elb")
	entry := c.String("entry")
	node := c.String("node")
	version := c.String("version")
	if app == "" || elb == "" {
		log.Fatal("[Publish] need appname or elb url")
	}
	upstreamName := c.String("upstream_name")
	if upstreamName == "" {
		upstreamName = fmt.Sprintf("%s_%s_%s", app, version, entry)
	}

	lsOpts := &pb.DeployStatusOptions{
		Appname:    app,
		Entrypoint: entry,
		Nodename:   node,
	}

	resp, err := client.ListContainers(context.Background(), lsOpts)
	if err != nil {
		log.Fatalf("[Publish] check container failed %v", err)
	}

	upstreams := map[string]map[string]string{}
	for _, container := range resp.Containers {
		containerJSON := &enginetypes.ContainerJSON{}
		if err := json.Unmarshal(container.Inspect, containerJSON); err != nil {
			log.Error(err)
			continue
		}

		labels := containerJSON.Config.Labels
		// 筛选
		if v, ok := labels["version"]; version != "" && (!ok || v != version) {
			log.Warnf("[Publish] version not fit %s", v)
			continue
		}

		if err != nil {
			log.Errorf("[Publish] parse container name failed %v", err)
			continue
		}

		upstream, ok := upstreams[upstreamName]
		if !ok {
			upstreams[upstreamName] = map[string]string{}
			upstream = upstreams[upstreamName]
		}

		for _, network := range containerJSON.NetworkSettings.Networks {
			if network.IPAddress == "" {
				log.Warnf("[Publish] can not get ip")
				continue
			}
			portstr, ok := labels["publish"]
			if !ok {
				continue
			}
			ports := strings.Split(portstr, ",")
			for _, port := range ports {
				publish := fmt.Sprintf("%s:%s", network.IPAddress, port)
				upstream[publish] = ""
			}
		}
	}

	if err := doUpdateUpstream(elb, upstreams); err != nil {
		return err
	}
	return doUpdateDomain(elb, domain)
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
