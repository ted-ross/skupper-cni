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
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/skupperproject/skupper-cni/pkg/config"
	"github.com/skupperproject/skupper-cni/pkg/ipam"
	"github.com/skupperproject/skupper-cni/pkg/tun"
)

// CmdAdd implements the CNI ADD command
func CmdAdd(args *skel.CmdArgs) error {
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

	// Create IPAM delegator
	ipamDelegator := ipam.NewDelegator("")

	// Allocate IP address from IPAM plugin
	ipamResult, err := ipamDelegator.AllocateIP(conf.IPAM, args.StdinData)
	if err != nil {
		return fmt.Errorf("failed to allocate IP: %v", err)
	}

	// Ensure we have at least one IP address
	if len(ipamResult.IPs) == 0 {
		return fmt.Errorf("IPAM plugin returned no IP addresses")
	}

	// Get the first IP address
	ipConfig := ipamResult.IPs[0]

	// Create TUN interface manager
	tunMgr := tun.NewManager(conf.InterfaceName, conf.MTU)

	// Create TUN interface in the container namespace
	if err := tunMgr.CreateTUN(netns); err != nil {
		// Clean up IPAM allocation on failure
		_ = ipamDelegator.ReleaseIP(conf.IPAM, args.StdinData)
		return fmt.Errorf("failed to create TUN interface: %v", err)
	}

	// Configure the TUN interface with the allocated IP
	if err := tunMgr.ConfigureInterface(netns, &ipConfig.Address); err != nil {
		// Clean up on failure
		_ = tunMgr.DeleteTUN(netns)
		_ = ipamDelegator.ReleaseIP(conf.IPAM, args.StdinData)
		return fmt.Errorf("failed to configure TUN interface: %v", err)
	}

	// Build the result to return
	result := &current.Result{
		CNIVersion: conf.CNIVersion,
		IPs:        ipamResult.IPs,
		Routes:     ipamResult.Routes,
		DNS:        conf.DNS,
		Interfaces: []*current.Interface{
			{
				Name:    conf.InterfaceName,
				Sandbox: args.Netns,
			},
		},
	}

	// Associate IPs with the interface
	for _, ip := range result.IPs {
		ip.Interface = current.Int(0)
	}

	return types.PrintResult(result, conf.CNIVersion)
}

// Made with Bob
