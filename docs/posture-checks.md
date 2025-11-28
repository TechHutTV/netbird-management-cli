# Posture Checks

[Home](../README.md) | [Getting Started](getting-started.md) | [Routes](routes.md) | [DNS](dns.md) | **Posture Checks** | [Events](events.md) | [More...](#documentation)

---

Manage device posture checks for zero-trust security. Posture checks validate device compliance before granting network access. Running `netbird-manage posture-check` by itself will display the help menu.

## Query Operations

```bash
# List all posture checks
netbird-manage posture-check --list

# Filter by name pattern
netbird-manage posture-check --list --filter-name "version-*"

# Filter by check type
netbird-manage posture-check --list --filter-type "nb-version"

# Inspect a specific posture check
netbird-manage posture-check --inspect <check-id>
```

## Check Types & Creation

### NetBird Version Check

Ensure peers run a minimum NetBird version.

```bash
netbird-manage posture-check --create "min-nb-version" \
  --type nb-version \
  --min-version "0.28.0" \
  --description "Require NetBird 0.28.0 or newer"
```

### OS Version Check

Validate minimum OS or kernel versions.

```bash
# Require macOS 13.0 or newer
netbird-manage posture-check --create "macos-13+" \
  --type os-version \
  --os darwin \
  --min-os-version "13.0"

# Require Linux kernel 5.10 or newer
netbird-manage posture-check --create "linux-kernel-5.10+" \
  --type os-version \
  --os linux \
  --min-kernel "5.10"

# Require Android 12 or newer
netbird-manage posture-check --create "android-12+" \
  --type os-version \
  --os android \
  --min-os-version "12"

# Supported OS types: android, darwin, ios, linux, windows
```

### Geo-Location Check

Restrict or allow access based on geographic location.

```bash
# Allow access only from US
netbird-manage posture-check --create "us-only" \
  --type geo-location \
  --locations "US" \
  --action allow

# Allow specific cities
netbird-manage posture-check --create "office-locations" \
  --type geo-location \
  --locations "US:NewYork,GB:London,DE:Berlin" \
  --action allow

# Deny access from specific countries
netbird-manage posture-check --create "geo-block" \
  --type geo-location \
  --locations "RU,CN,KP" \
  --action deny
```

### Network Range Check

Require peers to be on specific networks.

```bash
# Allow only corporate network ranges
netbird-manage posture-check --create "corporate-net" \
  --type network-range \
  --ranges "192.168.0.0/16,10.0.0.0/8" \
  --action allow

# Deny access from public networks
netbird-manage posture-check --create "block-public" \
  --type network-range \
  --ranges "0.0.0.0/0" \
  --action deny
```

### Process Check

Verify required security software is running.

```bash
# Check for antivirus on multiple platforms
netbird-manage posture-check --create "av-required" \
  --type process \
  --linux-path "/usr/bin/clamav" \
  --mac-path "/Applications/Antivirus.app" \
  --windows-path "C:\\Program Files\\Antivirus\\av.exe"

# Check for VPN client
netbird-manage posture-check --create "vpn-client" \
  --type process \
  --windows-path "C:\\Program Files\\VPN\\client.exe"
```

## Modification Operations

```bash
# Update a posture check
netbird-manage posture-check --update <check-id> \
  --type nb-version \
  --min-version "0.29.0" \
  --description "Updated minimum version"

# Delete a posture check
netbird-manage posture-check --delete <check-id>
```

## Posture Check Types

| Type | Description |
|------|-------------|
| `nb-version` | NetBird version check |
| `os-version` | Operating system version check |
| `geo-location` | Geographic location check |
| `network-range` | Peer network range check |
| `process` | Running process check |

## Notes

- Posture checks are evaluated before granting network access
- Multiple checks can be combined for defense-in-depth security
- Location checks use ISO 3166-1 alpha-2 country codes (e.g., US, GB, DE)
- Process checks support platform-specific paths for cross-platform security

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
| [DNS](dns.md) | DNS nameserver groups and settings |
| [Events](events.md) | Audit logs and traffic monitoring |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [DNS](dns.md) | **Posture Checks** | [Events](events.md) | [Geo-Locations](geo-locations.md) | [Accounts](accounts.md)
