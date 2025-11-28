# Groups

[Home](../README.md) | [Getting Started](getting-started.md) | [Peers](peers.md) | [Tokens](tokens.md) | **Groups** | [Networks](networks.md) | [More...](#documentation)

---

Manage peer groups. Running `netbird-manage group` by itself will display the help menu.

## Query Operations

```bash
netbird-manage group --list                    # List all groups in your network
  --filter-name <pattern>                      # Filter by name (supports wildcards: prod-*)

netbird-manage group --inspect <group-id>      # View detailed information for a specific group
```

## Modification Operations

```bash
netbird-manage group --create <group-name>     # Create a new group
  --peers <id1,id2,...>                        # (Optional) Add peers on creation

netbird-manage group --delete <group-id>       # Delete a group
netbird-manage group --delete-batch <id1,id2,...>  # Delete multiple groups (comma-separated IDs)
netbird-manage group --delete-unused           # Delete all unused groups (no peers, resources, or references)

netbird-manage group --rename <group-id>       # Rename a group
  --new-name <new-name>                        # New name for the group

netbird-manage group --add-peers <group-id>    # Add multiple peers to a group
  --peers <id1,id2,...>                        # Comma-separated peer IDs

netbird-manage group --remove-peers <group-id> # Remove multiple peers from a group
  --peers <id1,id2,...>                        # Comma-separated peer IDs
```

## Examples

```bash
# Create a new group
netbird-manage group --create "Production-Servers"

# Create a group with initial peers
netbird-manage group --create "Dev-Team" --peers "peer-id-1,peer-id-2,peer-id-3"

# List all groups containing "prod" in the name
netbird-manage group --list --filter-name "prod*"

# Inspect a specific group
netbird-manage group --inspect d2l17grl0ubs73bh4vpg

# Rename a group
netbird-manage group --rename d2l17grl0ubs73bh4vpg --new-name "Production"

# Add multiple peers to a group at once
netbird-manage group --add-peers d2l17grl0ubs73bh4vpg --peers "peer1,peer2,peer3"

# Remove peers from a group
netbird-manage group --remove-peers d2l17grl0ubs73bh4vpg --peers "peer1,peer2"

# Delete a group
netbird-manage group --delete d2l17grl0ubs73bh4vpg

# Delete all unused groups (scans for groups with no peers, resources, or references)
netbird-manage group --delete-unused
```

---

## Documentation

| Section | Description |
|---------|-------------|
| [Getting Started](getting-started.md) | Installation, safety features, debug mode |
| [Peers](peers.md) | Manage network peers |
| [Setup Keys](setup-keys.md) | Device registration and onboarding keys |
| [Users](users.md) | User management and invitations |
| [Tokens](tokens.md) | Personal access token management |
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

[Home](../README.md) | [Tokens](tokens.md) | **Groups** | [Networks](networks.md) | [Policies](policies.md) | [Routes](routes.md)
