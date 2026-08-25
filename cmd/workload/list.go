package workload

import (
	"context"
	"slices"
	"strings"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type listWorkloadsOptions struct {
	client corepb.CoreRPCClient

	appname string
	limit   int64

	entrypoint string
	nodename   string
	labels     map[string]string
	matchIPs   []string
	skipIPs    []string
	podnames   []string
	statistics bool
}

func (o *listWorkloadsOptions) run(ctx context.Context) error {
	opts := &corepb.ListWorkloadsOptions{
		Appname:    o.appname,
		Entrypoint: o.entrypoint,
		Nodename:   o.nodename,
		Labels:     o.labels,
		Limit:      o.limit,
	}

	resp, err := o.client.ListWorkloads(ctx, opts)
	if err != nil {
		return err
	}

	workloads := []*corepb.Workload{}
	if err := utils.EachMessage(resp.Recv, func(w *corepb.Workload) error {
		workloads = append(workloads, w)
		return nil
	}); err != nil {
		return err
	}

	f := filter{
		ips:      o.matchIPs,
		skipIPs:  o.skipIPs,
		podnames: o.podnames,
	}
	if o.nodename != "" {
		f.nodenames = []string{o.nodename}
	}

	workloads = f.filterIn(workloads)

	if o.statistics {
		describe.WorkloadsStatistics(workloads...)
	} else {
		describe.Workloads(workloads...)
	}

	return nil
}

func cmdWorkloadList(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	o := &listWorkloadsOptions{
		client:     client,
		appname:    cmd.Args().First(),
		entrypoint: cmd.String(flagEntry),
		nodename:   cmd.String(flagNode),
		labels:     utils.SplitEquality(cmd.StringSlice("label")),
		limit:      cmd.Int64("limit"),
		matchIPs:   cmd.StringSlice("match-ip"),
		skipIPs:    cmd.StringSlice("skip-ip"),
		podnames:   cmd.StringSlice(flagPod),
		statistics: cmd.Bool("statistics"),
	}
	return o.run(ctx)
}

type filter struct {
	ips       []string
	skipIPs   []string
	nodenames []string
	podnames  []string
}

func (f filter) filterIn(workloads []*corepb.Workload) []*corepb.Workload {
	return slices.DeleteFunc(workloads, f.skip)
}

func (f filter) skip(workload *corepb.Workload) bool {
	if workload == nil {
		return true
	}
	if len(f.nodenames) > 0 && !slices.Contains(f.nodenames, workload.Nodename) {
		return true
	}
	if len(f.podnames) > 0 && !slices.Contains(f.podnames, workload.Podname) {
		return true
	}

	if workload.Status == nil {
		return false
	}

	ips := []string{}
	for _, cidr := range workload.Status.Networks {
		ips = append(ips, strings.Split(cidr, "/")[0])
	}

	return (len(f.ips) > 0 && !hasIntersection(f.ips, ips)) ||
		(len(f.skipIPs) > 0 && hasIntersection(f.skipIPs, ips))
}

func hasIntersection(a, b []string) bool {
	return slices.ContainsFunc(b, func(v string) bool { return slices.Contains(a, v) })
}
