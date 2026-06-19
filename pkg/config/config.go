// Copyright 2024 Skupper Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"encoding/json"
	"fmt"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
)

// PluginConf represents the CNI network configuration for the Skupper plugin
type PluginConf struct {
	types.NetConf

	// MTU for the TUN interface (optional, defaults to 1500)
	MTU int `json:"mtu,omitempty"`

	// InterfaceName is the name of the TUN interface to create (optional, defaults to "tun0")
	InterfaceName string `json:"interfaceName,omitempty"`

	// RuntimeConfig contains runtime-specific configuration
	RuntimeConfig *struct {
		// Additional runtime configuration can be added here
	} `json:"runtimeConfig,omitempty"`

	// RawPrevResult is the previous result in the chain (for chaining)
	RawPrevResult map[string]interface{} `json:"prevResult,omitempty"`
	PrevResult    types.Result           `json:"-"`
}

// LoadConf parses and validates the CNI network configuration
func LoadConf(bytes []byte) (*PluginConf, error) {
	conf := &PluginConf{}
	if err := json.Unmarshal(bytes, conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}

	// Validate CNI version
	if conf.CNIVersion == "" {
		return nil, fmt.Errorf("CNI version must be specified")
	}

	// Set defaults
	if conf.MTU == 0 {
		conf.MTU = 1500
	}

	if conf.InterfaceName == "" {
		conf.InterfaceName = "tun0"
	}

	// Validate MTU
	if conf.MTU < 68 || conf.MTU > 65535 {
		return nil, fmt.Errorf("invalid MTU value: %d (must be between 68 and 65535)", conf.MTU)
	}

	// Parse previous result if present (for chaining)
	if conf.RawPrevResult != nil {
		resultBytes, err := json.Marshal(conf.RawPrevResult)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize previous result: %v", err)
		}

		conf.PrevResult, err = version.NewResult(conf.CNIVersion, resultBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse previous result: %v", err)
		}
	}

	return conf, nil
}

// Made with Bob
