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
	"testing"

	"github.com/containernetworking/cni/pkg/types"
)

func TestNewDelegator(t *testing.T) {
	tests := []struct {
		name       string
		pluginPath string
	}{
		{
			name:       "with plugin path",
			pluginPath: "/opt/cni/bin",
		},
		{
			name:       "without plugin path",
			pluginPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delegator := NewDelegator(tt.pluginPath)
			if delegator == nil {
				t.Errorf("NewDelegator() returned nil")
			}
			if delegator.pluginPath != tt.pluginPath {
				t.Errorf("NewDelegator() pluginPath = %v, want %v", delegator.pluginPath, tt.pluginPath)
			}
		})
	}
}

func TestAllocateIP_InvalidIPAM(t *testing.T) {
	delegator := NewDelegator("")

	// Test with empty IPAM type
	ipamConfig := types.IPAM{
		Type: "",
	}

	_, err := delegator.AllocateIP(ipamConfig, []byte(`{}`))
	if err == nil {
		t.Errorf("AllocateIP() expected error for empty IPAM type, got nil")
	}

	expectedMsg := "IPAM type not specified"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("AllocateIP() error = %v, want error containing %v", err, expectedMsg)
	}
}

func TestReleaseIP_InvalidIPAM(t *testing.T) {
	delegator := NewDelegator("")

	// Test with empty IPAM type
	ipamConfig := types.IPAM{
		Type: "",
	}

	err := delegator.ReleaseIP(ipamConfig, []byte(`{}`))
	if err == nil {
		t.Errorf("ReleaseIP() expected error for empty IPAM type, got nil")
	}

	expectedMsg := "IPAM type not specified"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("ReleaseIP() error = %v, want error containing %v", err, expectedMsg)
	}
}

func TestCreateExec(t *testing.T) {
	tests := []struct {
		name       string
		pluginPath string
	}{
		{
			name:       "with plugin path",
			pluginPath: "/opt/cni/bin",
		},
		{
			name:       "without plugin path",
			pluginPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delegator := NewDelegator(tt.pluginPath)
			exec := delegator.createExec()
			if exec == nil {
				t.Errorf("createExec() returned nil")
			}
		})
	}
}

// Made with Bob
