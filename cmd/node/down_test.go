package node

import (
	"context"
	"errors"
	"strings"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"google.golang.org/grpc"
)

func TestNodeDownChecksLiveness(t *testing.T) {
	tests := []struct {
		name     string
		alive    bool
		getErr   error
		wantErr  string
		wantDown bool
	}{
		{name: "alive node is refused", alive: true, wantErr: "alive"},
		{name: "a failed check refuses too", getErr: errors.New("core unavailable"), wantErr: "core unavailable"},
		{name: "dead node goes down", wantDown: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeNodeClient{alive: tt.alive, getErr: tt.getErr}
			o := &setNodeDownOptions{client: client, name: "node1", check: true, checkTimeout: 1}

			err := o.run(t.Context())
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("got %v, want it to mention %q", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("got %v, want nil", err)
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
	getErr     error
	markedDown bool
}

func (f *fakeNodeClient) GetNode(context.Context, *corepb.GetNodeOptions, ...grpc.CallOption) (*corepb.Node, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &corepb.Node{Available: f.alive}, nil
}

func (f *fakeNodeClient) SetNode(context.Context, *corepb.SetNodeOptions, ...grpc.CallOption) (*corepb.Node, error) {
	f.markedDown = true
	return &corepb.Node{}, nil
}
