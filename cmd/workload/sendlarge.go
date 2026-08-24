package workload

import (
	"context"
	"errors"
	"io"
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
	wg.Add(1)
	defer wg.Wait()
	go func() {
		defer wg.Done()
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
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
	}()

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
	maxChunkSize := types.SendLargeFileChunkSize
	ret := make([]*corepb.FileOptions, 0)
	for idx := 0; idx < len(o.content); idx += maxChunkSize {
		fileOption := &corepb.FileOptions{
			Ids:   o.ids,
			Dst:   o.dst,
			Size:  int64(len(o.content)),
			Mode:  o.modes,
			Owner: o.owners,
		}
		if idx+maxChunkSize > len(o.content) {
			fileOption.Chunk = o.content[idx:]
		} else {
			fileOption.Chunk = o.content[idx : idx+maxChunkSize]
		}
		ret = append(ret, fileOption)
	}
	return ret
}

func cmdWorkloadSendLarge(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	content, modes, owners := utils.GenerateFileOptions(cmd)
	if len(content) == 0 {
		return errors.New("files should not be empty")
	}
	if len(content) >= 2 {
		return errors.New("can not send multiple files at the same time")
	}

	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return errors.New("workload id(s) should not be empty")
	}

	targetFileName := func() string {
		for key := range content {
			return key
		}
		return ""
	}()
	o := &sendLargeWorkloadsOptions{
		client:  client,
		ids:     ids,
		dst:     targetFileName,
		content: content[targetFileName],
		modes:   modes[targetFileName],
		owners:  owners[targetFileName],
	}
	return o.run(ctx)
}
