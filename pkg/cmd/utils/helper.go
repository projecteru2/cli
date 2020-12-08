package utils

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
)

func GetNetworks(network string) map[string]string {
	var ip string
	networkInfo := strings.Split(network, "=")
	if len(networkInfo) == 2 {
		network = networkInfo[0]
		ip = networkInfo[1]
	}
	networkmode := container.NetworkMode(network)
	networks := map[string]string{}
	if network != "" && networkmode.IsUserDefined() {
		networks[network] = ip
	}
	return networks
}

func ParseRAMInHuman(ram string) (int64, error) {
	if ram == "" {
		return 0, nil
	}
	flag := int64(1)
	if strings.HasPrefix(ram, "-") {
		flag = int64(-1)
		ram = strings.TrimLeft(ram, "-")
	}
	ramInBytes, err := units.RAMInBytes(ram)
	if err != nil {
		return 0, err
	}
	return ramInBytes * flag, nil
}

// SplitEquality transfers a list of
// aaa=bbb, xxx=yyy into
// {aaa:bbb, xxx:yyy}
func SplitEquality(elements []string) map[string]string {
	r := map[string]string{}
	for _, e := range elements {
		p := strings.SplitN(e, "=", 2)
		if len(p) != 2 {
			continue
		}
		r[p[0]] = p[1]
	}
	return r
}

// EnvParser .
func EnvParser(b []byte) ([]byte, error) {
	tmpl, err := template.New("tmpl").Option("missingkey=default").Parse(string(b))
	if err != nil {
		return b, err
	}
	out := bytes.Buffer{}
	err = tmpl.Execute(&out, SplitEquality(os.Environ()))
	return out.Bytes(), err
}
