package utils

import (
	"fmt"
	"io/ioutil"
	"strings"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func GenerateDeployOptions(c *cli.Context) (*corepb.DeployOptions, error) {
	specURI := c.Args().First()
	if specURI == "" {
		return nil, fmt.Errorf("a specs must be given")
	}
	logrus.Debugf("[Deploy] Deploy %s", specURI)

	var (
		data []byte
		err  error
	)
	if strings.HasPrefix(specURI, "http") {
		data, err = GetSpecFromRemote(specURI)
	} else {
		data, err = ioutil.ReadFile(specURI)
	}
	if err != nil {
		return nil, err
	}

	memLimit, err := ParseRAMInHuman(c.String("memory-limit"))
	if err != nil {
		return nil, fmt.Errorf("[getDeployParams] parse memory failed %v", err)
	}
	memRequest, err := ParseRAMInHuman(c.String("memory-request"))
	if err != nil {
		return nil, fmt.Errorf("[getDeployParams] parse memory failed %v", err)
	}
	storageLimit, err := ParseRAMInHuman(c.String("storage-limit"))
	if err != nil {
		return nil, fmt.Errorf("[getDeployParams] parse storage failed %v", err)
	}
	storageRequest, err := ParseRAMInHuman(c.String("storage-request"))
	if err != nil {
		return nil, fmt.Errorf("[getDeployParams] parse storage failed %v", err)
	}
	specs := Specs{}
	if err := yaml.Unmarshal(data, specs); err != nil {
		return nil, fmt.Errorf("[generateOpts] get specs failed %v", err)
	}

	entry := c.String("entry")

	network := c.String("network")
	networks := GetNetworks(network)
	entrypoint, ok := specs.Entrypoints[entry]
	if !ok {
		return nil, fmt.Errorf("[generateOpts] get entry failed")
	}

	var hook *corepb.HookOptions
	if entrypoint.Hook != nil {
		hook = &corepb.HookOptions{
			AfterStart: entrypoint.Hook.AfterStart,
			BeforeStop: entrypoint.Hook.BeforeStop,
			Force:      entrypoint.Hook.Force,
		}
	}

	var healthCheck *corepb.HealthCheckOptions
	if entrypoint.HealthCheck != nil {
		healthCheck = &corepb.HealthCheckOptions{
			TcpPorts: entrypoint.HealthCheck.TCPPorts,
			HttpPort: entrypoint.HealthCheck.HTTPPort,
			Url:      entrypoint.HealthCheck.HTTPURL,
			Code:     int32(entrypoint.HealthCheck.HTTPCode),
		}
	}

	var logConfig *corepb.LogOptions
	if entrypoint.Log != nil {
		logConfig = &corepb.LogOptions{
			Type:   entrypoint.Log.Type,
			Config: entrypoint.Log.Config,
		}
	}

	rawArgs := c.String("raw-args")
	rawArgsByte := []byte{}
	if rawArgs != "" {
		rawArgsByte = []byte(rawArgs)
	}

	return &corepb.DeployOptions{
		Name: specs.Appname,
		Entrypoint: &corepb.EntrypointOptions{
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
		ResourceOpts: &corepb.ResourceOptions{
			CpuQuotaRequest: c.Float64("cpu-request"),
			CpuQuotaLimit:   c.Float64("cpu-limit"),
			CpuBind:         c.Bool("cpu-bind"),
			MemoryRequest:   memRequest,
			MemoryLimit:     memLimit,
			StorageRequest:  storageRequest,
			StorageLimit:    storageLimit,
			VolumesRequest:  specs.VolumesRequest,
			VolumesLimit:    specs.Volumes,
		},
		Podname:        c.String("pod"),
		Nodenames:      c.StringSlice("node"),
		Image:          c.String("image"),
		Count:          int32(c.Int("count")),
		Env:            c.StringSlice("env"),
		Networks:       networks,
		Networkmode:    network,
		Labels:         specs.Labels,
		Dns:            specs.DNS,
		ExtraHosts:     specs.ExtraHosts,
		Nodelabels:     SplitEquality(c.StringSlice("nodelabel")),
		DeployStrategy: corepb.DeployOptions_Strategy(corepb.DeployOptions_Strategy_value[strings.ToUpper(c.String("deploy-strategy"))]),
		Data:           ReadAllFiles(c.StringSlice("file")),
		User:           c.String("user"),
		Debug:          c.Bool("debug"),
		NodesLimit:     int32(c.Int("nodes-limit")),
		IgnoreHook:     c.Bool("ignore-hook"),
		AfterCreate:    c.StringSlice("after-create"),
		RawArgs:        rawArgsByte,
	}, nil
}
