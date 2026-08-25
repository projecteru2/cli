package workload

import (
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func TestFilterSkip(t *testing.T) {
	workload := &corepb.Workload{
		Podname:  "dev",
		Nodename: "node1",
		Status:   &corepb.WorkloadStatus{Networks: map[string]string{"bridge": "10.0.0.2/24"}},
	}

	tests := []struct {
		name     string
		filter   filter
		workload *corepb.Workload
		want     bool
	}{
		{name: "no filter keeps the workload", filter: filter{}, workload: workload},
		{name: "nil workload is skipped", filter: filter{}, want: true},
		{name: "matching node", filter: filter{nodenames: []string{"node1"}}, workload: workload},
		{name: "other node", filter: filter{nodenames: []string{"node2"}}, workload: workload, want: true},
		{name: "matching pod", filter: filter{podnames: []string{"dev"}}, workload: workload},
		{name: "other pod", filter: filter{podnames: []string{"prod"}}, workload: workload, want: true},
		{name: "matching ip", filter: filter{ips: []string{"10.0.0.2"}}, workload: workload},
		{name: "other ip", filter: filter{ips: []string{"10.0.0.3"}}, workload: workload, want: true},
		{name: "skipped ip", filter: filter{skipIPs: []string{"10.0.0.2"}}, workload: workload, want: true},
		{
			name:     "workload without status is kept",
			filter:   filter{ips: []string{"10.0.0.3"}},
			workload: &corepb.Workload{Podname: "dev", Nodename: "node1"},
		},
		{
			name:     "workload without status still honours the pod filter",
			filter:   filter{podnames: []string{"prod"}},
			workload: &corepb.Workload{Podname: "dev", Nodename: "node1"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.skip(tt.workload); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterIn(t *testing.T) {
	f := filter{podnames: []string{"dev"}}
	got := f.filterIn([]*corepb.Workload{
		{Id: "a", Podname: "dev", Status: &corepb.WorkloadStatus{}},
		{Id: "b", Podname: "prod", Status: &corepb.WorkloadStatus{}},
	})
	if len(got) != 1 || got[0].Id != "a" {
		t.Fatalf("got %v, want only workload a", got)
	}
}
