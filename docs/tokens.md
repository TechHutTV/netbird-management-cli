# Tokens

[Home](../README.md) | [Getting Started](getting-started.md) | [Users](users.md) | **Tokens** | [Groups](groups.md) | [Networks](networks.md) | [More...](#documentation)

---

Manage personal access tokens for API authentication. Running `netbird-manage token` by itself will display the help menu.

## Query Operations

```bash
# List all personal access tokens
netbird-manage token --list

# List tokens when using service user token (requires --user-id)
netbird-manage token --list --user-id <user-id>

# Inspect a specific token
netbird-manage token --inspect <token-id>

# Inspect token when using service user token (requires --user-id)
netbird-manage token --inspect <token-id> --user-id <user-id>
```

## Create Operations

```bash
# Create a token with default 90-day expiration
netbird-manage token --create --name "My CLI Token"

# Create a token with custom expiration
netbird-manage token --create \
  --name "CI/CD Token" \
  --expires-in 365

# Create a short-lived token
netbird-manage token --create \
  --name "Testing Token" \
  --expires-in 7
```

## Delete Operations

```bash
# Revoke/delete a token
netbird-manage token --revoke <token-id>
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `--expires-in` | Expiration in days (1-365) | 90 |
| `--user-id` | User ID for token operations | current user |

**Note:** `--user-id` is **required for service user tokens** since they cannot access `/users/current` endpoint. Use `netbird-manage user --list` to find your user ID.

## Examples

```bash
# Create a token for automation
netbird-manage token --create --name "Terraform Token" --expires-in 365

# List all tokens to check expiration dates
netbird-manage token --list

# List tokens when using service user token
netbird-manage token --list --user-id ef3799d6-3891-4769-9690-e6798258d5f6

# Revoke a compromised token
netbird-manage token --revoke tok-abc123xyz

# Check token details
netbird-manage token --inspect tok-abc123xyz

# Revoke token when using service user token
netbird-manage token --revoke tok-abc123xyz --user-id ef3799d6-3891-4769-9690-e6798258d5f6
```

## Important Notes

- **Token values are only shown once during creation** - save them immediately!
- Tokens are used for API authentication via `Authorization: Token <token>` header
- Revoked tokens cannot be recovered - you must create new ones
- Use tokens for CI/CD pipelines, automation, and programmatic access
- **Service user tokens**: Must provide `--user-id` flag for all token operations

---

## Documentation

| Section | Description |
|---------|-------------|
| [Getting Started](getting-started.md) | Installation, safety features, debug mode |
| [Peers](peers.md) | Manage network peers |
| [Setup Keys](setup-keys.md) | Device registration and onboarding keys |
| [Users](users.md) | User management and invitations |
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

[Home](../README.md) | [Users](users.md) | **Tokens** | [Groups](groups.md) | [Networks](networks.md) | [Policies](policies.md)
