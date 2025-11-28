# Users

[Home](../README.md) | [Getting Started](getting-started.md) | [Setup Keys](setup-keys.md) | **Users** | [Tokens](tokens.md) | [Groups](groups.md) | [More...](#documentation)

---

Manage users and user invitations. Running `netbird-manage user` by itself will display the help menu.

## Query Operations

```bash
# List all users
netbird-manage user --list

# List only service users
netbird-manage user --list --service-users

# List only regular users
netbird-manage user --list --regular-users

# Get current authenticated user information
netbird-manage user --me
# Note: --me is not available for service user tokens
```

## Invite/Create Operations

```bash
# Invite a new regular user
netbird-manage user --invite --email "user@example.com"

# Invite a user with specific role
netbird-manage user --invite \
  --email "admin@example.com" \
  --name "John Admin" \
  --role admin

# Invite a user with auto-groups
netbird-manage user --invite \
  --email "developer@example.com" \
  --role user \
  --auto-groups "group-id-1,group-id-2"

# Create a service user
netbird-manage user --invite \
  --email "ci-bot@example.com" \
  --role user \
  --service-user
```

## Update Operations

```bash
# Update user role
netbird-manage user --update <user-id> --role admin

# Update user auto-groups
netbird-manage user --update <user-id> --auto-groups "group-1,group-2"

# Block a user
netbird-manage user --update <user-id> --blocked

# Unblock a user
netbird-manage user --update <user-id> --unblocked

# Update role and block user
netbird-manage user --update <user-id> --role user --blocked
```

## Delete Operations

```bash
# Remove a user
netbird-manage user --remove <user-id>

# Resend invitation to a user
netbird-manage user --resend-invite <user-id>
```

## User Roles

| Role | Description |
|------|-------------|
| `admin` | Full administrative access |
| `user` | Standard user access |
| `owner` | Account owner (highest privileges) |

## Notes

- Service users are designed for API access and automation
- Blocked users cannot access the system but their configuration is preserved
- Auto-groups automatically assign new peers to specified groups

---

## Documentation

| Section | Description |
|---------|-------------|
| [Getting Started](getting-started.md) | Installation, safety features, debug mode |
| [Peers](peers.md) | Manage network peers |
| [Setup Keys](setup-keys.md) | Device registration and onboarding keys |
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

[Home](../README.md) | [Setup Keys](setup-keys.md) | **Users** | [Tokens](tokens.md) | [Groups](groups.md) | [Networks](networks.md)
