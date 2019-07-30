package types

import (
	"github.com/projecteru2/core/types"
)

// Specs correspond to app.yaml in repository
type Specs struct {
	Appname     string                      `yaml:"appname,omitempty"`
	Entrypoints map[string]types.Entrypoint `yaml:"entrypoints,omitempty,flow"`
	Volumes     []string                    `yaml:"volumes,omitempty,flow"`
	Labels      map[string]string           `yaml:"labels,omitempty,flow"`
	DNS         []string                    `yaml:"dns,omitempty,flow"`
	ExtraHosts  []string                    `yaml:"extra_hosts,omitempty,flow"`
}
