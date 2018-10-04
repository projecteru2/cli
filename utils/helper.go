package utils

import (
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// SplitFiles split file from params
func SplitFiles(files []string) map[string]string {
	ret := map[string]string{}
	for i := range files {
		paths := strings.Split(files[i], ":")
		ret[paths[0]] = paths[1]
	}
	return ret
}

// GetFilesStream get file from stream
func GetFilesStream(files []string) map[string][]byte {
	fileData := map[string][]byte{}
	for i := range files {
		paths := strings.Split(files[i], ":")
		if stream, err := ioutil.ReadFile(paths[0]); err != nil {
			log.Fatalf("Get file %s failed %v", paths[0], err)
			continue
		} else {
			fileData[paths[1]] = stream
		}
	}
	return fileData
}

// ParseEnvValue get value from env
func ParseEnvValue(f string) string {
	if !strings.HasPrefix(f, "$") {
		return f
	}

	f = strings.TrimLeft(f, "$")
	return os.Getenv(f)
}
