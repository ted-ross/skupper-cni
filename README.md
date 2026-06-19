# skupper-cni

Experimental CNI plugin for Skupper that creates and manages TUN interfaces with IP addresses allocated from external IPAM plugins.

## Overview

The Skupper CNI plugin is a Container Network Interface (CNI) plugin designed to create TUN (network tunnel) interfaces for Skupper networking. It delegates IP address management to external IPAM plugins like `whereabouts` or `host-local`, focusing solely on TUN interface lifecycle management.

## Features

- **TUN Interface Management**: Creates and configures TUN interfaces in container network namespaces
- **IPAM Delegation**: Integrates with external IPAM plugins (whereabouts, host-local, etc.)
- **Full CNI Compliance**: Supports CNI specification versions 0.3.0, 0.3.1, 0.4.0, 1.0.0, and 1.1.0
- **Standard Operations**: Implements ADD, DEL, CHECK, and VERSION operations
- **Minimal Routing**: Assigns IP addresses without complex routing (Skupper handles routing)
- **Configurable**: Supports custom MTU and interface naming

## Architecture

```
┌───────────────────────────────────────┐
│   Kubernetes / Container Runtime      │
└──────────────┬────────────────────────┘
               │ CNI Request
               ▼
┌───────────────────────────────────────┐
│        Skupper CNI Plugin             │
│  ┌─────────────────────────────────┐  │
│  │  ADD / DEL / CHECK / VERSION    │  │
│  └─────────────────────────────────┘  │
│  ┌──────────────┐  ┌───────────────┐  │
│  │ TUN Manager  │  │ IPAM Delegator│  │
│  └──────────────┘  └───────────────┘  │
└──────────┬──────────────────┬─────────┘
           │                  │
           ▼                  ▼
    ┌──────────┐      ┌──────────────┐
    │  Kernel  │      │ IPAM Plugin  │
    │ TUN/TAP  │      │ (whereabouts)│
    └──────────┘      └──────────────┘
```

## Installation

### Prerequisites

- Go 1.21 or later
- Linux kernel with TUN/TAP support
- CNI plugins directory (typically `/opt/cni/bin`)
- An IPAM plugin (whereabouts, host-local, etc.)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/skupperproject/skupper-cni.git
cd skupper-cni

# Build the plugin
make build

# Install to CNI plugins directory (requires root)
sudo make install

# Optionally install example configuration
sudo make install-config
```

### Using Docker

```bash
# Build Docker image
make docker-build

# The image contains the plugin binary at /opt/cni/bin/skupper-cni
```

## Configuration

### Basic Configuration

Create a CNI configuration file in `/etc/cni/net.d/`:

```json
{
  "cniVersion": "1.0.0",
  "name": "skupper-network",
  "type": "skupper-cni",
  "mtu": 1500,
  "interfaceName": "tun0",
  "ipam": {
    "type": "host-local",
    "subnet": "10.244.0.0/16",
    "rangeStart": "10.244.1.10",
    "rangeEnd": "10.244.1.250",
    "gateway": "10.244.1.1"
  }
}
```

### Configuration with Whereabouts IPAM

```json
{
  "cniVersion": "1.0.0",
  "name": "skupper-network",
  "plugins": [
    {
      "type": "skupper-cni",
      "mtu": 1500,
      "interfaceName": "tun0",
      "ipam": {
        "type": "whereabouts",
        "range": "10.244.0.0/16",
        "range_start": "10.244.1.0",
        "range_end": "10.244.1.255",
        "gateway": "10.244.1.1"
      }
    }
  ]
}
```

### Configuration Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `cniVersion` | string | Yes | - | CNI specification version (0.3.0, 0.3.1, 0.4.0, 1.0.0, or 1.1.0) |
| `name` | string | Yes | - | Network name |
| `type` | string | Yes | - | Plugin type (must be "skupper-cni") |
| `mtu` | integer | No | 1500 | MTU for the TUN interface (68-65535) |
| `interfaceName` | string | No | "tun0" | Name of the TUN interface to create |
| `ipam` | object | Yes | - | IPAM configuration (delegated to external plugin) |

## Usage

The plugin is invoked automatically by the container runtime (containerd, CRI-O, etc.) when containers are created or deleted. You don't typically invoke it directly.

### Testing with CNI Tools

You can test the plugin manually using CNI tools:

```bash
# Set up environment variables
export CNI_COMMAND=ADD
export CNI_CONTAINERID=test123
export CNI_NETNS=/var/run/netns/test
export CNI_IFNAME=tun0
export CNI_PATH=/opt/cni/bin

# Create a test network namespace
sudo ip netns add test

# Run the plugin
cat examples/10-skupper-simple.conf | sudo CNI_COMMAND=ADD CNI_CONTAINERID=test123 \
  CNI_NETNS=/var/run/netns/test CNI_IFNAME=tun0 CNI_PATH=/opt/cni/bin \
  /opt/cni/bin/skupper-cni

# Check the interface
sudo ip netns exec test ip addr show tun0

# Clean up
cat examples/10-skupper-simple.conf | sudo CNI_COMMAND=DEL CNI_CONTAINERID=test123 \
  CNI_NETNS=/var/run/netns/test CNI_IFNAME=tun0 CNI_PATH=/opt/cni/bin \
  /opt/cni/bin/skupper-cni

sudo ip netns del test
```

## Development

### Project Structure

```
skupper-cni/
├── cmd/
│   └── skupper-cni/        # Main entry point
├── pkg/
│   ├── config/             # Configuration parsing
│   ├── tun/                # TUN interface management
│   ├── ipam/               # IPAM delegation
│   ├── plugin/             # CNI operations (ADD, DEL, CHECK, VERSION)
│   └── utils/              # Utility functions
├── examples/               # Example configurations
├── Makefile               # Build automation
├── Dockerfile             # Container build
└── README.md              # This file
```

### Building

```bash
# Build the plugin
make build

# Build for Linux (cross-compile)
make build-linux

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean
```

### Testing

```bash
# Run all tests
make test

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test -v ./pkg/tun/
```

## CNI Operations

### ADD

Creates a TUN interface in the container's network namespace and assigns an IP address from the IPAM plugin.

**Flow:**
1. Parse CNI configuration
2. Delegate to IPAM plugin for IP allocation
3. Create TUN interface in container namespace
4. Configure interface with allocated IP
5. Bring interface up
6. Return result with interface details

### DEL

Removes the TUN interface and releases the IP address back to the IPAM plugin.

**Flow:**
1. Parse CNI configuration
2. Delete TUN interface from container namespace
3. Delegate to IPAM plugin to release IP
4. Return success

### CHECK

Verifies that the TUN interface exists and is properly configured.

**Flow:**
1. Parse CNI configuration and previous result
2. Check if TUN interface exists in container namespace
3. Verify interface is up
4. Verify IP address matches expected configuration
5. Return success or error

### VERSION

Returns the list of CNI specification versions supported by the plugin.

**Supported versions:** 0.3.0, 0.3.1, 0.4.0, 1.0.0, 1.1.0

## Troubleshooting

### Plugin Not Found

Ensure the plugin binary is in the CNI plugins directory:
```bash
ls -l /opt/cni/bin/skupper-cni
```

### Permission Denied

The plugin must be executable:
```bash
sudo chmod +x /opt/cni/bin/skupper-cni
```

### IPAM Plugin Errors

Ensure the IPAM plugin (whereabouts, host-local, etc.) is installed:
```bash
ls -l /opt/cni/bin/whereabouts
ls -l /opt/cni/bin/host-local
```

### TUN Device Creation Fails

Verify TUN/TAP kernel module is loaded:
```bash
lsmod | grep tun
# If not loaded:
sudo modprobe tun
```

### Debugging

Enable CNI debug logging by setting environment variables:
```bash
export CNI_DEBUG=1
export CNI_LOG_FILE=/var/log/cni.log
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make test` and `make lint`
6. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [Skupper](https://skupper.io/) - Multicloud communication for Kubernetes
- [CNI](https://github.com/containernetworking/cni) - Container Network Interface specification
- [CNI Plugins](https://github.com/containernetworking/plugins) - Standard CNI plugins
- [Whereabouts](https://github.com/k8snetworkplumbingwg/whereabouts) - IP Address Management (IPAM) CNI plugin

## Support

For issues, questions, or contributions, please visit:
- GitHub Issues: https://github.com/skupperproject/skupper-cni/issues
- Skupper Community: https://skupper.io/community/

## Acknowledgments

This plugin is built using:
- [containernetworking/cni](https://github.com/containernetworking/cni) - CNI specification and libraries
- [containernetworking/plugins](https://github.com/containernetworking/plugins) - CNI plugin utilities
- [vishvananda/netlink](https://github.com/vishvananda/netlink) - Linux netlink library for Go
