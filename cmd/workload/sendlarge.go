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
	wg.Go(func() {
		recvErr = utils.EachMessage(stream.Recv, func(msg *corepb.SendMessage) error {
			if msg.Error != "" {
				return fmt.Errorf("send %s to %s: %s", msg.Path, msg.Id, msg.Error)
			}
			logger.Infof(ctx, "send %s to %s success", msg.Path, msg.Id)
			return nil
		})
	})

	for chunk := range slices.Chunk(o.content, types.SendLargeFileChunkSize) {
		if err := stream.Send(&corepb.FileOptions{
			Ids:   o.ids,
			Dst:   o.dst,
			Size:  int64(len(o.content)),
			Mode:  o.modes,
			Owner: o.owners,
			Chunk: chunk,
		}); err != nil {
			logger.Errorf(ctx, err, "send %s failed", o.dst)
			wg.Wait()
			if errors.Is(err, io.EOF) {
				err = nil
				if recvErr == nil {
					recvErr = errors.New("the send stream closed before every chunk landed")
				}
			}
			return errors.Join(recvErr, err)
		}
	}
	if err := stream.CloseSend(); err != nil {
		wg.Wait()
		return errors.Join(recvErr, err)
	}

	wg.Wait()
	return recvErr
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

	ids, err := argIDs(cmd)
	if err != nil {
		return err
	}

	dst := slices.Collect(maps.Keys(files.Data))[0]
	if len(files.Data[dst]) == 0 {
		return fmt.Errorf("%s is empty, nothing to send", dst)
	}
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
