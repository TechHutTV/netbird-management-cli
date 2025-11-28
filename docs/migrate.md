# Migrate

[Home](../README.md) | [Getting Started](getting-started.md) | [Accounts](accounts.md) | [Export & Import](export-import.md) | **Migrate**

---

Migrate peers and/or complete configuration between NetBird accounts. Running `netbird-manage migrate` by itself will display the help menu.

## Use Cases

- **Full Account Migration**: Move everything from one account to another with `--all`
- **Configuration Replication**: Copy network configuration between accounts with `--config`
- **Cloud to Self-Hosted Migration**: Move peers and config from NetBird Cloud to self-hosted
- **Self-Hosted to Cloud**: Move from self-hosted to NetBird Cloud
- **Account Consolidation**: Merge multiple NetBird accounts into one
- **Configuration Backup & Restore**: Migrate configuration between environments

## Configuration Migration

Migrate groups, policies, networks, routes, DNS, posture checks, and setup keys between accounts.

```bash
# Preview configuration migration (dry-run, recommended first)
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --config --dry-run

# Migrate all configuration
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --config

# Migrate configuration, skip existing resources
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --config --skip-existing

# Migrate configuration and update existing resources
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --config --update
```

## Selective Configuration Migration

Migrate only specific resource types:

```bash
# Migrate only groups
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --groups

# Migrate only policies (requires groups to exist in destination)
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --policies --skip-existing

# Migrate groups and policies together
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --groups --policies

# Migrate network configuration (routes, DNS, networks)
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --routes --dns --networks --skip-existing
```

## Full Migration (Configuration + Peers)

Migrate everything including configuration and generate peer migration commands:

```bash
# Migrate everything (config + all peers)
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --all

# Migrate everything with verbose output
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --all --verbose
```

## Single Peer Migration

```bash
# Migrate a single peer between cloud accounts
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --peer "abc123def"

# Migrate from cloud to self-hosted
netbird-manage migrate \
  --source-token "nbp_cloud..." \
  --source-url "https://api.netbird.io/api" \
  --dest-token "nbp_selfhost..." \
  --dest-url "https://netbird.mycompany.com/api" \
  --peer "abc123def"

# Migrate from self-hosted to cloud
netbird-manage migrate \
  --source-token "nbp_selfhost..." \
  --source-url "https://netbird.mycompany.com/api" \
  --dest-token "nbp_cloud..." \
  --peer "abc123def"
```

## Batch Peer Migration (by Group)

```bash
# Migrate all peers in a group
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --group "production-servers"

# Migrate group with custom key expiry
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --group "dev-team" \
  --key-expiry "7d"
```

## Configuration Migration Example Output

```
$ netbird-manage migrate \
    --source-token "nbp_Sxxxx..." \
    --dest-token "nbp_Dxxxx..." \
    --config --dry-run --verbose

Configuration Migration Preview (Dry Run)
==========================================
  Source: https://api.netbird.io/api
  Destination: https://api.netbird.io/api

Fetching current state...
  Source: 5 groups, 3 policies, 2 networks, 4 routes, 2 DNS, 1 posture checks, 3 setup keys, 12 peers
  Destination: 2 groups, 1 policies, 0 networks, 0 DNS, 0 posture checks, 0 setup keys, 0 peers

Groups:
  CREATE   developers (would create)
  CREATE   production-servers (would create)
  SKIP     All (already exists)

Posture Checks:
  CREATE   min-version-check (would create)

Policies:
  CREATE   dev-to-prod (would create)
  CREATE   ssh-access (would create)
  SKIP     Default (already exists)

...
```

## Peer Migration Example Output

```
$ netbird-manage migrate \
    --source-token "nbp_Sxxxx..." \
    --dest-token "nbp_Dxxxx..." \
    --dest-url "https://netbird.mycompany.com/api" \
    --peer "abc123def"

Fetching peer from source account...
  Source: https://api.netbird.io/api

Source Peer Details:
  Name:       laptop-dev-01
  Hostname:   brandon-laptop
  ID:         abc123def456
  IP:         100.64.0.15
  OS:         Linux (Ubuntu 22.04)
  Version:    0.28.4
  Groups:     developers, ssh-access, monitoring

Connecting to destination account...
  Destination: https://netbird.mycompany.com/api

Creating setup key in destination...
  Key Name:   migrate-laptop-dev-01-20251123
  Type:       one-off
  Auto-Groups: developers, ssh-access, monitoring
  Groups created in destination: monitoring

Setup key created successfully.

========================================================================
MIGRATION COMMAND - Run this on the peer device:
========================================================================

  sudo netbird down && sudo netbird up \
    --setup-key XXXX-XXXX-XXXX-XXXX \
    --hostname brandon-laptop \
    --management-url https://netbird.mycompany.com

========================================================================

Notes:
  - The peer will disconnect from the source network
  - A new peer ID and IP will be assigned in the destination
  - The setup key expires in 1 day(s) and is single-use
  - Old peer entry in source account must be manually removed

To remove the old peer from source after migration:
  netbird-manage peer --remove abc123def456
```

## Flags Reference

### Required Flags

| Flag | Description |
|------|-------------|
| `--source-token` | API token for the source (exporting) account |
| `--dest-token` | API token for the destination (importing) account |

### Migration Type Flags

| Flag | Description |
|------|-------------|
| `--config` | Migrate all configuration (groups, policies, networks, routes, DNS, posture checks, setup keys) |
| `--all` | Migrate everything (configuration + generate peer migration commands) |
| `--peer <id>` | Migrate a single peer by ID |
| `--group <name>` | Migrate all peers in a group |

### Selective Configuration Flags

| Flag | Description |
|------|-------------|
| `--groups` | Migrate only groups |
| `--policies` | Migrate only policies |
| `--networks` | Migrate only networks |
| `--routes` | Migrate only routes |
| `--dns` | Migrate only DNS nameserver groups |
| `--posture-checks` | Migrate only posture checks |
| `--setup-keys` | Migrate only setup keys |

### Configuration Migration Options

| Flag | Default | Description |
|------|---------|-------------|
| `--skip-existing` | `false` | Skip resources that already exist in destination |
| `--update` | `false` | Update existing resources in destination |
| `--dry-run` | `false` | Preview changes without applying them |
| `--verbose` | `false` | Show detailed output |

### Peer Migration Options

| Flag | Default | Description |
|------|---------|-------------|
| `--source-url` | `https://api.netbird.io/api` | Management URL for source |
| `--dest-url` | `https://api.netbird.io/api` | Management URL for destination |
| `--create-groups` | `true` | Create missing groups in destination |
| `--key-expiry` | `24h` | Setup key expiration (e.g., 1h, 24h, 7d) |

## What Gets Migrated

### Configuration Migration (`--config`)

- Groups (created empty - without peers)
- Policies (with all rules, group references resolved)
- Networks (with resources and routers where possible)
- Routes (peer group routes only - routes referencing specific peers are skipped)
- DNS nameserver groups
- Posture checks (all 5 check types)
- Setup keys (with resolved auto-groups)

### Peer Migration (`--peer` or `--group`)

- Hostname (preserved via `--hostname` flag)
- Group memberships (groups created in destination if missing)
- **Not migrated**: Peer ID, IP address, connection history/statistics (new ones assigned)

## Recommended Migration Order

For a complete migration:

```bash
# Step 1: Preview configuration migration
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --config --dry-run

# Step 2: Apply configuration migration
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --config --skip-existing

# Step 3: Migrate peers (generates commands)
netbird-manage migrate \
  --source-token "nbp_source..." \
  --dest-token "nbp_dest..." \
  --group "all-servers"

# Step 4: Run the generated commands on each peer device

# Step 5: Clean up old peers from source account
netbird-manage --yes peer --remove-batch <old-peer-ids>
```

## Cleaning Up Old Configuration

When migrating to a different management server, you may need to remove existing NetBird configuration files on each peer:

**Linux:**
```bash
sudo netbird down
sudo rm -rf /etc/netbird/
sudo rm -rf /var/lib/netbird/
# Then run the migration command
```

**macOS:**
```bash
sudo netbird down
sudo rm -rf /etc/netbird/
sudo rm -rf /var/db/netbird/
# Then run the migration command
```

**Windows (Run as Administrator):**
```powershell
netbird down
Remove-Item -Recurse -Force "C:\ProgramData\Netbird"
# Then run the migration command
```

## Important Notes

- **Configuration Migration**: Resources are migrated in dependency order (groups → posture checks → policies → routes → DNS → networks → setup keys)
- **Peer Dependencies**: Routes with specific peer routing are skipped. Migrate peers first, then re-run configuration migration with `--update`.
- **Groups are Empty**: When migrating groups via configuration, they are created without peers. Use peer migration to add peers.
- **Dry Run First**: Always use `--dry-run` to preview changes before applying
- **Skip Existing**: Use `--skip-existing` to safely re-run migrations after fixing errors

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
| [Posture Checks](posture-checks.md) | Device compliance validation |
| [Events](events.md) | Audit logs and traffic monitoring |
| [Geo-Locations](geo-locations.md) | Geographic location data |
| [Accounts](accounts.md) | Account settings and configuration |
| [Ingress Ports](ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](export-import.md) | YAML/JSON configuration management |

---

[Home](../README.md) | [Export & Import](export-import.md) | **Migrate**
