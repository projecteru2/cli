package describe

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	corepb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"

	"github.com/jedib0t/go-pretty/v6/table"
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

// WorkloadsStatistics describes the statistics of the Workloads
func WorkloadsStatistics(workloads ...*corepb.Workload) {
	stat := struct {
		CPUs    float64
		Memory  int64
		Storage int64
	}{}
	for _, w := range workloads {
		rawResourceArgs := map[string]map[string]interface{}{}
		if err := json.Unmarshal([]byte(w.ResourceArgs), &rawResourceArgs); err != nil {
			continue
		}
		cpu := rawResourceArgs["cpumem"]["cpu_request"].(float64)
		mem := rawResourceArgs["cpumem"]["memory_request"].(float64)
		storage := rawResourceArgs["volume"]["storage_request"].(float64)
		stat.CPUs += cpu
		stat.Memory += int64(coreutils.Round(mem))
		stat.Storage += int64(coreutils.Round(storage))
	}

	describeStatistics := func() {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"CPUs", "Memory", "Storage"})

		rows := [][]string{
			{fmt.Sprintf("%f", stat.CPUs)},
			{fmt.Sprintf("%d", stat.Memory)},
			{fmt.Sprintf("%d", stat.Storage)},
		}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()

		t.SetStyle(table.StyleLight)
		t.Render()
	}

	switch {
	case isJSON():
		describeAsJSON(stat)
	case isYAML():
		describeAsYAML(stat)
	default:
		describeStatistics()
	}
}

func describeWorkloads(workloads []*corepb.Workload) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	var once sync.Once

	for _, c := range workloads {
		header, cells := parseWorkloadPluginResources(c)
		once.Do(func() {
			header = append([]interface{}{"Name/ID/Pod/Node/Priviledged", "Networks"}, header...)
			t.AppendHeader(header)
		})

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
			{c.Name, c.Id, c.Podname, c.Nodename, fmt.Sprintf("Priviledged: %v", c.Privileged)},
			ns,
		}
		rows = append(rows, cells...)
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

func parseWorkloadPluginResources(workload *corepb.Workload) (header []interface{}, cells [][]string) {
	usageMap := map[string]map[string]interface{}{}
	if len(workload.ResourceArgs) > 0 {
		_ = json.Unmarshal([]byte(workload.ResourceArgs), &usageMap)
	}

	for plugin := range usageMap {
		header = append(header, plugin)
	}
	sort.Slice(header, func(i, j int) bool {
		return header[i].(string) < header[j].(string)
	})

	for _, plugin := range header {
		row := []string{}
		usage := usageMap[plugin.(string)]

		for key, value := range usage {
			row = append(row, parse(key, value)...)
		}
		cells = append(cells, row)
	}
	return
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
		if len(s.Extension) != 0 {
			if err := json.Unmarshal(s.Extension, &extensions); err != nil {
				continue
			}
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
