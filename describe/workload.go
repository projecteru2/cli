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

type workloadStatistics struct {
	CPUs    float64 `json:"cpus" yaml:"cpus"`
	Memory  int64   `json:"memory" yaml:"memory"`
	Storage int64   `json:"storage" yaml:"storage"`
}

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
	stat := workloadStatistics{}
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
	if len(workloads) == 0 {
		return
	}

	resources := make([]resourcetypes.Resources, len(workloads))
	plugins := map[string]struct{}{}
	for i, workload := range workloads {
		resources[i] = unmarshalResources(workload.Resources)
		for plugin := range resources[i] {
			plugins[plugin] = struct{}{}
		}
	}
	names := slices.Sorted(maps.Keys(plugins))

	header := []any{"Name/ID/Pod/Node/Privileged", "Networks"}
	for _, name := range names {
		header = append(header, name)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)

	for i, c := range workloads {
		rows := [][]string{
			{c.Name, c.Id, c.Podname, c.Nodename, fmt.Sprintf("Privileged: %v", c.Privileged)},
			workloadNetworks(c),
		}
		for _, name := range names {
			rows = append(rows, parseAll(resources[i][name]))
		}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

func workloadNetworks(workload *corepb.Workload) []string {
	addresses := map[string]string{}
	if workload.Status != nil {
		maps.Copy(addresses, workload.Status.Networks)
	}
	for name := range workload.Publish {
		if _, ok := addresses[name]; !ok {
			addresses[name] = ""
		}
	}

	ns := []string{}
	for _, name := range slices.Sorted(maps.Keys(addresses)) {
		published, ok := workload.Publish[name]
		if !ok {
			ns = append(ns, fmt.Sprintf("%s: %s", name, addresses[name]))
			continue
		}

		parts := strings.Split(published, ",")
		firstLine := fmt.Sprintf("%s: %s", name, parts[0])
		ns = append(ns, firstLine)

		format := fmt.Sprintf("%%%ds", len(firstLine))
		for _, address := range parts[1:] {
			ns = append(ns, fmt.Sprintf(format, address))
		}
	}
	return ns
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
