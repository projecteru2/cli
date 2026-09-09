package types

import (
	"strings"

	"github.com/projecteru2/core/types"
)

// Specs is the deploy spec file of an application.
type Specs struct {
	Appname        string                `yaml:"appname"`
	Entrypoints    map[string]Entrypoint `yaml:"entrypoints"`
	Volumes        []string              `yaml:"volumes"`
	VolumesRequest []string              `yaml:"volumes_request"`
	Labels         map[string]string     `yaml:"labels"`
	DNS            []string              `yaml:"dns"`
	ExtraHosts     []string              `yaml:"extra_hosts"`
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
