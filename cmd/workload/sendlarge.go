package workload

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/core/types"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type sendLargeWorkloadsOptions struct {
	client corepb.CoreRPCClient
	// workload ids
	ids     []string
	dst     string
	content []byte
	modes   *corepb.FileMode
	owners  *corepb.FileOwner
}

func (o *sendLargeWorkloadsOptions) run(ctx context.Context) error {
	stream, err := o.client.SendLargeFile(ctx)
	if err != nil {
		logrus.Errorf("[SendLarge] Failed send %s", o.dst)
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
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
				logrus.Errorf("[SendLarge] Failed send %s to %s", msg.Path, msg.Id)
			} else {
				logrus.Infof("[SendLarge] Send %s to %s success", msg.Path, msg.Id)
			}
		}
	}()

	fileOptions := o.toSendLargeFileChunks()
	for _, chunk := range fileOptions {
		err := stream.Send(chunk)
		if err != nil {
			logrus.Errorf("[SendLarge] Failed send %s", chunk.Dst)
			return err
		}
	}
	stream.CloseSend()
	wg.Wait()
	return nil
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

func cmdWorkloadSendLarge(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	content, modes, owners := utils.GenerateFileOptions(c)
	if len(content) == 0 {
		return fmt.Errorf("files should not be empty")
	}
	if len(content) >= 2 {
		return fmt.Errorf("can not send multiple files at the same time")
	}

	ids := c.Args().Slice()
	if len(ids) == 0 {
		return fmt.Errorf("Workload ID(s) should not be empty")
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
	return o.run(c.Context)
}
