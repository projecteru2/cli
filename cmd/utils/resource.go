package utils

import (
	"encoding/json"
	"fmt"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/urfave/cli/v3"
)

const (
	ResourceCPUMem  = "cpumem"
	ResourceStorage = "resource-storage"

	FlagExtraResources = "extra-resources"
	FlagFile           = "file"
)

// StorageParams builds the storage plugin request; zero values stay out, so an untouched entry defers to --extra-resources.
func StorageParams(storageRequest, storageLimit int64, volumesRequest, volumesLimit []string) resourcetypes.RawParams {
	params := resourcetypes.RawParams{}
	if storageRequest != 0 {
		params["storage-request"] = storageRequest
	}
	if storageLimit != 0 {
		params["storage-limit"] = storageLimit
	}
	if len(volumesRequest) != 0 {
		params["volumes-request"] = volumesRequest
	}
	if len(volumesLimit) != 0 {
		params["volumes-limit"] = volumesLimit
	}
	return params
}

// EncodeResources encodes plugin params for the core rpc; --extra-resources fills only the plugins not already present.
func EncodeResources(cmd *cli.Command, resources resourcetypes.Resources) (map[string][]byte, error) {
	encoded := make(map[string][]byte, len(resources))
	for plugin, params := range resources {
		if len(params) == 0 {
			continue
		}
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
	extraResources := cmd.String(FlagExtraResources)
	if extraResources == "" {
		return nil, nil
	}
	parsed := map[string]any{}
	if err := json.Unmarshal([]byte(extraResources), &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}
