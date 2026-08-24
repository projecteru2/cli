package workload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"sync"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type sendLargeWorkloadsOptions struct {
	client  corepb.CoreRPCClient
	ids     []string
	dst     string
	content []byte
	modes   *corepb.FileMode
	owners  *corepb.FileOwner
}

func (o *sendLargeWorkloadsOptions) run(ctx context.Context) error {
	logger := log.WithFunc("workload.sendLargeWorkloadsOptions.run")
	stream, err := o.client.SendLargeFile(ctx)
	if err != nil {
		logger.Errorf(ctx, err, "send %s failed", o.dst)
		return err
	}

	var recvErr error
	wg := sync.WaitGroup{}
	defer wg.Wait()
	wg.Go(func() {
		for {
			msg, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				recvErr = errors.Join(recvErr, err)
				return
			}

			if msg.Error != "" {
				recvErr = errors.Join(recvErr, fmt.Errorf("send %s to %s: %s", msg.Path, msg.Id, msg.Error))
				continue
			}
			logger.Infof(ctx, "send %s to %s success", msg.Path, msg.Id)
		}
	})

	for _, chunk := range o.toSendLargeFileChunks() {
		if err := stream.Send(chunk); err != nil {
			logger.Errorf(ctx, err, "send %s failed", chunk.Dst)
			return err
		}
	}
	if err := stream.CloseSend(); err != nil {
		return err
	}

	wg.Wait()
	return recvErr
}

func (o *sendLargeWorkloadsOptions) toSendLargeFileChunks() []*corepb.FileOptions {
	ret := make([]*corepb.FileOptions, 0)
	for chunk := range slices.Chunk(o.content, types.SendLargeFileChunkSize) {
		ret = append(ret, &corepb.FileOptions{
			Ids:   o.ids,
			Dst:   o.dst,
			Size:  int64(len(o.content)),
			Mode:  o.modes,
			Owner: o.owners,
			Chunk: chunk,
		})
	}
	return ret
}

func cmdWorkloadSendLarge(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	files, err := utils.GenerateFileOptions(cmd)
	if err != nil {
		return err
	}
	if len(files.Data) == 0 {
		return errors.New("files should not be empty")
	}
	if len(files.Data) >= 2 {
		return errors.New("can not send multiple files at the same time")
	}

	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return errors.New("workload id(s) should not be empty")
	}

	dst := slices.Collect(maps.Keys(files.Data))[0]
	o := &sendLargeWorkloadsOptions{
		client:  client,
		ids:     ids,
		dst:     dst,
		content: files.Data[dst],
		modes:   files.Modes[dst],
		owners:  files.Owners[dst],
	}
	return o.run(ctx)
}
