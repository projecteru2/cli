package utils

import (
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetFilesStream get file from stream
func GetFilesStream(files []string) map[string][]byte {
	fileData := map[string][]byte{}
	if len(files) > 0 {
		for i := range files {
			paths := strings.Split(files[i], ":")
			if stream, err := ioutil.ReadFile(paths[0]); err != nil {
				log.Fatalf("Get file %s failed %v", paths[0], err)
				continue
			} else {
				fileData[paths[1]] = stream
			}
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
