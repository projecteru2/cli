package describe

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"sigs.k8s.io/yaml"

	"github.com/projecteru2/cli/cmd/utils"
)

const headerName = "Name"

// Format selects the output format: json, yaml, or empty for a table.
var Format string

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

// ToResourcePercent reports node usage as a fraction of capacity, per resource.
func ToResourcePercent(resource *corepb.NodeResource) (cpumem, storage map[string]float64, err error) {
	var resUsage resourcetypes.Resources
	var resCap resourcetypes.Resources
	if err := json.Unmarshal([]byte(resource.ResourceUsage), &resUsage); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal([]byte(resource.ResourceCapacity), &resCap); err != nil {
		return nil, nil, err
	}
	cpumemUsage := resUsage[utils.ResourceCPUMem]
	storageUsage := resUsage[utils.ResourceStorage]
	cpumemCap := resCap[utils.ResourceCPUMem]
	storageCap := resCap[utils.ResourceStorage]
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
			sr["storage"] = stUsage / stCap
		}
		vu := 0.0
		vc := 0.0
		for k := range volumesUsage {
			vu += volumesUsage.Float64(k)
		}
		for k := range volumesCap {
			vc += volumesCap.Float64(k)
		}
		if vc != 0 {
			sr["volumes"] = vu / vc
		}
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

	rs := make([]table.Row, 0, maxLength)
	for i := range maxLength {
		lines := make([]any, 0, total)
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

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func parse(key, value any) []string {
	res := []string{}
	switch v := value.(type) {
	case map[string]any:
		for _, k := range slices.Sorted(maps.Keys(v)) {
			res = append(res, fmt.Sprintf("%s[%s]: %v", key, k, toJSON(v[k])))
		}
	case []any:
		for i, item := range v {
			res = append(res, fmt.Sprintf("%s[%d]: %v", key, i, toJSON(item)))
		}
	default:
		res = append(res, fmt.Sprintf("%s: %v", key, toJSON(value)))
	}
	return res
}

func parseAll(params resourcetypes.RawParams) []string {
	rows := []string{}
	for _, key := range slices.Sorted(maps.Keys(params)) {
		rows = append(rows, parse(key, params[key])...)
	}
	return rows
}

func sortedKVLines(m map[string]string) []string {
	lines := make([]string, 0, len(m))
	for _, k := range slices.Sorted(maps.Keys(m)) {
		lines = append(lines, fmt.Sprintf("%s: %s", k, m[k]))
	}
	return lines
}

func pluginNames(resourceSets ...[]resourcetypes.Resources) []string {
	plugins := map[string]struct{}{}
	for _, set := range resourceSets {
		for _, resources := range set {
			for plugin := range resources {
				plugins[plugin] = struct{}{}
			}
		}
	}
	return slices.Sorted(maps.Keys(plugins))
}

func unmarshalResources(encoded string) resourcetypes.Resources {
	res := resourcetypes.Resources{}
	if len(encoded) > 0 {
		_ = json.Unmarshal([]byte(encoded), &res)
	}
	return res
}

func describeAsJSON(o any) {
	j, _ := json.MarshalIndent(o, "", "  ")
	fmt.Println(string(j))
}

func describeAsYAML(o any) {
	y, _ := yaml.Marshal(o)
	fmt.Println(string(y))
}

func describeOr[T any](v T, fallback func(T)) {
	switch {
	case isJSON():
		describeAsJSON(v)
	case isYAML():
		describeAsYAML(v)
	default:
		fallback(v)
	}
}

func describeChOr[T any](ch <-chan T, fallback func(<-chan T)) {
	collect := func() []T {
		items := []T{}
		for t := range ch {
			items = append(items, t)
		}
		return items
	}
	switch {
	case isJSON():
		describeAsJSON(collect())
	case isYAML():
		describeAsYAML(collect())
	default:
		fallback(ch)
	}
}

func renderTable(header []string, groups ...[][]string) {
	h := make(table.Row, len(header))
	for i, name := range header {
		h[i] = name
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(h)
	for _, rows := range groups {
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}
	t.SetStyle(table.StyleLight)
	t.Render()
}
