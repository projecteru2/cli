package utils

import (
	"io/ioutil"
	"net/http"
	"strings"
)

// ReadAllFiles open each pair in files
// and returns a map with key as dstfile, value as content of srcfile
// files: list of srcfile:dstfile
func ReadAllFiles(files []string) map[string][]byte {
	m := map[string][]byte{}
	for _, file := range files {
		ps := strings.Split(file, ":")
		if len(ps) != 2 {
			continue
		}

		content, err := ioutil.ReadFile(ps[0])
		if err != nil {
			continue
		}

		m[ps[1]] = content
	}
	return m
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
