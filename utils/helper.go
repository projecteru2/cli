package utils

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

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

// EnvParser .
func EnvParser(b []byte) ([]byte, error) {
	tmpl, err := template.New("tmpl").
		Option("missingkey=default").
		Parse(string(b))
	if err != nil {
		return b, err
	}
	out := bytes.Buffer{}
	err = tmpl.Execute(&out, splitEnv(os.Environ()))
	return out.Bytes(), err
}

func splitEnv(env []string) map[string]interface{} {
	r := map[string]interface{}{}
	for _, e := range env {
		p := strings.SplitN(e, "=", 2)
		r[p[0]] = p[1]
	}
	return r
}
