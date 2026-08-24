package workload

import (
	"context"
	"errors"
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

	wg := sync.WaitGroup{}
	defer wg.Wait()
	wg.Go(func() {
		for {
			msg, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return
			}

			if msg.Error != "" {
				logger.Errorf(ctx, errors.New(msg.Error), "send %s to %s failed", msg.Path, msg.Id)
			} else {
				logger.Infof(ctx, "send %s to %s success", msg.Path, msg.Id)
			}
		}
	})

	fileOptions := o.toSendLargeFileChunks()
	for _, chunk := range fileOptions {
		err := stream.Send(chunk)
		if err != nil {
			logger.Errorf(ctx, err, "send %s failed", chunk.Dst)
			return err
		}
	}
	return stream.CloseSend()
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
