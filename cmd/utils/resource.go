package utils

import (
	"encoding/json"
	"fmt"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/urfave/cli/v3"
)

// EncodeResources encodes plugin params for the core rpc; --extra-resources fills only the plugins not already present.
func EncodeResources(cmd *cli.Command, resources resourcetypes.Resources) (map[string][]byte, error) {
	encoded := make(map[string][]byte, len(resources))
	for plugin, params := range resources {
		b, err := encodeResource(plugin, params)
		if err != nil {
			return nil, err
		}
		encoded[plugin] = b
	}

	extra, err := parseExtraResources(cmd)
	if err != nil {
		return nil, fmt.Errorf("parse extra resources: %w", err)
	}
	for plugin, params := range extra {
		if _, ok := encoded[plugin]; ok {
			continue
		}
		b, err := encodeResource(plugin, params)
		if err != nil {
			return nil, err
		}
		encoded[plugin] = b
	}
	return encoded, nil
}

func encodeResource(plugin string, params any) ([]byte, error) {
	b, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode %s resource: %w", plugin, err)
	}
	return b, nil
}

func parseExtraResources(cmd *cli.Command) (map[string]any, error) {
	extraResources := cmd.String("extra-resources")
	if extraResources == "" {
		return nil, nil
	}
	parsed := map[string]any{}
	if err := json.Unmarshal([]byte(extraResources), &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}
