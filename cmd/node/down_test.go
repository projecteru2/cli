package node

import (
	"context"
	"strings"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"google.golang.org/grpc"
)

func TestNodeDownChecksLiveness(t *testing.T) {
	tests := []struct {
		name     string
		alive    bool
		wantErr  bool
		wantDown bool
	}{
		{name: "alive node is refused", alive: true, wantErr: true},
		{name: "dead node goes down", wantDown: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeNodeClient{alive: tt.alive}
			o := &setNodeDownOptions{client: client, name: "node1", check: true, checkTimeout: 1}

			err := o.run(t.Context())
			if (err != nil) != tt.wantErr {
				t.Fatalf("got %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), "alive") {
				t.Errorf("got %v, want it to say the node is alive", err)
			}
			if client.markedDown != tt.wantDown {
				t.Errorf("got markedDown=%v, want %v", client.markedDown, tt.wantDown)
			}
		})
	}
}

type fakeNodeClient struct {
	corepb.CoreRPCClient
	alive      bool
	markedDown bool
}

func (f *fakeNodeClient) GetNodeStatus(context.Context, *corepb.GetNodeStatusOptions, ...grpc.CallOption) (*corepb.NodeStatusStreamMessage, error) {
	return &corepb.NodeStatusStreamMessage{Alive: f.alive}, nil
}

func (f *fakeNodeClient) SetNode(context.Context, *corepb.SetNodeOptions, ...grpc.CallOption) (*corepb.Node, error) {
	f.markedDown = true
	return &corepb.Node{}, nil
}
