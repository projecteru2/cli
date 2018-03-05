package commands

import (
	"io"
	"io/ioutil"
	"strings"

	log "github.com/Sirupsen/logrus"
	enginecontainer "github.com/docker/docker/api/types/container"
	"github.com/projecteru2/cli/types"
	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
	"gopkg.in/yaml.v2"
)

func deployContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	specURI := c.Args().First()
	log.Debugf("[Deploy] Deploy %s", specURI)

	pod, node, entry, image, network, cpu, mem, envs, count, nodeLabels, resource := getDeployParams(c)
	var data []byte
	if strings.HasPrefix(specURI, "http") {
		data, err = utils.GetSpecFromRemote(specURI)
	} else {
		data, err = ioutil.ReadFile(specURI)
	}
	if err != nil {
		return cli.Exit(err, -1)
	}
	opts := generateDeployOpts(data, pod, node, entry, image, network, cpu, mem, envs, count, nodeLabels, resource)
	resp, err := client.CreateContainer(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Success {
			log.Infof("[Deploy] Success %s %s %s %v %v %d", msg.Id, msg.Name, msg.Nodename, msg.Cpu, msg.Quota, msg.Memory)
			if len(msg.Hook) > 0 {
				log.Infof("[Deploy] Hook output \n%s", msg.Hook)
			}
			for name, publish := range msg.Publish {
				log.Infof("[Deploy] Bound %s ip %s", name, publish)
			}
		} else {
			log.Errorf("[Deploy] Failed %v", msg.Error)
		}
	}
	return nil
}

func getDeployParams(c *cli.Context) (string, string, string, string, string, float64, int64, []string, int32, map[string]string, bool) {
	pod := c.String("pod")
	node := c.String("node")
	entry := c.String("entry")
	image := c.String("image")
	network := c.String("network")
	cpu := c.Float64("cpu")
	mem := c.Int64("mem")
	envs := c.StringSlice("env")
	count := int32(c.Int("count"))
	resource := c.Bool("with-resource")
	if pod == "" || entry == "" || image == "" {
		log.Fatal("[Deploy] no pod or entry or image")
	}
	labels := map[string]string{}
	for _, d := range c.StringSlice("nodelabel") {
		parts := strings.Split(d, "=")
		labels[parts[0]] = parts[1]
	}
	return pod, node, entry, image, network, cpu, mem, envs, count, labels, resource
}

func generateDeployOpts(data []byte, pod, node, entry, image, network string, cpu float64, mem int64, envs []string, count int32, nodeLabels map[string]string, resource bool) *pb.DeployOptions {
	specs := &types.Specs{}
	if err := yaml.Unmarshal(data, specs); err != nil {
		log.Fatalf("[generateOpts] get specs failed %v", err)
	}

	networkmode := enginecontainer.NetworkMode(network)
	networks := map[string]string{network: ""}
	if !networkmode.IsUserDefined() {
		networks = map[string]string{}
	}
	entrypoint, ok := specs.Entrypoints[entry]
	if !ok {
		log.Fatal("[generateOpts] get entry failed")
	}

	hook := &pb.HookOptions{}
	if entrypoint.Hook != nil {
		hook.AfterStart = entrypoint.Hook.AfterStart
		hook.BeforeStop = entrypoint.Hook.BeforeStop
		hook.Force = entrypoint.Hook.Force
	}

	healthCheck := &pb.HealthCheckOptions{}
	if entrypoint.HealthCheck != nil {
		healthCheck.TcpPorts = entrypoint.HealthCheck.TCPPorts
		healthCheck.HttpPort = entrypoint.HealthCheck.HTTPPort
		healthCheck.Url = entrypoint.HealthCheck.HTTPURL
		healthCheck.Code = int32(entrypoint.HealthCheck.HTTPCode)
	}

	opts := &pb.DeployOptions{
		Name: specs.Appname,
		Entrypoint: &pb.EntrypointOptions{
			Name:          entry,
			Command:       entrypoint.Command,
			Privileged:    entrypoint.Privileged,
			Dir:           entrypoint.Dir,
			LogConfig:     entrypoint.LogConfig,
			Publish:       entrypoint.Publish,
			Healthcheck:   healthCheck,
			Hook:          hook,
			RestartPolicy: entrypoint.RestartPolicy,
		},
		Podname:     pod,
		Nodename:    node,
		Image:       image,
		CpuQuota:    cpu,
		Memory:      mem,
		Count:       count,
		Env:         envs,
		Networks:    networks,
		Networkmode: network,
		Volumes:     specs.Volumes,
		Meta:        specs.Meta,
		Dns:         specs.DNS,
		ExtraHosts:  specs.ExtraHosts,
		Nodelabels:  nodeLabels,
		RawResource: resource,
	}
	return opts
}
