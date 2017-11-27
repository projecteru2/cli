package types

import (
	agenttypes "github.com/projecteru2/agent/types"
	"github.com/projecteru2/core/types"
)

// correspond to app.yaml in repository
type Specs struct {
	Appname     string                      `yaml:"appname,omitempty"`
	Entrypoints map[string]types.Entrypoint `yaml:"entrypoints,omitempty,flow"`
	Volumes     []string                    `yaml:"volumes,omitempty,flow"`
	Meta        map[string]string           `yaml:"meta,omitempty,flow"`
	DNS         []string                    `yaml:"dns,omitempty,flow"`
	ExtraHosts  []string                    `yaml:"extra_hosts,omitempty,flow"`
}

type Container struct {
	agenttypes.Container
	Nodename string
}
