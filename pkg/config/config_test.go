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
	"testing"
)

func TestLoadConf(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			input: `{
				"cniVersion": "1.0.0",
				"name": "test-network",
				"type": "skupper-cni",
				"mtu": 1500,
				"interfaceName": "tun0",
				"ipam": {
					"type": "host-local",
					"subnet": "10.244.0.0/16"
				}
			}`,
			wantErr: false,
		},
		{
			name: "valid configuration with defaults",
			input: `{
				"cniVersion": "1.0.0",
				"name": "test-network",
				"type": "skupper-cni",
				"ipam": {
					"type": "host-local",
					"subnet": "10.244.0.0/16"
				}
			}`,
			wantErr: false,
		},
		{
			name: "missing CNI version",
			input: `{
				"name": "test-network",
				"type": "skupper-cni",
				"ipam": {
					"type": "host-local"
				}
			}`,
			wantErr: true,
			errMsg:  "CNI version must be specified",
		},
		{
			name: "invalid MTU - too low",
			input: `{
				"cniVersion": "1.0.0",
				"name": "test-network",
				"type": "skupper-cni",
				"mtu": 50,
				"ipam": {
					"type": "host-local"
				}
			}`,
			wantErr: true,
			errMsg:  "invalid MTU value",
		},
		{
			name: "invalid MTU - too high",
			input: `{
				"cniVersion": "1.0.0",
				"name": "test-network",
				"type": "skupper-cni",
				"mtu": 70000,
				"ipam": {
					"type": "host-local"
				}
			}`,
			wantErr: true,
			errMsg:  "invalid MTU value",
		},
		{
			name:    "invalid JSON",
			input:   `{invalid json}`,
			wantErr: true,
			errMsg:  "failed to parse network configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf, err := LoadConf([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadConf() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("LoadConf() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("LoadConf() unexpected error = %v", err)
				return
			}

			if conf == nil {
				t.Errorf("LoadConf() returned nil config")
				return
			}

			// Verify defaults are applied
			if conf.MTU == 0 {
				t.Errorf("LoadConf() MTU not set to default")
			}
			if conf.InterfaceName == "" {
				t.Errorf("LoadConf() InterfaceName not set to default")
			}
		})
	}
}

func TestLoadConfDefaults(t *testing.T) {
	input := `{
		"cniVersion": "1.0.0",
		"name": "test-network",
		"type": "skupper-cni",
		"ipam": {
			"type": "host-local"
		}
	}`

	conf, err := LoadConf([]byte(input))
	if err != nil {
		t.Fatalf("LoadConf() unexpected error = %v", err)
	}

	if conf.MTU != 1500 {
		t.Errorf("LoadConf() MTU = %v, want 1500", conf.MTU)
	}

	if conf.InterfaceName != "tun0" {
		t.Errorf("LoadConf() InterfaceName = %v, want tun0", conf.InterfaceName)
	}
}

// Made with Bob
