package utils

import (
	"io/ioutil"
	"net/http"
)

// GetSpecFromRemote get spec from remote
func GetSpecFromRemote(uri string) ([]byte, error) {
	resp, err := http.Get(uri) // nolint
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
