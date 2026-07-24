# Jobs

[Home](../README.md) | [Getting Started](getting-started.md) | [Posture Checks](posture-checks.md) | [Events](events.md) | **Jobs** | [Notifications](notifications.md) | [More...](#documentation)

---

Manage asynchronous peer jobs, such as collecting debug bundles from a peer. All job operations require `--peer <peer-id>`. Running `netbird-manage job` by itself will display the help menu.

## Query Operations

```bash
# List jobs for a peer
netbird-manage job --list --peer <peer-id>

# Inspect a specific job
netbird-manage job --inspect <job-id> --peer <peer-id>

# Export to JSON
netbird-manage job --list --peer <peer-id> --output json
```

## Create Operations

```bash
# Create a debug bundle job for a peer
netbird-manage job --create --peer <peer-id> --type bundle

# Collect the bundle over a period of time (in seconds)
netbird-manage job --create --peer <peer-id> \
  --type bundle \
  --bundle-for-time 60

# Include rotated log files and anonymize the bundle
netbird-manage job --create --peer <peer-id> \
  --type bundle \
  --log-file-count 3 \
  --anonymize
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--peer` | Peer ID (required for all operations) | - |
| `--type` | Job workload type (currently only `bundle`) | bundle |
| `--bundle-for-time` | Collect the bundle over this many seconds | 0 |
| `--log-file-count` | Number of rotated log files to include | 0 |
| `--anonymize` | Anonymize the collected bundle | false |

## Examples

```bash
# Collect a quick debug bundle from a peer
netbird-manage job --create --peer d3mjakrl0ubs738ajj00

# Collect an anonymized 2-minute bundle with extra logs
netbird-manage job --create --peer d3mjakrl0ubs738ajj00 \
  --bundle-for-time 120 \
  --log-file-count 5 \
  --anonymize

# Check job status
netbird-manage job --list --peer d3mjakrl0ubs738ajj00
```

## Notes

- Jobs run asynchronously on the peer; use `--list` or `--inspect` to check their status
- The job list shows ID, type, status, created/completed timestamps, and who triggered the job
- `bundle` is currently the only supported workload type
- Use `--anonymize` to strip sensitive data (IPs, hostnames) from the collected bundle

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
| [Notifications](notifications.md) | Notification channels (email/webhook) |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Events](events.md) | **Jobs** | [Notifications](notifications.md) | [Geo-Locations](geo-locations.md) | [Accounts](accounts.md)
