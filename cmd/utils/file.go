package utils

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/core/types"

	"github.com/urfave/cli/v2"
)

// ReadAllFiles open each pair in files
// and returns a map with key as dstfile, value as linux file
// files: list of srcfile:dstfile:mode:uid:gid
func ReadAllFiles(files []string) map[string]*types.LinuxFile {
	m := map[string]*types.LinuxFile{}
	for _, file := range files {
		ps := strings.Split(file, ":")
		f := &types.LinuxFile{}
		var err error

		switch {
		case len(ps) >= 5:
			// srcfile:dstfile:mode:uid:gid
			var uid, gid int64
			uid, err = strconv.ParseInt(ps[3], 10, 0)
			if err != nil {
				break
			}
			gid, err = strconv.ParseInt(ps[3], 10, 0)
			if err != nil {
				break
			}
			f.UID = int(uid)
			f.GID = int(gid)
			fallthrough
		case len(ps) >= 3:
			// srcfile:dstfile:mode
			f.Mode, err = strconv.ParseInt(ps[2], 8, 0)
			if err != nil {
				break
			}
			fallthrough
		case len(ps) >= 2:
			// srcfile:dstfile
			f.Content, err = ioutil.ReadFile(ps[0])
			if err != nil {
				break
			}
			m[ps[1]] = f
		}
	}
	return m
}

// GenerateFileOptions returns file options
func GenerateFileOptions(c *cli.Context) (map[string][]byte, map[string]*corepb.FileMode, map[string]*corepb.FileOwner) {
	data := map[string][]byte{}
	modes := map[string]*corepb.FileMode{}
	owners := map[string]*corepb.FileOwner{}

	m := ReadAllFiles(c.StringSlice("file"))
	for dst, file := range m {
		data[dst] = file.Content
		modes[dst] = &corepb.FileMode{Mode: file.Mode}
		owners[dst] = &corepb.FileOwner{Uid: int32(file.UID), Gid: int32(file.GID)}
	}

	return data, modes, owners
}

// SplitFiles transfers a list of
// src:dst to
// {src: dst}
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

// GetSpecFromRemote gets specs from a remote position
func GetSpecFromRemote(uri string) ([]byte, error) {
	resp, err := http.Get(uri) // nolint
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
