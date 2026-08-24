package workload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/types"
)

type replaceWorkloadsOptions struct {
	client         corepb.CoreRPCClient
	opts           *corepb.DeployOptions
	labels         map[string]string
	copys          map[string]string
	networkInherit bool
}

func (o *replaceWorkloadsOptions) run(ctx context.Context) error {
	return doReplaceWorkload(ctx, o.client, o.opts, o.networkInherit, o.labels, o.copys)
}

func cmdWorkloadReplace(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	for _, key := range []string{"entry", "image"} {
		if cmd.String(key) == "" {
			return fmt.Errorf("no %s given", key)
		}
	}
	if strings.Contains(cmd.String("entry"), "_") {
		return errors.New("entry can not contain _")
	}

	opts, err := generateReplaceOptions(ctx, cmd)
	if err != nil {
		return err
	}

	networkInherit := cmd.Bool("network-inherit")
	if len(opts.Networks) > 0 {
		log.WithFunc("workload.cmdWorkloadReplace").Warn(ctx, "network is not empty, so network-inherit is set to false")
		networkInherit = false
	}
	o := &replaceWorkloadsOptions{
		client:         client,
		opts:           opts,
		copys:          utils.SplitFiles(cmd.StringSlice("copy")),
		labels:         utils.SplitEquality(cmd.StringSlice("label")),
		networkInherit: networkInherit,
	}
	return o.run(ctx)
}

func doReplaceWorkload(ctx context.Context, client corepb.CoreRPCClient, deployOpts *corepb.DeployOptions, networkInherit bool, labels, copys map[string]string) error {
	logger := log.WithFunc("workload.doReplaceWorkload")
	opts := &corepb.ReplaceOptions{
		DeployOpt:      deployOpts,
		Networkinherit: networkInherit,
		FilterLabels:   labels,
		Copy:           copys,
	}
	resp, err := client.ReplaceWorkload(ctx, opts)
	if err != nil {
		return err
	}
	for {
		msg, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		logger.Infof(ctx, "replace %s", msg.Remove.Id)
		if msg.Error != "" {
			logger.Errorf(ctx, errors.New(msg.Error), "replace %s failed, hook %s", msg.Remove.Id, msg.Remove.Hook)
			if msg.Create != nil && msg.Create.Success {
				logger.Infof(ctx, "but create done id %s name %s", msg.Create.Id, msg.Create.Name)
			}
			continue
		} else if msg.Remove.Hook != "" {
			logger.Infof(ctx, "hook output \n%s", msg.Remove.Hook)
		}

		// a remove message is always returned and always succeeds
		removeMsg := msg.Remove
		logger.Infof(ctx, "workload %s removed", removeMsg.Id)

		// create has succeeded here, otherwise the error surfaces in msg.Error
		createMsg := msg.Create
		logger.Infof(ctx, "new workload %s, resource: %s", createMsg.Name, createMsg.Resources)
		if len(createMsg.Hook) > 0 {
			logger.Infof(ctx, "other output \n%s", createMsg.Hook)
		}
		for name, publish := range createMsg.Publish {
			logger.Infof(ctx, "bound %s ip %s", name, publish)
		}
	}
	return nil
}

func generateReplaceOptions(ctx context.Context, cmd *cli.Command) (*corepb.DeployOptions, error) {
	specURI := cmd.Args().First()
	if specURI == "" {
		return nil, errors.New("a spec must be given")
	}
	log.WithFunc("workload.generateReplaceOptions").Debugf(ctx, "replace with %s", specURI)

	var (
		data []byte
		err  error
	)
	if strings.HasPrefix(specURI, "http") {
		data, err = utils.GetSpecFromRemote(ctx, specURI)
	} else {
		data, err = os.ReadFile(specURI)
	}
	if err != nil {
		return nil, err
	}

	specs := &types.Specs{}
	if err := yaml.Unmarshal(data, specs); err != nil {
		return nil, fmt.Errorf("parse specs: %w", err)
	}

	entry := cmd.String("entry")

	network := cmd.String("network")
	networks := utils.GetNetworks(network)
	entrypoint, ok := specs.Entrypoints[entry]
	if !ok {
		return nil, fmt.Errorf("entry %s not found in specs", entry)
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

	content, modes, owners := utils.GenerateFileOptions(cmd)

	return &corepb.DeployOptions{
		Name: specs.Appname,
		Entrypoint: &corepb.EntrypointOptions{
			Name:        entry,
			Commands:    entrypoint.GetCommands(),
			Privileged:  entrypoint.Privileged,
			Dir:         entrypoint.Dir,
			Log:         logConfig,
			Publish:     entrypoint.Publish,
			Healthcheck: healthCheck,
			Hook:        hook,
			Restart:     entrypoint.Restart,
			Sysctls:     entrypoint.Sysctls,
		},
		Resources: nil,
		Podname:   cmd.String("pod"),
		NodeFilter: &corepb.NodeFilter{
			Includes: cmd.StringSlice("node"),
			Labels:   nil,
		},
		Image:          cmd.String("image"),
		Count:          int32(cmd.Int("count")),
		Env:            cmd.StringSlice("env"),
		Networks:       networks,
		Labels:         specs.Labels,
		Dns:            specs.DNS,
		ExtraHosts:     specs.ExtraHosts,
		DeployStrategy: corepb.DeployOptions_Strategy(corepb.DeployOptions_Strategy_value[""]),
		Data:           content,
		Modes:          modes,
		Owners:         owners,
		User:           cmd.String("user"),
		Debug:          cmd.Bool("debug"),
		NodesLimit:     0,
		IgnoreHook:     cmd.Bool("ignore-hook"),
		AfterCreate:    cmd.StringSlice("after-create"),
		RawArgs:        []byte{},
	}, nil
}
