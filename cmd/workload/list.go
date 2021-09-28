package workload

import (
	"context"
	"io"
	"strings"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type listWorkloadsOptions struct {
	client corepb.CoreRPCClient
	// must be set
	appname string
	limit   int64
	// filters
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
	for {
		w, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		workloads = append(workloads, w)
	}

	f := filter{
		ips:       o.matchIPs,
		skipIPs:   o.skipIPs,
		nodenames: []string{},
		podnames:  []string{},
	}
	if len(o.nodename) > 0 {
		f.nodenames = append(f.nodenames, o.nodename)
	}
	if len(o.podnames) > 0 {
		f.podnames = append(f.podnames, o.podnames...)
	}

	workloads = f.filterIn(workloads)

	if o.statistics {
		describe.WorkloadsStatistics(workloads...)
	} else {
		describe.Workloads(workloads...)
	}

	return nil
}

type filter struct {
	ips       []string
	skipIPs   []string
	nodenames []string
	podnames  []string
}

func (wf filter) filterIn(workloads []*corepb.Workload) []*corepb.Workload {
	ans := []*corepb.Workload{}
	for _, workload := range workloads {
		if !wf.skip(workload) {
			ans = append(ans, workload)
		}
	}
	return ans
}

func (wf filter) skip(workload *corepb.Workload) bool {
	if workload == nil {
		return true
	}
	if len(wf.nodenames) > 0 && !wf.hasIntersection(wf.nodenames, []string{workload.Nodename}) {
		return true
	}

	// Don't skip any workload if there isn't Status.
	if workload.Status == nil {
		return false
	}

	ips := []string{}
	for _, cidr := range workload.Status.Networks {
		ips = append(ips, strings.Split(cidr, "/")[0])
	}

	return (len(wf.ips) > 0 && !wf.hasIntersection(wf.ips, ips)) ||
		(len(wf.skipIPs) > 0 && wf.hasIntersection(wf.skipIPs, ips)) ||
		(len(wf.podnames) > 0 && !wf.hasIntersection(wf.podnames, []string{workload.Podname}))
}

func (wf filter) hasIntersection(a, b []string) bool {
	hash := map[string]bool{}
	for _, v := range a {
		hash[v] = true
	}

	for _, v := range b {
		if _, exists := hash[v]; exists {
			return true
		}
	}

	return false
}

func cmdWorkloadList(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	o := &listWorkloadsOptions{
		client:     client,
		appname:    c.Args().First(),
		entrypoint: c.String("entry"),
		nodename:   c.String("nodename"),
		labels:     utils.SplitEquality(c.StringSlice("label")),
		limit:      c.Int64("limit"),
		matchIPs:   c.StringSlice("match-ip"),
		skipIPs:    c.StringSlice("skip-ip"),
		podnames:   c.StringSlice("pod"),
		statistics: c.Bool("statistics"),
	}
	return o.run(c.Context)
}
