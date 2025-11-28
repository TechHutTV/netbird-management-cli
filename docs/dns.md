# DNS

[Home](../README.md) | [Getting Started](getting-started.md) | [Policies](policies.md) | [Routes](routes.md) | **DNS** | [Posture Checks](posture-checks.md) | [More...](#documentation)

---

Manage DNS nameserver groups and settings. DNS groups control domain resolution for specific peer groups. Running `netbird-manage dns` by itself will display the help menu.

## Query Operations

```bash
# List all DNS nameserver groups
netbird-manage dns --list

# Filter by name pattern
netbird-manage dns --list --filter-name "corp-*"

# Show only primary DNS groups
netbird-manage dns --list --primary-only

# Show only enabled groups
netbird-manage dns --list --enabled-only

# Inspect a specific DNS group
netbird-manage dns --inspect <group-id>

# Get DNS settings for the account
netbird-manage dns --get-settings
```

## Modification Operations

```bash
# Create a DNS group with Google and Cloudflare DNS
netbird-manage dns --create "corp-dns" \
  --nameservers "8.8.8.8:53,1.1.1.1:53" \
  --groups <group-id>

# Create a DNS group with domain matching
netbird-manage dns --create "internal-dns" \
  --nameservers "10.0.0.53:53" \
  --groups <group-id> \
  --domains "example.com,internal.local" \
  --search-domains \
  --primary

# Create DNS group with description
netbird-manage dns --create "public-dns" \
  --nameservers "1.1.1.1:53,8.8.8.8:53" \
  --groups <group-id> \
  --description "Public DNS resolvers for external access"

# Update nameservers for a group
netbird-manage dns --update <group-id> \
  --nameservers "9.9.9.9:53,149.112.112.112:53"

# Set a group as primary
netbird-manage dns --update <group-id> --primary

# Enable/disable a DNS group
netbird-manage dns --enable <group-id>
netbird-manage dns --disable <group-id>

# Update DNS settings (disable management for specific groups)
netbird-manage dns --update-settings \
  --disabled-groups <group-id-1>,<group-id-2>

# Delete a DNS group
netbird-manage dns --delete <group-id>
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--nameservers` | DNS servers in `IP:port` format | - |
| `--groups` | Target peer group IDs (required, comma-separated) | - |
| `--domains` | Match specific domains (optional, comma-separated) | - |
| `--search-domains` | Enable search domains | false |
| `--primary` | Set as primary DNS group | false |
| `--description` | DNS group description | - |

## Notes

- Nameserver format: `8.8.8.8:53` or just `8.8.8.8` (defaults to port 53)
- Primary DNS group is used when no domain-specific match is found
- Search domains append the domain to short hostnames
- Only one primary DNS group should be active at a time

---

## Documentation

| Section | Description |
|---------|-------------|
| [Getting Started](getting-started.md) | Installation, safety features, debug mode |
| [Peers](peers.md) | Manage network peers |
| [Setup Keys](setup-keys.md) | Device registration and onboarding keys |
| [Users](users.md) | User management and invitations |
| [Tokens](tokens.md) | Personal access token management |
| [Groups](groups.md) | Peer group management |
| [Networks](networks.md) | Networks, resources, and routers |
| [Policies](policies.md) | Access control policies and firewall rules |
| [Routes](routes.md) | Network routing configuration |
| [Posture Checks](posture-checks.md) | Device compliance validation |
| [Events](events.md) | Audit logs and traffic monitoring |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Routes](routes.md) | **DNS** | [Posture Checks](posture-checks.md) | [Events](events.md) | [Accounts](accounts.md)
