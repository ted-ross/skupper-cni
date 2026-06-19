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

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/skupperproject/skupper-cni/pkg/config"
	"github.com/skupperproject/skupper-cni/pkg/ipam"
	"github.com/skupperproject/skupper-cni/pkg/tun"
)

// CmdDel implements the CNI DEL command
func CmdDel(args *skel.CmdArgs) error {
	// Parse network configuration
	conf, err := config.LoadConf(args.StdinData)
	if err != nil {
		return err
	}

	// Get the container's network namespace
	// Note: The namespace might not exist if the container has already been deleted
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		// If namespace doesn't exist, we still need to clean up IPAM
		// but we can skip the interface deletion
		if _, ok := err.(ns.NSPathNotExistErr); ok {
			// Release IP from IPAM
			ipamDelegator := ipam.NewDelegator("")
			if err := ipamDelegator.ReleaseIP(conf.IPAM, args.StdinData); err != nil {
				// Log but don't fail on IPAM release errors during cleanup
				return fmt.Errorf("failed to release IP: %v", err)
			}
			return nil
		}
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	// Create TUN interface manager
	tunMgr := tun.NewManager(conf.InterfaceName, conf.MTU)

	// Delete TUN interface from the container namespace
	if err := tunMgr.DeleteTUN(netns); err != nil {
		// Log but continue to IPAM cleanup even if interface deletion fails
		return fmt.Errorf("failed to delete TUN interface: %v", err)
	}

	// Release IP from IPAM
	ipamDelegator := ipam.NewDelegator("")
	if err := ipamDelegator.ReleaseIP(conf.IPAM, args.StdinData); err != nil {
		return fmt.Errorf("failed to release IP: %v", err)
	}

	return nil
}

// Made with Bob
