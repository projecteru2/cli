package utils

import (
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v3"
)

// ReadAllFiles reads srcfile:dstfile[:mode[:uid:gid]] pairs into a dstfile keyed map.
func ReadAllFiles(files []string) map[string]*types.LinuxFile {
	m := map[string]*types.LinuxFile{}
	for _, file := range files {
		ps := strings.Split(file, ":")
		f := &types.LinuxFile{}
		var err error

		switch {
		case len(ps) >= 5:
			var uid, gid int64
			uid, err = strconv.ParseInt(ps[3], 10, 0)
			if err != nil {
				break
			}
			gid, err = strconv.ParseInt(ps[4], 10, 0)
			if err != nil {
				break
			}
			f.UID = int(uid)
			f.GID = int(gid)
			fallthrough
		case len(ps) >= 3:
			f.Mode, err = strconv.ParseInt(ps[2], 8, 0)
			if err != nil {
				break
			}
			fallthrough
		case len(ps) >= 2:
			f.Content, err = os.ReadFile(ps[0])
			if err != nil {
				break
			}
			m[ps[1]] = f
		}
	}
	return m
}

// GenerateFileOptions reads the --file pairs into data, mode and owner maps.
func GenerateFileOptions(cmd *cli.Command) (map[string][]byte, map[string]*corepb.FileMode, map[string]*corepb.FileOwner) {
	data := map[string][]byte{}
	modes := map[string]*corepb.FileMode{}
	owners := map[string]*corepb.FileOwner{}

	m := ReadAllFiles(cmd.StringSlice("file"))
	for dst, file := range m {
		data[dst] = file.Content
		modes[dst] = &corepb.FileMode{Mode: file.Mode}
		owners[dst] = &corepb.FileOwner{Uid: int32(file.UID), Gid: int32(file.GID)} //nolint:gosec
	}

	return data, modes, owners
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
	return io.ReadAll(resp.Body)
}
