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

package ipam

import (
	"context"
	"fmt"

	"github.com/containernetworking/cni/pkg/invoke"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
)

// Delegator handles IPAM delegation to external plugins
type Delegator struct {
	pluginPath string
}

// NewDelegator creates a new IPAM delegator
func NewDelegator(pluginPath string) *Delegator {
	return &Delegator{
		pluginPath: pluginPath,
	}
}

// AllocateIP delegates IP allocation to the configured IPAM plugin
func (d *Delegator) AllocateIP(ipamConfig types.IPAM, stdinData []byte) (*current.Result, error) {
	if ipamConfig.Type == "" {
		return nil, fmt.Errorf("IPAM type not specified in configuration")
	}

	// Execute the IPAM plugin
	r, err := invoke.DelegateAdd(context.TODO(), ipamConfig.Type, stdinData, d.createExec())
	if err != nil {
		return nil, fmt.Errorf("failed to delegate IPAM ADD: %v", err)
	}

	// Convert result to current version
	result, err := current.NewResultFromResult(r)
	if err != nil {
		return nil, fmt.Errorf("failed to convert IPAM result: %v", err)
	}

	if len(result.IPs) == 0 {
		return nil, fmt.Errorf("IPAM plugin returned no IP addresses")
	}

	return result, nil
}

// ReleaseIP delegates IP release to the configured IPAM plugin
func (d *Delegator) ReleaseIP(ipamConfig types.IPAM, stdinData []byte) error {
	if ipamConfig.Type == "" {
		return fmt.Errorf("IPAM type not specified in configuration")
	}

	// Execute the IPAM plugin DEL operation
	err := invoke.DelegateDel(context.TODO(), ipamConfig.Type, stdinData, d.createExec())
	if err != nil {
		return fmt.Errorf("failed to delegate IPAM DEL: %v", err)
	}

	return nil
}

// createExec creates an Exec instance for invoking IPAM plugins
func (d *Delegator) createExec() invoke.Exec {
	return &invoke.DefaultExec{
		RawExec: &invoke.RawExec{Stderr: nil},
	}
}

// Made with Bob
