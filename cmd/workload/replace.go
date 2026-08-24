package workload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type replaceWorkloadsOptions struct {
	client         corepb.CoreRPCClient
	opts           *corepb.DeployOptions
	labels         map[string]string
	copies         map[string]string
	networkInherit bool
}

func (o *replaceWorkloadsOptions) run(ctx context.Context) error {
	return doReplaceWorkload(ctx, o.client, o.opts, o.networkInherit, o.labels, o.copies)
}

func cmdWorkloadReplace(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	for _, key := range []string{flagEntry, flagImage} {
		if cmd.String(key) == "" {
			return fmt.Errorf("no %s given", key)
		}
	}
	if strings.Contains(cmd.String(flagEntry), "_") {
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
		copies:         utils.SplitFiles(cmd.StringSlice("copy")),
		labels:         utils.SplitEquality(cmd.StringSlice("label")),
		networkInherit: networkInherit,
	}
	return o.run(ctx)
}

func doReplaceWorkload(ctx context.Context, client corepb.CoreRPCClient, deployOpts *corepb.DeployOptions, networkInherit bool, labels, copies map[string]string) error {
	logger := log.WithFunc("workload.doReplaceWorkload")
	opts := &corepb.ReplaceOptions{
		DeployOpt:      deployOpts,
		Networkinherit: networkInherit,
		FilterLabels:   labels,
		Copy:           copies,
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

		logger.Infof(ctx, "workload %s removed", msg.Remove.Id)
		logger.Infof(ctx, "new workload %s, resource: %s", msg.Create.Name, msg.Create.Resources)
		if len(msg.Create.Hook) > 0 {
			logger.Infof(ctx, "other output \n%s", msg.Create.Hook)
		}
		for name, publish := range msg.Create.Publish {
			logger.Infof(ctx, "bound %s ip %s", name, publish)
		}
	}
	return nil
}

func generateReplaceOptions(ctx context.Context, cmd *cli.Command) (*corepb.DeployOptions, error) {
	specs, err := loadSpecs(ctx, cmd)
	if err != nil {
		return nil, err
	}

	entrypoint, err := entrypointOptions(specs, cmd.String(flagEntry))
	if err != nil {
		return nil, err
	}

	content, modes, owners := utils.GenerateFileOptions(cmd)

	return &corepb.DeployOptions{
		Name:       specs.Appname,
		Entrypoint: entrypoint,
		Podname:    cmd.String(flagPod),
		NodeFilter: &corepb.NodeFilter{
			Includes: cmd.StringSlice(flagNode),
		},
		Image:          cmd.String(flagImage),
		Count:          int32(cmd.Int("count")), //nolint:gosec
		Env:            cmd.StringSlice(flagEnv),
		Networks:       utils.GetNetworks(cmd.String(flagNetwork)),
		Labels:         specs.Labels,
		Dns:            specs.DNS,
		ExtraHosts:     specs.ExtraHosts,
		DeployStrategy: corepb.DeployOptions_AUTO,
		Data:           content,
		Modes:          modes,
		Owners:         owners,
		User:           cmd.String("user"),
		Debug:          cmd.Bool("debug"),
		IgnoreHook:     cmd.Bool("ignore-hook"),
		AfterCreate:    cmd.StringSlice("after-create"),
	}, nil
}
