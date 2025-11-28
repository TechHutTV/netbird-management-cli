# Peers

[Home](../README.md) | [Getting Started](getting-started.md) | **Peers** | [Groups](groups.md) | [Networks](networks.md) | [Policies](policies.md) | [More...](#documentation)

---

Manage network peers. Running `netbird-manage peer` by itself will display the help menu.

## Query Operations

```bash
netbird-manage peer --list                     # List all peers in your network
  --filter-name <pattern>                      # Filter by name (supports wildcards: ubuntu*)
  --filter-ip <pattern>                        # Filter by IP address pattern

netbird-manage peer --inspect <peer-id>        # View detailed information for a single peer

netbird-manage peer --accessible-peers <peer-id>  # List peers accessible from the specified peer
```

## Modification Operations

```bash
netbird-manage peer --remove <peer-id>         # Remove a peer from your network
netbird-manage peer --remove-batch <id1,id2,...>  # Remove multiple peers (comma-separated IDs)

netbird-manage peer --edit <peer-id>           # Edit peer group membership
  --add-group <group-id>                       # Add peer to a specified group
  --remove-group <group-id>                    # Remove peer from a specified group

netbird-manage peer --update <peer-id>         # Update peer settings
  --rename <new-name>                          # Change peer name
  --ssh-enabled <true|false>                   # Enable/disable SSH access
  --login-expiration <true|false>              # Enable/disable login expiration
  --inactivity-expiration <true|false>         # Enable/disable inactivity expiration
  --approval-required <true|false>             # Require approval (cloud-only)
  --ip <ip-address>                            # Set IP (must be in 100.64.0.0/10 range)
```

## Examples

```bash
# List all peers with "ubuntu" in the name
netbird-manage peer --list --filter-name "ubuntu*"

# Rename a peer
netbird-manage peer --update d3mjakrl0ubs738ajj00 --rename "UbuntuServer"

# Enable SSH and disable login expiration
netbird-manage peer --update d3mjakrl0ubs738ajj00 --ssh-enabled true --login-expiration false

# Set a custom IP address
netbird-manage peer --update d3mjakrl0ubs738ajj00 --ip 100.64.1.50

# Check which peers a specific peer can access
netbird-manage peer --accessible-peers d3mjakrl0ubs738ajj00

# Remove multiple peers at once
netbird-manage peer --remove-batch abc123,def456,ghi789
```

---

## Documentation

| Section | Description |
|---------|-------------|
| [Getting Started](getting-started.md) | Installation, safety features, debug mode |
| [Setup Keys](setup-keys.md) | Device registration and onboarding keys |
| [Users](users.md) | User management and invitations |
| [Tokens](tokens.md) | Personal access token management |
| [Groups](groups.md) | Peer group management |
| [Networks](networks.md) | Networks, resources, and routers |
| [Policies](policies.md) | Access control policies and firewall rules |
| [Routes](routes.md) | Network routing configuration |
| [DNS](dns.md) | DNS nameserver groups and settings |
| [Posture Checks](posture-checks.md) | Device compliance validation |
| [Events](events.md) | Audit logs and traffic monitoring |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Getting Started](getting-started.md) | **Peers** | [Groups](groups.md) | [Networks](networks.md) | [Policies](policies.md)
