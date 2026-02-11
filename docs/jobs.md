# Jobs

[Home](../README.md) | [Getting Started](getting-started.md) | [Instance](instance.md) | **Jobs** | [Export & Import](export-import.md) | [More...](#documentation)

---

Manage peer jobs for debugging and bundle collection. Jobs are peer-scoped and used to collect diagnostic bundles from peers. Running `netbird-manage job` by itself will display the help menu.

## Query Operations

```bash
# List all jobs for a peer
netbird-manage job --list <peer-id>

# Inspect a specific job
netbird-manage job --inspect --peer-id <peer-id> --job-id <job-id>

# JSON output
netbird-manage job --list <peer-id> --output json
```

## Create Operations

```bash
# Create a basic bundle collection job
netbird-manage job --create <peer-id>

# Create a job with time duration
netbird-manage job --create <peer-id> --bundle-for-time "300"

# Create a job with log file collection
netbird-manage job --create <peer-id> \
  --log-file-count 10 \
  --anonymize

# Create a job with all options
netbird-manage job --create <peer-id> \
  --bundle-for-time "600" \
  --log-file-count 5 \
  --anonymize
```

## Job Statuses

| Status | Description |
|--------|-------------|
| `pending` | Job created, waiting for peer to pick it up |
| `succeeded` | Job completed successfully |
| `failed` | Job failed (check `failed_reason` for details) |

## Examples

```bash
# List all jobs for a specific peer
netbird-manage job --list d3mjakrl0ubs738ajj00

# Create a diagnostic bundle job
netbird-manage job --create d3mjakrl0ubs738ajj00 --log-file-count 5

# Check job status
netbird-manage job --inspect \
  --peer-id d3mjakrl0ubs738ajj00 \
  --job-id job-abc123
```

**Example inspect output:**
```
Job: job-abc123
---------------------------------
  Status:       succeeded
  Created At:   2025-01-15T10:30:00Z
  Completed At: 2025-01-15T10:31:45Z
  Triggered By: user@example.com
  Workload Type: bundle
  Parameters:
    log_file_count: 5
    anonymize: true
  Result:
    upload_key: bundle-xyz789
```

## Notes

- Jobs are scoped to individual peers
- The primary job type is `bundle` for collecting diagnostic information
- The `--anonymize` flag removes sensitive data from collected bundles
- Job results may include an `upload_key` for retrieving the bundle

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
| [Instance](instance.md) | Self-hosted instance setup |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Instance](instance.md) | **Jobs** | [Export & Import](export-import.md) | [Migrate](migrate.md)
