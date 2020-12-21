package describe

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

// Workloads describes a list of Workload
// output format can be json or yaml or table
func Workloads(workloads ...*corepb.Workload) {
	switch {
	case isJSON():
		describeAsJSON(workloads)
	case isYAML():
		describeAsYAML(workloads)
	default:
		describeWorkloads(workloads)
	}
}

func describeWorkloads(workloads []*corepb.Workload) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name/ID/Pod/Node", "Status", "Volume", "Networks"})

	for _, c := range workloads {
		// networks
		ns := []string{}
		if c.Status != nil {
			for name, ip := range c.Status.Networks {
				if published, ok := c.Publish[name]; ok {
					addresses := strings.Split(published, ",")

					firstLine := fmt.Sprintf("%s: %s", name, addresses[0])
					ns = append(ns, firstLine)

					if len(addresses) > 1 {
						format := fmt.Sprintf("%%%ds", len(firstLine))
						for _, address := range addresses[1:] {
							ns = append(ns, fmt.Sprintf(format, address))
						}
					}
				} else {
					ns = append(ns, fmt.Sprintf("%s: %s", name, ip))
				}
			}
		}

		rows := [][]string{
			{c.Name, c.Id, c.Podname, c.Nodename},
			{
				fmt.Sprintf("QuotaRequest: %f", c.Resource.CpuQuotaRequest),
				fmt.Sprintf("QuotaLimit: %f", c.Resource.CpuQuotaLimit),
				fmt.Sprintf("MemoryRequest: %v", c.Resource.MemoryRequest),
				fmt.Sprintf("MemoryLimit: %v", c.Resource.MemoryLimit),
				fmt.Sprintf("StorageRequest: %v", c.Resource.StorageRequest),
				fmt.Sprintf("StorageLimit: %v", c.Resource.StorageLimit),
				fmt.Sprintf("Privileged: %v", c.Privileged),
			},
			{
				fmt.Sprintf("VolumePlanRequest: %+v", c.Resource.VolumePlanRequest),
				fmt.Sprintf("VolumePlanLimit: %+v", c.Resource.VolumePlanLimit),
			},
			ns,
		}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

// WorkloadStatuses describes a list of WorkloadStatus
// output format can be json or yaml or table
func WorkloadStatuses(workloadStatuses ...*corepb.WorkloadStatus) {
	switch {
	case isJSON():
		describeAsJSON(workloadStatuses)
	case isYAML():
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
