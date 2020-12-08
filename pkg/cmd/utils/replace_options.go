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

func GenerateReplaceOptions(c *cli.Context) (*corepb.DeployOptions, error) {
	specURI := c.Args().First()
	if specURI == "" {
		return nil, fmt.Errorf("a specs must be given")
	}
	logrus.Debugf("[Replace] Replace with %s", specURI)

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
			CpuQuotaRequest: 0,
			CpuQuotaLimit:   0,
			CpuBind:         false,
			MemoryRequest:   0,
			MemoryLimit:     0,
			StorageRequest:  0,
			StorageLimit:    0,
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
		Nodelabels:     nil,
		DeployStrategy: corepb.DeployOptions_Strategy(corepb.DeployOptions_Strategy_value[""]),
		Data:           ReadAllFiles(c.StringSlice("file")),
		User:           c.String("user"),
		Debug:          c.Bool("debug"),
		NodesLimit:     0,
		IgnoreHook:     c.Bool("ignore-hook"),
		AfterCreate:    c.StringSlice("after-create"),
		RawArgs:        []byte{},
	}, nil
}
