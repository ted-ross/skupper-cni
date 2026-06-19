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

package plugin

import (
	"fmt"
	"net"

	"github.com/containernetworking/cni/pkg/skel"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/skupperproject/skupper-cni/pkg/config"
	"github.com/skupperproject/skupper-cni/pkg/tun"
)

// CmdCheck implements the CNI CHECK command
func CmdCheck(args *skel.CmdArgs) error {
	// Parse network configuration
	conf, err := config.LoadConf(args.StdinData)
	if err != nil {
		return err
	}

	// Get the container's network namespace
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	// Parse the previous result to get expected IP configuration
	var expectedIP *current.IPConfig
	if conf.PrevResult != nil {
		prevResult, err := current.NewResultFromResult(conf.PrevResult)
		if err != nil {
			return fmt.Errorf("failed to parse previous result: %v", err)
		}

		if len(prevResult.IPs) > 0 {
			expectedIP = prevResult.IPs[0]
		}
	}

	// Create TUN interface manager
	tunMgr := tun.NewManager(conf.InterfaceName, conf.MTU)

	// Check if the TUN interface exists and is properly configured
	var expectedIPNet *net.IPNet
	if expectedIP != nil {
		expectedIPNet = &expectedIP.Address
	}

	if err := tunMgr.CheckInterface(netns, expectedIPNet); err != nil {
		return fmt.Errorf("interface check failed: %v", err)
	}

	return nil
}

// Made with Bob
