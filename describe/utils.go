package describe

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/jedib0t/go-pretty/v6/table"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
)

// Format indicates the output format
// can be yaml / yml / json or empty as default
// default will be table
var Format string

func isJSON() bool {
	return strings.ToLower(Format) == "json"
}

func isYAML() bool {
	y := strings.ToLower(Format)
	return y == "yaml" || y == "yml"
}

// actually i need a `zip longest` function
// like in python itertools
func toTableRows(rows [][]string) []table.Row {
	total := len(rows)
	maxLength := 0
	for _, row := range rows {
		if len(row) > maxLength {
			maxLength = len(row)
		}
	}

	rs := []table.Row{}
	for i := 0; i < maxLength; i++ {
		lines := []interface{}{}
		for j := 0; j < total; j++ {
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

func describeAsJSON(o interface{}) {
	j, _ := json.MarshalIndent(o, "", "  ")
	fmt.Println(string(j))
}

func describeChNodeAsJSON(ch <-chan *corepb.Node) {
	for t := range ch {
		j, _ := json.MarshalIndent(t, "", "  ")
		fmt.Println(string(j))
	}
}
func describeChNodeResourceAsJSON(ch chan *corepb.NodeResource) {
	for t := range ch {
		j, _ := json.MarshalIndent(t, "", "  ")
		fmt.Println(string(j))
	}
}

func describeAsYAML(o interface{}) {
	y, _ := yaml.Marshal(o)
	fmt.Println(string(y))
}

func describeChNodeAsYAML(ch <-chan *corepb.Node) {
	for t := range ch {
		j, _ := yaml.Marshal(t)
		fmt.Println(string(j))
	}
}
func describeChNodeResourceAsYAML(ch chan *corepb.NodeResource) {
	for t := range ch {
		j, _ := yaml.Marshal(t)
		fmt.Println(string(j))
	}
}

// ToNodeChan is to be rewritten using generic
func ToNodeChan(nodes ...*corepb.Node) chan *corepb.Node {
	ch := make(chan *corepb.Node)
	go func() {
		defer close(ch)
		for _, node := range nodes {
			ch <- node
		}
	}()
	return ch
}

// ToNodeResourceChan is to be rewritten using generic
func ToNodeResourceChan(resources ...*corepb.NodeResource) chan *corepb.NodeResource {
	ch := make(chan *corepb.NodeResource)
	go func() {
		defer close(ch)
		for _, resource := range resources {
			ch <- resource
		}
	}()
	return ch
}

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
