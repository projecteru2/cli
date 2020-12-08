package describe

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

func DescribeWorkloads(workloads ...*corepb.Workload) {
	switch strings.ToLower(Format) {
	case "json":
		describeAsJSON(workloads)
	case "yaml", "yml":
		describeAsYAML(workloads)
	default:
		describeWorkloads(workloads)
	}
}

func describeWorkloads(workloads []*corepb.Workload) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name/ID", "Pod", "Node", "Status", "Volume", "IP", "Networks"})

	for _, c := range workloads {
		// publish ip
		ips := []string{}
		for networkName, ip := range c.Publish {
			ips = append(ips, fmt.Sprintf("%s: %s", networkName, ip))
		}

		// networks
		ns := []string{}
		if c.Status != nil {
			for name, ip := range c.Status.Networks {
				ns = append(ns, fmt.Sprintf("%s: %s", name, ip))
			}
		}

		rows := [][]string{
			{c.Name, c.Id},
			{c.Podname},
			{c.Nodename},
			{
				fmt.Sprintf("QuotaRequest: %f", c.Resource.CpuQuotaRequest),
				fmt.Sprintf("QuotaLimit: %f", c.Resource.CpuQuotaLimit),
				fmt.Sprintf("MemoryRequest: %v", c.Resource.MemoryRequest),
				fmt.Sprintf("MemoryLimit: %v", c.Resource.MemoryLimit),
				fmt.Sprintf("StorageRequest: %v", c.Resource.StorageRequest),
				fmt.Sprintf("StorageLimit: %v", c.Resource.StorageLimit),
				fmt.Sprintf("Privileged: %v", c.Privileged),
			},
			c.Resource.VolumesRequest,
			c.Resource.VolumesLimit,
			{fmt.Sprintf("VolumePlanRequest: %+v", c.Resource.VolumePlanRequest)},
			{fmt.Sprintf("VolumePlanLimit: %+v", c.Resource.VolumePlanLimit)},
			ips,
			ns,
		}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

func DescribeWorkloadStatuses(workloadStatuses ...*corepb.WorkloadStatus) {
	switch strings.ToLower(Format) {
	case "json":
		describeAsJSON(workloadStatuses)
	case "yaml", "yml":
		describeAsYAML(workloadStatuses)
	default:
		describeWorkloadStatuses(workloadStatuses)
	}
}

func describeWorkloadStatuses(workloadStatuses []*corepb.WorkloadStatus) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Status", "Networks", "Extensions"})

	for _, s := range workloadStatuses {
		// networks
		ns := []string{}
		for name, ip := range s.Networks {
			ns = append(ns, fmt.Sprintf("%s: %s", name, ip))
		}

		// extensions
		extensions := map[string]string{}
		if err := json.Unmarshal(s.Extension, &extensions); err != nil {
			continue
		}
		es := []string{}
		for k, v := range extensions {
			es = append(es, fmt.Sprintf("%s: %s", k, v))
		}

		rows := [][]string{
			{s.Id},
			{fmt.Sprintf("Running: %v", s.Running), fmt.Sprintf("Healthy: %v", s.Healthy)},
			ns,
			es,
		}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}
