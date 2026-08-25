package interactive

import (
	"io"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func BenchmarkOutputWriter(b *testing.B) {
	msg := &corepb.AttachWorkloadMessage{WorkloadId: "cid1", Data: []byte("a line of workload output\n")}
	write := outputWriter(false)
	b.ReportAllocs()
	for b.Loop() {
		if err := write(io.Discard, msg); err != nil {
			b.Fatalf("write: %v", err)
		}
	}
}

func BenchmarkOutputWriterWithID(b *testing.B) {
	msg := &corepb.AttachWorkloadMessage{WorkloadId: "cid1", Data: []byte("a line of workload output\n")}
	write := outputWriter(true)
	b.ReportAllocs()
	for b.Loop() {
		if err := write(io.Discard, msg); err != nil {
			b.Fatalf("write: %v", err)
		}
	}
}
