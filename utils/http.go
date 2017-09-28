package utils

import (
	"io/ioutil"
	"net/http"
)

func GetSpecFromRemote(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
