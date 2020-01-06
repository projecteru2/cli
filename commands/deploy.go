package commands

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/projecteru2/cli/types"
	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/net/context"
	yaml "gopkg.in/yaml.v2"
)

func deployContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	specURI := c.Args().First()
	log.Debugf("[Deploy] Deploy %s", specURI)

	autoReplace := c.Bool("auto-replace")
	pod, node, entry, image, network, cpu, mem, storage, envs, count, nodeLabels, deployMethod, files, user, debug, softlimit, nodesLimit, cpubind, ignoreHook, afterCreate, rawArgs := getDeployParams(c)
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

	deployOpts := generateDeployOpts(data, pod, node, entry, image, network, cpu, mem, storage, envs, count, nodeLabels, deployMethod, files, user, debug, softlimit, cpubind, ignoreHook, nodesLimit, afterCreate, rawArgs)
	if autoReplace {
		lsOpts := &pb.ListContainersOptions{
			Appname:    deployOpts.Name,
			Entrypoint: deployOpts.Entrypoint.Name,
			Nodename:   node,
			Labels:     nil,
			Limit:      1, // 至少有一个可以被替换的
		}
		resp, err := client.ListContainers(context.Background(), lsOpts)
		if err != nil {
			log.Warnf("[Deploy] check container failed %v", err)
		} else {
			_, err := resp.Recv()
			if err == io.EOF {
				log.Warn("[Deploy] there is no containers for replace")
				goto DOCREATE
			}
			if err != nil {
				return cli.Exit(err, -1)
			}
			// 强制继承网络
			networkInherit := true
			if network != "" {
				networkInherit = false
			}
			return doReplaceContainer(client, deployOpts, networkInherit, nil, nil)
		}
	}

DOCREATE:
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
			log.Infof("[Deploy] Success %s %s %s %v %v %d %v", msg.Id, msg.Name, msg.Nodename, msg.Cpu, msg.Quota, msg.Memory, msg.VolumePlan)
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

func getDeployParams(c *cli.Context) (string, string, string, string, string, float64, int64, int64, []string, int32, map[string]string, string, []string, string, bool, bool, int, bool, bool, []string, string) {
	pod := c.String("pod")
	node := c.String("node")
	entry := c.String("entry")
	image := c.String("image")
	network := c.String("network")
	cpu := c.Float64("cpu")
	mem, err := parseRAMInHuman(c.String("memory"))
	if err != nil {
		log.Fatalf("[getDeployParams] parse memory failed %v", err)
	}
	storage, err := parseRAMInHuman(c.String("storage"))
	if err != nil {
		log.Fatalf("[getDeployParams] parse storage failed %v", err)
	}
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
	nodesLimit := c.Int("nodes-limit")
	cpubind := c.Bool("cpu-bind")
	ignoreHook := c.Bool("ignore-hook")
	afterCreate := c.StringSlice("after-create")
	rawArgs := c.String("raw-args")
	return pod, node, entry, image, network, cpu, mem, storage, envs, count, labels, deployMethod, files, user, debug, softlimit, nodesLimit, cpubind, ignoreHook, afterCreate, rawArgs
}

func generateDeployOpts(data []byte, pod, node, entry, image, network string, cpu float64, mem, storage int64, envs []string, count int32, nodeLabels map[string]string, deployMethod string, files []string, user string, debug, softlimit, cpubind, ignoreHook bool, nodesLimit int, afterCreate []string, rawArgs string) *pb.DeployOptions {
	specs := &types.Specs{}
	if err := yaml.Unmarshal(data, specs); err != nil {
		log.Fatalf("[generateOpts] get specs failed %v", err)
	}
	networks := getNetworks(network)
	entrypoint, ok := specs.Entrypoints[entry]
	if !ok {
		log.Fatal("[generateOpts] get entry failed")
	}

	var hook *pb.HookOptions
	if entrypoint.Hook != nil {
		hook = &pb.HookOptions{}
		hook.AfterStart = entrypoint.Hook.AfterStart
		hook.BeforeStop = entrypoint.Hook.BeforeStop
		hook.Force = entrypoint.Hook.Force
	}

	var healthCheck *pb.HealthCheckOptions
	if entrypoint.HealthCheck != nil {
		healthCheck = &pb.HealthCheckOptions{}
		healthCheck.TcpPorts = entrypoint.HealthCheck.TCPPorts
		healthCheck.HttpPort = entrypoint.HealthCheck.HTTPPort
		healthCheck.Url = entrypoint.HealthCheck.HTTPURL
		healthCheck.Code = int32(entrypoint.HealthCheck.HTTPCode)
	}

	var logConfig *pb.LogOptions
	if entrypoint.Log != nil {
		logConfig = &pb.LogOptions{}
		logConfig.Type = entrypoint.Log.Type
		logConfig.Config = entrypoint.Log.Config
	}

	fileData := utils.GetFilesStream(files)
	rawArgsByte := []byte{}
	if rawArgs != "" {
		rawArgsByte = []byte(rawArgs)
	}

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
		Storage:      storage,
		Count:        count,
		Env:          envs,
		Networks:     networks,
		Networkmode:  network,
		Volumes:      specs.Volumes,
		Labels:       specs.Labels,
		Dns:          specs.DNS,
		ExtraHosts:   specs.ExtraHosts,
		Nodelabels:   nodeLabels,
		DeployMethod: deployMethod,
		Data:         fileData,
		User:         user,
		Debug:        debug,
		SoftLimit:    softlimit,
		NodesLimit:   int32(nodesLimit),
		CpuBind:      cpubind,
		IgnoreHook:   ignoreHook,
		AfterCreate:  afterCreate,
		RawArgs:      rawArgsByte,
	}
	return opts
}
