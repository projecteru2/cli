package workload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type sendLargeWorkloadsOptions struct {
	client corepb.CoreRPCClient
	ids    []string
	dst    string
	src    io.Reader
	size   int64
	modes  *corepb.FileMode
	owners *corepb.FileOwner
}

func (o *sendLargeWorkloadsOptions) run(ctx context.Context) error {
	logger := log.WithFunc("workload.sendLargeWorkloadsOptions.run")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
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

	chunk := make([]byte, types.SendLargeFileChunkSize)
	for {
		n, readErr := io.ReadFull(o.src, chunk)
		if n == 0 && errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil && !errors.Is(readErr, io.EOF) && !errors.Is(readErr, io.ErrUnexpectedEOF) {
			cancel()
			wg.Wait()
			return readErr
		}
		if err := stream.Send(&corepb.FileOptions{
			Ids:   o.ids,
			Dst:   o.dst,
			Size:  o.size,
			Mode:  o.modes,
			Owner: o.owners,
			Chunk: chunk[:n],
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
		if readErr != nil {
			break
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

	files := cmd.StringSlice(utils.FlagFile)
	if len(files) != 1 {
		return errors.New("sendlarge takes exactly one --file")
	}
	spec, err := utils.ParseFileSpec(files[0])
	if err != nil {
		return err
	}

	ids, err := argIDs(cmd)
	if err != nil {
		return err
	}

	src, err := os.Open(spec.Src)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()
	stat, err := src.Stat()
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		return fmt.Errorf("%s is empty, nothing to send", spec.Src)
	}
	o := &sendLargeWorkloadsOptions{
		client: client,
		ids:    ids,
		dst:    spec.Dst,
		src:    src,
		size:   stat.Size(),
		modes:  &corepb.FileMode{Mode: spec.File.Mode},
		owners: &corepb.FileOwner{Uid: int32(spec.File.UID), Gid: int32(spec.File.GID)}, //nolint:gosec
	}
	return o.run(ctx)
}
