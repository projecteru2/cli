package commands

import (
	"strings"

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
