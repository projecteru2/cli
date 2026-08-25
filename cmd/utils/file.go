package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v3"
)

// FileOptions carries the --file pairs in the shape the core rpc expects.
type FileOptions struct {
	Data   map[string][]byte
	Modes  map[string]*corepb.FileMode
	Owners map[string]*corepb.FileOwner
}

// GenerateFileOptions reads the --file pairs into data, mode and owner maps.
func GenerateFileOptions(cmd *cli.Command) (*FileOptions, error) {
	files, err := ReadAllFiles(cmd.StringSlice("file"))
	if err != nil {
		return nil, err
	}

	o := &FileOptions{
		Data:   map[string][]byte{},
		Modes:  map[string]*corepb.FileMode{},
		Owners: map[string]*corepb.FileOwner{},
	}
	for dst, file := range files {
		o.Data[dst] = file.Content
		o.Modes[dst] = &corepb.FileMode{Mode: file.Mode}
		o.Owners[dst] = &corepb.FileOwner{Uid: int32(file.UID), Gid: int32(file.GID)} //nolint:gosec
	}
	return o, nil
}

// ReadAllFiles reads srcfile:dstfile[:mode[:uid:gid]] pairs into a dstfile keyed map.
func ReadAllFiles(files []string) (map[string]*types.LinuxFile, error) {
	m := map[string]*types.LinuxFile{}
	for _, file := range files {
		ps := strings.Split(file, ":")
		if len(ps) != 2 && len(ps) != 3 && len(ps) != 5 {
			return nil, fmt.Errorf("invalid file %q, want src:dst[:mode[:uid:gid]]", file)
		}

		f := &types.LinuxFile{}
		if len(ps) == 5 {
			uid, err := strconv.ParseInt(ps[3], 10, 0)
			if err != nil {
				return nil, fmt.Errorf("parse uid of %q: %w", file, err)
			}
			gid, err := strconv.ParseInt(ps[4], 10, 0)
			if err != nil {
				return nil, fmt.Errorf("parse gid of %q: %w", file, err)
			}
			f.UID = int(uid)
			f.GID = int(gid)
		}
		if len(ps) >= 3 {
			mode, err := strconv.ParseInt(ps[2], 8, 0)
			if err != nil {
				return nil, fmt.Errorf("parse mode of %q: %w", file, err)
			}
			f.Mode = mode
		}

		content, err := os.ReadFile(ps[0]) //nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", ps[0], err)
		}
		f.Content = content
		m[ps[1]] = f
	}
	return m, nil
}

// SplitFiles turns a list of src:dst strings into a map.
func SplitFiles(files []string) map[string]string {
	ret := map[string]string{}
	for _, f := range files {
		ps := strings.Split(f, ":")
		if len(ps) < 2 {
			continue
		}
		ret[ps[0]] = ps[1]
	}
	return ret
}

// GetSpecFromRemote fetches a spec over HTTP.
func GetSpecFromRemote(ctx context.Context, uri string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: %s", uri, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
