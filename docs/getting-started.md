# Getting Started

[Home](../README.md) | **Getting Started** | [Peers](peers.md) | [Groups](groups.md) | [Networks](networks.md) | [Policies](policies.md) | [More...](#documentation)

---

## Installation

You must have the [Go toolchain](https://go.dev/doc/install) (version 1.18 or later) installed on your system. Clone this repository and build the binary:

```bash
git clone https://github.com/TechHutTV/netbird-management-cli.git
cd netbird-management-cli
go build ./cmd/netbird-manage
```

This creates a `netbird-manage` binary in the current directory. Optionally, move it to a directory in your PATH:

```bash
sudo mv netbird-manage /usr/local/bin/
```

## Safety Features

### Confirmation Prompts

To prevent accidental data loss, all destructive operations (delete, remove, revoke) now require explicit confirmation before executing. When you attempt to delete a resource, you'll see detailed information about what will be removed:

```bash
$ netbird-manage peer --remove abc123

About to remove peer:
  Name:      laptop-001
  ID:        abc123
  IP:        100.64.0.5
  Hostname:  laptop-001.local
  OS:        Linux
  Connected: true
  Groups:    2 (developers, ssh-access)

⚠️  This action cannot be undone. Continue? [y/N]: _
```

**For bulk operations** (like `--delete-unused` or `--delete-all`), you'll need to type a confirmation phrase:

```bash
$ netbird-manage group --delete-unused

🔴 This will delete 3 groups:
  - old-servers (ID: def456)
  - test-group (ID: ghi789)
  - unused-group (ID: jkl012)

Type 'delete 3 groups' to confirm: _
```

### Automation Mode

For scripts and automation, use the `--yes` (or `-y`) flag to skip all confirmation prompts:

```bash
# Skip confirmation for automation
netbird-manage --yes peer --remove abc123

# Also works with short flag
netbird-manage -y group --delete def456
```

**Warning:** When using `--yes`, deletions happen immediately without any prompts. Use with caution!

## Debug Mode

Enable verbose debug output to see all HTTP requests and responses. This is invaluable for troubleshooting API issues or understanding what's happening under the hood:

```bash
# Enable debug mode with --debug or -d flag
netbird-manage --debug peer --list

# Debug output shows:
═══ DEBUG: HTTP REQUEST ═══
GET https://api.netbird.io/api/peers
Headers:
  Authorization: Token [REDACTED]
  Accept: application/json

═══ DEBUG: HTTP RESPONSE ═══
Status: 200 OK
Headers:
  Content-Type: application/json
Response Body:
[
  {
    "id": "abc123",
    "name": "laptop-001",
    ...
  }
]
═══════════════════════════
```

**Features:**
- Shows full HTTP method and URL
- Displays request headers (token redacted for security)
- Pretty-prints JSON request/response bodies
- All debug output goes to stderr (keeps stdout clean for scripting)

## Batch Operations

Process multiple resources at once for efficient bulk operations. All batch operations support the same confirmation prompts as single deletions:

```bash
# Remove multiple peers at once
netbird-manage peer --remove-batch abc123,def456,ghi789

# Delete multiple groups
netbird-manage group --delete-batch dev-team,test-group,old-servers

# Delete multiple setup keys
netbird-manage setup-key --delete-batch key1,key2,key3

# Combine with --yes for automation
netbird-manage --yes peer --remove-batch abc123,def456,ghi789
```

**Batch operation features:**
- Fetches and displays details for all resources before confirmation
- Shows progress indicator during processing (e.g., `[2/5] Removing peer...`)
- Continues processing even if some operations fail
- Provides summary at the end: `Completed: 4 succeeded, 1 failed`
- Supports type-to-confirm for safety (type `delete N resources` to proceed)

**Example batch removal:**

```bash
$ netbird-manage group --delete-batch old-servers,test-group,unused

Fetching group details...
🔴 This will delete 3 groups:
  - old-servers (ID: abc123, Peers: 0, Resources: 0)
  - test-group (ID: def456, Peers: 2, Resources: 1)
  - unused (ID: ghi789, Peers: 0, Resources: 0)

Type 'delete 3 groups' to confirm: delete 3 groups

[1/3] Deleting group 'old-servers'... ✓ Done
[2/3] Deleting group 'test-group'... ✓ Done
[3/3] Deleting group 'unused'... ✓ Done

✓ All 3 groups deleted successfully
```

---

## Documentation

| Section | Description |
|---------|-------------|
| [Peers](peers.md) | Manage network peers |
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

[Home](../README.md) | **Getting Started** | [Peers](peers.md) | [Groups](groups.md) | [Networks](networks.md) | [Policies](policies.md)
