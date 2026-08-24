package describe

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"sigs.k8s.io/yaml"
)

const headerName = "Name"

// Format selects the output format: json, yaml, or empty for a table.
var Format string

// ToChan streams items over a channel.
func ToChan[T any](items ...T) chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for _, item := range items {
			ch <- item
		}
	}()
	return ch
}

// ToResourcePrecent reports node usage as a fraction of capacity, per resource.
func ToResourcePrecent(resource *corepb.NodeResource) (map[string]float64, map[string]float64, error) {
	var resUsage resourcetypes.Resources
	var resCap resourcetypes.Resources
	if err := json.Unmarshal([]byte(resource.ResourceUsage), &resUsage); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal([]byte(resource.ResourceCapacity), &resCap); err != nil {
		return nil, nil, err
	}
	cpumemUsage := resUsage["cpumem"]
	storageUsage := resUsage["storage"]
	cpumemCap := resCap["cpumem"]
	storageCap := resCap["storage"]
	cr, sr := map[string]float64{}, map[string]float64{}
	if cpumemUsage != nil && cpumemCap != nil {
		cpuUsage := cpumemUsage.Float64("cpu")
		cpuCap := cpumemCap.Float64("cpu")
		memUsage := cpumemUsage.Float64("memory")
		memCap := cpumemCap.Float64("memory")
		cr["cpu"] = 0.0
		cr["memory"] = 0.0
		if cpuCap != 0 {
			cr["cpu"] = cpuUsage / cpuCap
		}
		if memCap != 0 {
			cr["memory"] = memUsage / memCap
		}
	}
	if storageUsage != nil && storageCap != nil {
		stUsage := storageUsage.Float64("storage")
		stCap := storageCap.Float64("storage")
		volumesUsage := storageUsage.RawParams("volumes")
		volumesCap := storageCap.RawParams("volumes")
		sr["storage"] = 0.0
		sr["volumes"] = 0.0
		if stCap != 0 {
			cr["storage"] = stUsage / stCap
		}
		vu := 0.0
		vc := 0.0
		for k := range volumesUsage {
			vu += volumesUsage.Float64(k)
		}
		for k := range volumesCap {
			vc += volumesCap.Float64(k)
		}
		sr["volumes"] = vu / vc
	}
	return cr, sr, nil
}

func isJSON() bool {
	return strings.ToLower(Format) == "json"
}

func isYAML() bool {
	y := strings.ToLower(Format)
	return y == "yaml" || y == "yml"
}

func toTableRows(rows [][]string) []table.Row {
	total := len(rows)
	maxLength := 0
	for _, row := range rows {
		maxLength = max(maxLength, len(row))
	}

	rs := []table.Row{}
	for i := range maxLength {
		lines := []any{}
		for j := range total {
			if i < len(rows[j]) {
				lines = append(lines, rows[j][i])
			} else {
				lines = append(lines, "")
			}
		}
		rs = append(rs, table.Row(lines))
	}
	return rs
}

func describeAsJSON(o any) {
	j, _ := json.MarshalIndent(o, "", "  ")
	fmt.Println(string(j))
}

func describeAsYAML(o any) {
	y, _ := yaml.Marshal(o)
	fmt.Println(string(y))
}

func describeChAsJSON[T any](ch <-chan T) {
	for t := range ch {
		describeAsJSON(t)
	}
}

func describeChAsYAML[T any](ch <-chan T) {
	for t := range ch {
		describeAsYAML(t)
	}
}
