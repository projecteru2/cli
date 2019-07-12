package commands

import (
	"strings"

	"github.com/docker/go-units"

	enginecontainer "github.com/docker/docker/api/types/container"
)

func makeLabels(labels []string) map[string]string {
	ext := map[string]string{}
	for _, d := range labels {
		if d == "" {
			continue
		}
		p := strings.Split(d, "=")
		if len(p) != 2 {
			continue
		}
		ext[p[0]] = p[1]
	}
	return ext
}

func getNetworks(network string) map[string]string {
	var ip string
	networkInfo := strings.Split(network, "=")
	if len(networkInfo) == 2 {
		network = networkInfo[0]
		ip = networkInfo[1]
	}
	networkmode := enginecontainer.NetworkMode(network)
	networks := map[string]string{}
	if network != "" && networkmode.IsUserDefined() {
		networks[network] = ip
	}
	return networks
}

func parseRAMInHuman(ramStr string) (int64, error) {
	if ramStr == "" {
		return 0, nil
	}
	flag := int64(1)
	if strings.HasPrefix(ramStr, "-") {
		flag = int64(-1)
		ramStr = strings.TrimLeft(ramStr, "-")
	}
	ramInBytes, err := units.RAMInBytes(ramStr)
	if err != nil {
		return 0, err
	}
	return ramInBytes * flag, nil
}
