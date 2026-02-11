# Identity Providers

[Home](../README.md) | [Getting Started](getting-started.md) | [DNS Zones](dns-zones.md) | **Identity Providers** | [Instance](instance.md) | [More...](#documentation)

---

Manage identity provider (IdP) configurations for your NetBird account. Useful for self-hosted instances that need to configure OIDC/OAuth2 authentication. Running `netbird-manage idp` by itself will display the help menu.

## Query Operations

```bash
# List all identity providers
netbird-manage idp --list

# Inspect a specific provider
netbird-manage idp --inspect <idp-id>

# JSON output
netbird-manage idp --list --output json
```

## Create Operations

```bash
# Create an OIDC identity provider
netbird-manage idp --create "My SSO" \
  --type oidc \
  --issuer "https://auth.example.com" \
  --client-id "your-client-id" \
  --client-secret "your-client-secret"

# Create a Google identity provider
netbird-manage idp --create "Google SSO" \
  --type google \
  --issuer "https://accounts.google.com" \
  --client-id "123456.apps.googleusercontent.com" \
  --client-secret "GOCSPX-..."

# Create an Entra (Azure AD) identity provider
netbird-manage idp --create "Azure AD" \
  --type entra \
  --issuer "https://login.microsoftonline.com/tenant-id/v2.0" \
  --client-id "app-id" \
  --client-secret "app-secret"
```

## Update Operations

```bash
# Update provider name
netbird-manage idp --update <idp-id> --name "Updated SSO"

# Update client credentials
netbird-manage idp --update <idp-id> \
  --client-id "new-client-id" \
  --client-secret "new-client-secret"

# Change issuer URL
netbird-manage idp --update <idp-id> \
  --issuer "https://new-auth.example.com"
```

## Delete Operations

```bash
# Delete an identity provider (with confirmation)
netbird-manage idp --delete <idp-id>
```

## Supported Provider Types

| Type | Description |
|------|-------------|
| `oidc` | Generic OpenID Connect provider |
| `zitadel` | Zitadel identity platform |
| `entra` | Microsoft Entra ID (Azure AD) |
| `google` | Google Workspace / Cloud Identity |
| `okta` | Okta identity provider |
| `pocketid` | PocketID identity provider |
| `microsoft` | Microsoft identity platform |

## Notes

- Identity provider management is primarily useful for self-hosted NetBird instances
- The `client_secret` is write-only and will not be returned in API responses
- Ensure the issuer URL matches your IdP configuration exactly
- Only one identity provider can typically be active at a time

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
| [Instance](instance.md) | Self-hosted instance setup |
| [Jobs](jobs.md) | Peer bundle collection and debugging |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [DNS Zones](dns-zones.md) | **Identity Providers** | [Instance](instance.md) | [Jobs](jobs.md)
