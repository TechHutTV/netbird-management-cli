# Setup Keys

[Home](../README.md) | [Getting Started](getting-started.md) | [Peers](peers.md) | **Setup Keys** | [Users](users.md) | [Tokens](tokens.md) | [More...](#documentation)

---

Manage device registration and onboarding keys. Setup keys are used to register new peers to your NetBird network. Running `netbird-manage setup-key` by itself will display the help menu.

## Query Operations

```bash
# List all setup keys
netbird-manage setup-key --list

# Filter by name (supports wildcards)
netbird-manage setup-key --list --filter-name "office-*"

# Filter by type (one-off or reusable)
netbird-manage setup-key --list --filter-type reusable

# Show only valid (non-revoked, non-expired) keys
netbird-manage setup-key --list --valid-only

# Inspect a specific setup key
netbird-manage setup-key --inspect <key-id>
```

## Create Operations

```bash
# Quick create a one-off key (7d expiration, single use)
netbird-manage setup-key --quick "office-laptop"

# Create a one-off key with custom settings
netbird-manage setup-key --create "temp-access" \
  --type one-off \
  --expires-in 1d \
  --usage-limit 1

# Create a reusable key for team onboarding
netbird-manage setup-key --create "team-onboarding" \
  --type reusable \
  --expires-in 30d \
  --usage-limit 10 \
  --auto-groups "group-id-1,group-id-2"

# Create an ephemeral peer key (peer deleted when disconnected)
netbird-manage setup-key --create "ephemeral-test" \
  --expires-in 7d \
  --ephemeral
```

## Update Operations

```bash
# Revoke a setup key (prevent new device registrations)
netbird-manage setup-key --revoke <key-id>

# Enable a previously revoked key
netbird-manage setup-key --enable <key-id>

# Update auto-groups for a key
netbird-manage setup-key --update-groups <key-id> \
  --groups "new-group-1,new-group-2"
```

## Delete Operations

```bash
# Delete a setup key
netbird-manage setup-key --delete <key-id>

# Delete multiple setup keys at once
netbird-manage setup-key --delete-batch <key-id-1,key-id-2,key-id-3>

# Delete all setup keys (with confirmation)
netbird-manage setup-key --delete-all
```

## Examples

```bash
# Create a quick one-off key for a new office computer
netbird-manage setup-key --quick "johns-laptop"

# Create a reusable key that expires in 90 days for contractor onboarding
netbird-manage setup-key --create "contractor-key" \
  --type reusable \
  --expires-in 90d \
  --usage-limit 5 \
  --auto-groups "contractors,limited-access"

# List all valid keys
netbird-manage setup-key --list --valid-only

# Revoke a compromised key immediately
netbird-manage setup-key --revoke 12345

# Inspect a key to check usage statistics
netbird-manage setup-key --inspect 12345
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--type` | `one-off` (single use) or `reusable` (multiple uses) | one-off |
| `--expires-in` | Human-readable duration: `1d`, `7d`, `30d`, `90d`, `1y` | 7d |
| `--usage-limit` | Maximum number of uses, `0` = unlimited | 0 |
| `--auto-groups` | Comma-separated group IDs for automatic peer assignment | - |
| `--ephemeral` | Mark peers registered with this key as ephemeral | false |
| `--allow-extra-dns-labels` | Allow additional DNS labels for registered peers | false |

## Notes

- Setup keys are displayed **only once** during creation - save them immediately!
- Expiration must be between 1 day and 1 year (API constraint)
- Revoked keys cannot be used to register new devices but existing devices remain active
- One-off keys are automatically revoked after first use

---

## Documentation

| Section | Description |
|---------|-------------|
| [Getting Started](getting-started.md) | Installation, safety features, debug mode |
| [Peers](peers.md) | Manage network peers |
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

[Home](../README.md) | [Peers](peers.md) | **Setup Keys** | [Users](users.md) | [Tokens](tokens.md) | [Groups](groups.md)
