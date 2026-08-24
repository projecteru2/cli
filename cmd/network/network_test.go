package network

import (
	"context"
	"errors"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"google.golang.org/grpc"
)

func TestConnectNetworkReturnsEveryFailure(t *testing.T) {
	want := errors.New("no such workload")
	o := &connectNetworkOptions{
		client:  &fakeNetworkClient{err: want},
		ids:     []string{"c1", "c2"},
		network: "bridge",
	}

	err := o.run(t.Context())
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}

func TestConnectNetworkSucceedsWhenEveryLegSucceeds(t *testing.T) {
	o := &connectNetworkOptions{
		client:  &fakeNetworkClient{},
		ids:     []string{"c1"},
		network: "bridge",
	}

	if err := o.run(t.Context()); err != nil {
		t.Fatalf("got %v, want nil", err)
	}
}

func TestDisconnectNetworkReturnsEveryFailure(t *testing.T) {
	want := errors.New("no such workload")
	o := &disconnectNetworkOptions{
		client:  &fakeNetworkClient{err: want},
		ids:     []string{"c1", "c2"},
		network: "bridge",
	}

	err := o.run(t.Context())
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}

type fakeNetworkClient struct {
	corepb.CoreRPCClient
	err error
}

func (f *fakeNetworkClient) ConnectNetwork(context.Context, *corepb.ConnectNetworkOptions, ...grpc.CallOption) (*corepb.Network, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &corepb.Network{Name: "bridge"}, nil
}

func (f *fakeNetworkClient) DisconnectNetwork(context.Context, *corepb.DisconnectNetworkOptions, ...grpc.CallOption) (*corepb.Empty, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &corepb.Empty{}, nil
}
