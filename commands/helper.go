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

func filterContainer(extend map[string]string, labels map[string]string) bool {
	for k, v := range labels {
		if n, ok := extend[k]; !ok || n != v {
			return false
		}
	}
	return true
}

func getNetworks(network string) map[string]string {
	networkmode := enginecontainer.NetworkMode(network)
	networks := map[string]string{}
	if network != "" && networkmode.IsUserDefined() {
		networks[network] = ""
	}
	return networks
}
