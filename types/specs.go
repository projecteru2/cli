package types

import (
	"strings"

	"github.com/projecteru2/core/types"
)

// Specs is the deploy spec file of an application.
type Specs struct {
	Appname        string                `yaml:"appname,omitempty"`
	Entrypoints    map[string]Entrypoint `yaml:"entrypoints,omitempty,flow"`
	Volumes        []string              `yaml:"volumes,omitempty,flow"`
	VolumesRequest []string              `yaml:"volumes_request,omitempty,flow"`
	Labels         map[string]string     `yaml:"labels,omitempty,flow"`
	DNS            []string              `yaml:"dns,omitempty,flow"`
	ExtraHosts     []string              `yaml:"extra_hosts,omitempty,flow"`
}

// Entrypoint accepts both the legacy `cmd` string and the current `commands` list.
type Entrypoint struct {
	types.Entrypoint `yaml:",inline"`
	Command          string `yaml:"cmd,omitempty"`
}

func (e Entrypoint) GetCommands() []string {
	if len(e.Commands) > 0 {
		return e.Commands
	}
	return strings.Fields(e.Command)
}
