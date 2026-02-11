# Instance Management

[Home](../README.md) | [Getting Started](getting-started.md) | [Identity Providers](identity-providers.md) | **Instance** | [Jobs](jobs.md) | [More...](#documentation)

---

Manage self-hosted NetBird instance setup. These commands do **not** require authentication and are used for initial instance configuration. Running `netbird-manage instance` by itself will display the help menu.

## Check Instance Status

```bash
# Check if the instance requires setup
netbird-manage instance --status

# With custom management URL
netbird-manage instance --status --management-url https://your-server.com/api

# JSON output
netbird-manage instance --status --output json
```

**Example output:**
```
Instance Status:
  Setup Required: Yes
  Run 'netbird-manage instance --setup --email <email> --password <password> --name <name>' to set up
```

## Set Up Instance

```bash
# Create the initial admin user
netbird-manage instance --setup \
  --email "admin@example.com" \
  --password "secure_password_here" \
  --name "Admin User"

# With custom management URL
netbird-manage instance --setup \
  --email "admin@example.com" \
  --password "secure_password_here" \
  --name "Admin User" \
  --management-url https://your-server.com/api
```

**Example output:**
```
Instance set up successfully!
  User ID: abc123
  Email:   admin@example.com

You can now connect with:
  netbird-manage connect --token <your-token>
```

## Notes

- These commands do **not** require an API token or prior `connect` configuration
- Instance setup only works when no accounts exist (fresh installation)
- The embedded IdP must be enabled for setup to work
- Password must be at least 8 characters
- After setup, generate a token from the dashboard and use `netbird-manage connect` to authenticate

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
| [DNS Zones](dns-zones.md) | Custom DNS zones and records |
| [Posture Checks](posture-checks.md) | Device compliance validation |
| [Events](events.md) | Audit logs and traffic monitoring |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Identity Providers](identity-providers.md) | OIDC/OAuth2 identity provider management |
| [Jobs](jobs.md) | Peer bundle collection and debugging |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Identity Providers](identity-providers.md) | **Instance** | [Jobs](jobs.md) | [Export & Import](export-import.md)
