package types

import (
	"strings"

	"github.com/projecteru2/core/types"
)

// Specs correspond to app.yaml in repository
type Specs struct {
	Appname        string                `yaml:"appname,omitempty"`
	Entrypoints    map[string]Entrypoint `yaml:"entrypoints,omitempty,flow"`
	Volumes        []string              `yaml:"volumes,omitempty,flow"`
	VolumesRequest []string              `yaml:"volumes_request,omitempty,flow"`
	Labels         map[string]string     `yaml:"labels,omitempty,flow"`
	DNS            []string              `yaml:"dns,omitempty,flow"`
	ExtraHosts     []string              `yaml:"extra_hosts,omitempty,flow"`
}

// Entrypoint is a facade of old stype `cmd` and new stype `commands`
type Entrypoint struct {
	types.Entrypoint `yaml:",inline"`
	Command          string `yaml:"cmd,omitempty"`
}

// GetCommands .
func (e Entrypoint) GetCommands() []string {
	if len(e.Commands) > 0 {
		return e.Commands
	}
	return strings.Split(e.Command, " ")
}
