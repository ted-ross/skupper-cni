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

package tun

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name          string
		interfaceName string
		mtu           int
	}{
		{
			name:          "default values",
			interfaceName: "tun0",
			mtu:           1500,
		},
		{
			name:          "custom values",
			interfaceName: "tun1",
			mtu:           9000,
		},
		{
			name:          "minimum MTU",
			interfaceName: "tun2",
			mtu:           68,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.interfaceName, tt.mtu)

			if mgr == nil {
				t.Errorf("NewManager() returned nil")
				return
			}

			if mgr.interfaceName != tt.interfaceName {
				t.Errorf("NewManager() interfaceName = %v, want %v", mgr.interfaceName, tt.interfaceName)
			}

			if mgr.mtu != tt.mtu {
				t.Errorf("NewManager() mtu = %v, want %v", mgr.mtu, tt.mtu)
			}
		})
	}
}

func TestGetInterfaceName(t *testing.T) {
	interfaceName := "test-tun"
	mgr := NewManager(interfaceName, 1500)

	if mgr.GetInterfaceName() != interfaceName {
		t.Errorf("GetInterfaceName() = %v, want %v", mgr.GetInterfaceName(), interfaceName)
	}
}

// Note: Testing actual TUN interface creation, configuration, and deletion
// requires root privileges and a real network namespace. These tests would
// typically be run as integration tests in a proper test environment.
// For unit tests, we focus on testing the manager creation and basic methods.

func TestManagerStructure(t *testing.T) {
	mgr := NewManager("tun0", 1500)

	// Verify the manager has the expected structure
	if mgr.interfaceName == "" {
		t.Errorf("Manager interfaceName should not be empty")
	}

	if mgr.mtu <= 0 {
		t.Errorf("Manager mtu should be positive, got %d", mgr.mtu)
	}
}

// Made with Bob
