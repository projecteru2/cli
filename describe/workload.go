package describe

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
)

// Workloads describes workloads as json, yaml or a table.
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

// WorkloadsStatistics describes the aggregated resource use of workloads.
func WorkloadsStatistics(workloads ...*corepb.Workload) {
	stat := struct {
		CPUs    float64
		Memory  int64
		Storage int64
	}{}
	for _, w := range workloads {
		res := resourcetypes.Resources{}
		if err := json.Unmarshal([]byte(w.Resources), &res); err != nil {
			continue
		}
		stat.CPUs += res["cpumem"].Float64("cpu_request")
		stat.Memory += int64(coreutils.Round(res["cpumem"].Float64("memory_request")))
		stat.Storage += int64(coreutils.Round(res["storage"].Float64("storage_request")))
	}

	switch {
	case isJSON():
		describeAsJSON(stat)
	case isYAML():
		describeAsYAML(stat)
	default:
		describeStatistics(stat.CPUs, stat.Memory, stat.Storage)
	}
}

// WorkloadStatuses describes workload statuses as json, yaml or a table.
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

func describeStatistics(cpus float64, memory, storage int64) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"CPUs", "Memory", "Storage"})

	rows := [][]string{
		{fmt.Sprintf("%f", cpus)},
		{strconv.FormatInt(memory, 10)},
		{strconv.FormatInt(storage, 10)},
	}
	t.AppendRows(toTableRows(rows))
	t.AppendSeparator()

	t.SetStyle(table.StyleLight)
	t.Render()
}

func describeWorkloads(workloads []*corepb.Workload) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	first := true
	for _, c := range workloads {
		header, cells := parseWorkloadPluginResources(c)
		if first {
			first = false
			header = append([]any{"Name/ID/Pod/Node/Privileged", "Networks"}, header...)
			t.AppendHeader(header)
		}

		ns := []string{}
		if c.Status != nil {
			for _, name := range slices.Sorted(maps.Keys(c.Status.Networks)) {
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
					ns = append(ns, fmt.Sprintf("%s: %s", name, c.Status.Networks[name]))
				}
			}
		}

		rows := [][]string{
			{c.Name, c.Id, c.Podname, c.Nodename, fmt.Sprintf("Privileged: %v", c.Privileged)},
			ns,
		}
		rows = append(rows, cells...)
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

func parseWorkloadPluginResources(workload *corepb.Workload) (header []any, cells [][]string) {
	usages := resourcetypes.Resources{}
	if len(workload.Resources) > 0 {
		_ = json.Unmarshal([]byte(workload.Resources), &usages)
	}

	for _, plugin := range slices.Sorted(maps.Keys(usages)) {
		header = append(header, plugin)

		row := []string{}
		for _, key := range slices.Sorted(maps.Keys(usages[plugin])) {
			row = append(row, parse(key, usages[plugin][key])...)
		}
		cells = append(cells, row)
	}
	return header, cells
}

func describeWorkloadStatuses(workloadStatuses []*corepb.WorkloadStatus) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Status", "Networks", "Extensions"})

	for _, s := range workloadStatuses {
		ns := []string{}
		for _, name := range slices.Sorted(maps.Keys(s.Networks)) {
			ns = append(ns, fmt.Sprintf("%s: %s", name, s.Networks[name]))
		}

		extensions := map[string]string{}
		if len(s.Extension) != 0 {
			if err := json.Unmarshal(s.Extension, &extensions); err != nil {
				continue
			}
		}
		es := []string{}
		for _, k := range slices.Sorted(maps.Keys(extensions)) {
			es = append(es, fmt.Sprintf("%s: %s", k, extensions[k]))
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
