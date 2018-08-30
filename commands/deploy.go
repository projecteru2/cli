package commands

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/projecteru2/cli/types"
	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
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

	autoReplace := c.Bool("auto_replace")
	pod, node, entry, image, network, cpu, mem, envs, count, nodeLabels, deployMethod, files, user, debug, softlimit := getDeployParams(c)
	if pod == "" || entry == "" || image == "" {
		log.Fatal("[Deploy] no pod or entry or image")
	}
	if strings.Contains(entry, "_") {
		log.Fatal("[Deploy] entry can not contain _")
	}

	var data []byte
	if strings.HasPrefix(specURI, "http") {
		data, err = utils.GetSpecFromRemote(specURI)
	} else {
		data, err = ioutil.ReadFile(specURI)
	}
	if err != nil {
		return cli.Exit(err, -1)
	}

	deployOpts := generateDeployOpts(data, pod, node, entry, image, network, cpu, mem, envs, count, nodeLabels, deployMethod, files, user, debug, softlimit)
	if autoReplace {
		lsOpts := &pb.DeployStatusOptions{
			Appname:    deployOpts.Name,
			Entrypoint: deployOpts.Entrypoint.Name,
			Nodename:   node,
		}
		resp, err := client.ListContainers(context.Background(), lsOpts)
		if err != nil {
			log.Errorf("[Deploy] check container failed %v", err)
		} else {
			if len(resp.Containers) > 0 {
				return doReplaceContainer(client, deployOpts, true)
			}
		}
	}

	return doCreateContainer(client, deployOpts)
}

func doCreateContainer(client pb.CoreRPCClient, deployOpts *pb.DeployOptions) error {
	resp, err := client.CreateContainer(context.Background(), deployOpts)
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

func getDeployParams(c *cli.Context) (string, string, string, string, string, float64, int64, []string, int32, map[string]string, string, []string, string, bool, bool) {
	pod := c.String("pod")
	node := c.String("node")
	entry := c.String("entry")
	image := c.String("image")
	network := c.String("network")
	cpu := c.Float64("cpu")
	mem := c.Int64("mem")
	envs := c.StringSlice("env")
	files := c.StringSlice("file")
	count := int32(c.Int("count"))
	deployMethod := c.String("deploy-method")
	user := c.String("user")
	debug := c.Bool("debug")
	softlimit := c.Bool("softlimit")
	labels := map[string]string{}
	for _, d := range c.StringSlice("nodelabel") {
		parts := strings.Split(d, "=")
		labels[parts[0]] = parts[1]
	}
	return pod, node, entry, image, network, cpu, mem, envs, count, labels, deployMethod, files, user, debug, softlimit
}

func generateDeployOpts(data []byte, pod, node, entry, image, network string, cpu float64, mem int64, envs []string, count int32, nodeLabels map[string]string, deployMethod string, files []string, user string, debug, softlimit bool) *pb.DeployOptions {
	specs := &types.Specs{}
	if err := yaml.Unmarshal(data, specs); err != nil {
		log.Fatalf("[generateOpts] get specs failed %v", err)
	}

	networks := getNetworks(network)
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

	logConfig := &pb.LogOptions{}
	if entrypoint.Log != nil {
		logConfig.Type = entrypoint.Log.Type
		logConfig.Config = entrypoint.Log.Config
	}

	fileData := utils.GetFilesStream(files)

	opts := &pb.DeployOptions{
		Name: specs.Appname,
		Entrypoint: &pb.EntrypointOptions{
			Name:          entry,
			Command:       entrypoint.Command,
			Privileged:    entrypoint.Privileged,
			Dir:           entrypoint.Dir,
			Log:           logConfig,
			Publish:       entrypoint.Publish,
			Healthcheck:   healthCheck,
			Hook:          hook,
			RestartPolicy: entrypoint.RestartPolicy,
			Sysctls:       entrypoint.Sysctls,
		},
		Podname:      pod,
		Nodename:     node,
		Image:        image,
		CpuQuota:     cpu,
		Memory:       mem,
		Count:        count,
		Env:          envs,
		Networks:     networks,
		Networkmode:  network,
		Volumes:      specs.Volumes,
		Meta:         specs.Meta,
		Dns:          specs.DNS,
		ExtraHosts:   specs.ExtraHosts,
		Nodelabels:   nodeLabels,
		DeployMethod: deployMethod,
		Data:         fileData,
		User:         user,
		Debug:        debug,
		Softlimit:    softlimit,
	}
	return opts
}
