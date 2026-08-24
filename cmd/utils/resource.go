package utils

import (
	"encoding/json"
	"fmt"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/urfave/cli/v3"
)

// EncodeResources encodes resource plugin parameters for the core rpc and adds
// the --extra-resources object under the plugin names it does not already carry.
func EncodeResources(cmd *cli.Command, resources resourcetypes.Resources) (map[string][]byte, error) {
	encoded := make(map[string][]byte, len(resources))
	for plugin, params := range resources {
		if err := encodeResource(encoded, plugin, params); err != nil {
			return nil, err
		}
	}

	extra, err := parseExtraResources(cmd)
	if err != nil {
		return nil, fmt.Errorf("parse extra resources: %w", err)
	}
	for plugin, params := range extra {
		if _, ok := encoded[plugin]; ok {
			continue
		}
		if err := encodeResource(encoded, plugin, params); err != nil {
			return nil, err
		}
	}
	return encoded, nil
}

func encodeResource(encoded map[string][]byte, plugin string, params any) error {
	b, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("encode %s resource: %w", plugin, err)
	}
	encoded[plugin] = b
	return nil
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
