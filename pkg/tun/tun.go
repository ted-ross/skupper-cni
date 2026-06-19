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
	"fmt"
	"net"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

// Manager handles TUN interface operations
type Manager struct {
	interfaceName string
	mtu           int
}

// NewManager creates a new TUN interface manager
func NewManager(interfaceName string, mtu int) *Manager {
	return &Manager{
		interfaceName: interfaceName,
		mtu:           mtu,
	}
}

// CreateTUN creates a TUN interface in the specified network namespace
func (m *Manager) CreateTUN(netns ns.NetNS) error {
	return netns.Do(func(_ ns.NetNS) error {
		// Create TUN link
		tun := &netlink.Tuntap{
			LinkAttrs: netlink.LinkAttrs{
				Name:  m.interfaceName,
				MTU:   m.mtu,
				Flags: net.FlagUp,
			},
			Mode: netlink.TUNTAP_MODE_TUN,
		}

		// Add the TUN interface
		if err := netlink.LinkAdd(tun); err != nil {
			return fmt.Errorf("failed to create TUN interface: %v", err)
		}

		return nil
	})
}

// ConfigureInterface configures the TUN interface with IP address and brings it up
func (m *Manager) ConfigureInterface(netns ns.NetNS, ipAddr *net.IPNet) error {
	return netns.Do(func(_ ns.NetNS) error {
		// Get the link
		link, err := netlink.LinkByName(m.interfaceName)
		if err != nil {
			return fmt.Errorf("failed to find interface %s: %v", m.interfaceName, err)
		}

		// Add IP address to the interface
		addr := &netlink.Addr{
			IPNet: ipAddr,
		}
		if err := netlink.AddrAdd(link, addr); err != nil {
			return fmt.Errorf("failed to add IP address to interface: %v", err)
		}

		// Bring the interface up
		if err := netlink.LinkSetUp(link); err != nil {
			return fmt.Errorf("failed to bring interface up: %v", err)
		}

		return nil
	})
}

// DeleteTUN deletes the TUN interface from the specified network namespace
func (m *Manager) DeleteTUN(netns ns.NetNS) error {
	return netns.Do(func(_ ns.NetNS) error {
		// Get the link
		link, err := netlink.LinkByName(m.interfaceName)
		if err != nil {
			// If the interface doesn't exist, consider it already deleted
			if _, ok := err.(netlink.LinkNotFoundError); ok {
				return nil
			}
			return fmt.Errorf("failed to find interface %s: %v", m.interfaceName, err)
		}

		// Delete the interface
		if err := netlink.LinkDel(link); err != nil {
			return fmt.Errorf("failed to delete interface: %v", err)
		}

		return nil
	})
}

// CheckInterface verifies that the TUN interface exists and has the expected configuration
func (m *Manager) CheckInterface(netns ns.NetNS, expectedIP *net.IPNet) error {
	return netns.Do(func(_ ns.NetNS) error {
		// Get the link
		link, err := netlink.LinkByName(m.interfaceName)
		if err != nil {
			return fmt.Errorf("interface %s not found: %v", m.interfaceName, err)
		}

		// Check if interface is up
		if link.Attrs().Flags&net.FlagUp == 0 {
			return fmt.Errorf("interface %s is not up", m.interfaceName)
		}

		// Get addresses on the interface
		addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return fmt.Errorf("failed to list addresses on interface: %v", err)
		}

		// Check if expected IP is present
		if expectedIP != nil {
			found := false
			for _, addr := range addrs {
				if addr.IPNet.String() == expectedIP.String() {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("expected IP address %s not found on interface", expectedIP.String())
			}
		}

		return nil
	})
}

// GetInterfaceName returns the name of the TUN interface
func (m *Manager) GetInterfaceName() string {
	return m.interfaceName
}

// Made with Bob
