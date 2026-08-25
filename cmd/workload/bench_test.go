package workload

import (
	"fmt"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func BenchmarkFilterSkipNoIPFilter(b *testing.B) {
	f := filter{}
	workloads := benchWorkloads(5000)
	b.ReportAllocs()
	for b.Loop() {
		for _, w := range workloads {
			_ = f.skip(w)
		}
	}
}

func BenchmarkFilterSkipMatchIP(b *testing.B) {
	f := filter{ips: []string{"10.0.0.4999"}}
	workloads := benchWorkloads(5000)
	b.ReportAllocs()
	for b.Loop() {
		for _, w := range workloads {
			_ = f.skip(w)
		}
	}
}

func benchWorkloads(n int) []*corepb.Workload {
	workloads := make([]*corepb.Workload, 0, n)
	for i := range n {
		workloads = append(workloads, &corepb.Workload{
			Id:       fmt.Sprintf("cid%d", i),
			Nodename: "node1",
			Podname:  "pod1",
			Status: &corepb.WorkloadStatus{
				Networks: map[string]string{
					"bridge":  fmt.Sprintf("10.0.0.%d/24", i),
					"overlay": fmt.Sprintf("10.1.0.%d/24", i),
				},
			},
		})
	}
	return workloads
}
