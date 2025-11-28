# Export & Import

[Home](../README.md) | [Getting Started](getting-started.md) | [Accounts](accounts.md) | [Ingress Ports](ingress-ports.md) | **Export & Import** | [Migrate](migrate.md)

---

Export and import NetBird configuration for GitOps workflows, backup, and infrastructure-as-code management.

## Export

Export your NetBird configuration to YAML or JSON files. Running `netbird-manage export` by itself will display the help menu.

### Export Operations

```bash
# Export to single YAML file (default)
netbird-manage export
netbird-manage export --full
netbird-manage export --format yaml

# Export to single JSON file
netbird-manage export --format json

# Export to single file in specific directory
netbird-manage export --full ./backups
netbird-manage export --format json ./backups

# Export to split files (one file per resource type)
netbird-manage export --split
netbird-manage export --split --format json

# Export to split files in specific directory
netbird-manage export --split ~/exports
netbird-manage export --split --format json ~/exports
```

### Output Formats

**Single File** (`netbird-manage-export-YYMMDD.{yml,json}`):
- All resources in one file
- Clean map-based structure
- Perfect for backups and small deployments

**Split Files** (`netbird-manage-export-YYMMDD/`):
```
├── config.{yml,json}          # Metadata and import order
├── groups.{yml,json}          # Group definitions with peer names
├── policies.{yml,json}        # Access control policies with rules
├── networks.{yml,json}        # Networks with resources and routers
├── routes.{yml,json}          # Custom network routes
├── dns.{yml,json}             # DNS nameserver groups
├── posture-checks.{yml,json}  # Device compliance checks
└── setup-keys.{yml,json}      # Device onboarding keys
```

### Exported Resources

- Groups (with peer names - **for reference only, not imported**)
- Policies (with rules and group references)
- Networks (with resources and routers)
- Routes (with routing configuration)
- DNS (nameserver groups and domains)
- Posture Checks (all 5 check types)
- Setup Keys (with auto-groups)

### Sample YAML Structure

```yaml
groups:
  developers:
    description: "Development team members"
    peers:
      - "alice-laptop"
      - "bob-workstation"

policies:
  allow-devs-to-prod:
    description: "SSH access for developers"
    enabled: true
    rules:
      ssh-access:
        action: "accept"
        protocol: "tcp"
        ports: ["22"]
        sources: ["developers"]
        destinations: ["production-servers"]
```

### Sample JSON Structure

```json
{
  "metadata": {
    "version": "1.0",
    "exported_at": "2025-01-15T10:30:00Z",
    "management_url": "https://api.netbird.io/api",
    "_important_note": "PEERS CANNOT BE IMPORTED - Use 'netbird-manage migrate' to migrate peers"
  },
  "groups": {
    "developers": {
      "description": "Group with 2 peers",
      "peers": ["alice-laptop", "bob-workstation"],
      "_peers_note": "These peers are for reference only and will NOT be imported"
    },
    "production-servers": {
      "description": "Group with 3 peers"
    }
  },
  "policies": {
    "allow-devs-to-prod": {
      "description": "SSH access for developers",
      "enabled": true,
      "rules": {
        "ssh-access": {
          "description": "Allow SSH",
          "enabled": true,
          "action": "accept",
          "bidirectional": false,
          "protocol": "tcp",
          "ports": ["22"],
          "sources": ["developers"],
          "destinations": ["production-servers"]
        }
      }
    }
  },
  "networks": {
    "office-network": {
      "description": "Main office network",
      "resources": {
        "internal-services": {
          "type": "subnet",
          "address": "10.0.0.0/24",
          "enabled": true,
          "description": "Internal service subnet",
          "groups": ["developers"]
        }
      },
      "routers": {
        "router-1": {
          "metric": 100,
          "masquerade": true,
          "enabled": true,
          "peer_groups": ["gateway-peers"]
        }
      }
    }
  },
  "routes": {
    "external-access": {
      "description": "Route to external services",
      "network": "192.168.1.0/24",
      "metric": 100,
      "masquerade": true,
      "enabled": true,
      "groups": ["developers"],
      "peer_groups": ["gateway-peers"]
    }
  },
  "dns": {
    "internal-dns": {
      "description": "Internal DNS servers",
      "nameservers": [
        {"ip": "10.0.0.53", "ns_type": "udp", "port": 53}
      ],
      "groups": ["developers"],
      "domains": ["internal.company.com"],
      "search_domains_enabled": true,
      "primary": false,
      "enabled": true
    }
  },
  "posture_checks": {
    "require-latest-version": {
      "description": "Require minimum NetBird version",
      "checks": {
        "nb_version_check": {
          "min_version": "0.28.0"
        }
      }
    }
  },
  "setup_keys": {
    "onboarding-key": {
      "description": "Type: reusable, State: valid",
      "type": "reusable",
      "expires_in": 30,
      "auto_groups": ["developers"],
      "usage_limit": 0,
      "ephemeral": false
    }
  }
}
```

---

## Import

Import NetBird configuration from YAML files. Running `netbird-manage import` by itself will display the help menu.

> **IMPORTANT: Peers cannot be imported via YAML.** Groups will be created/updated WITHOUT their peers. Peer data in the exported YAML is for reference and backup purposes only. To migrate peers between accounts, use the `migrate` command - see [Migrate](migrate.md).

### Import Operations

```bash
# Dry-run (preview changes without applying) - DEFAULT
netbird-manage import config.yml

# Apply changes
netbird-manage import --apply config.yml

# Import from split directory
netbird-manage import --apply ./netbird-manage-export-251117/

# Selective import (specific resource types)
netbird-manage import --apply --groups-only config.yml
netbird-manage import --apply --policies-only config.yml
```

### Conflict Resolution

When importing resources that already exist, choose how to handle conflicts:

```bash
# Default: Fail on conflicts (safe, prevents overwrites)
netbird-manage import config.yml

# Update existing resources
netbird-manage import --apply --update config.yml

# Skip existing resources
netbird-manage import --apply --skip-existing config.yml

# Force create or update (upsert)
netbird-manage import --apply --force config.yml
```

| Mode | Behavior | Use Case |
|------|----------|----------|
| **Default** | Fail on existing resources | Safe mode, requires manual resolution |
| `--update` | Update existing resources with YAML values | Apply configuration changes |
| `--skip-existing` | Skip resources that already exist | Import only new resources |
| `--force` | Create new or update existing (upsert) | Full declarative sync |

### Import Process

1. **Parse YAML** - Validate syntax and structure
2. **Fetch Current State** - Load existing resources from API
3. **Resolve References** - Convert names to IDs (e.g., group names → group IDs)
4. **Detect Conflicts** - Check for existing resources
5. **Validate** - Verify all references exist and data is valid
6. **Execute** - Apply changes in dependency order
7. **Report** - Show created/updated/skipped/failed resources

### Dependency Order

Resources are imported in the correct order to satisfy dependencies:
1. Groups (no dependencies)
2. Posture Checks (no dependencies)
3. Policies (depends on groups, posture checks)
4. Routes (depends on groups)
5. DNS (depends on groups)
6. Networks (depends on groups, policies)
7. Setup Keys (depends on groups)

### Example Output

```
▶ Importing NetBird configuration...

📦 Groups:
  ✓ CREATED  qa-team
  ⚠ SKIP     developers (already exists)
  ✓ UPDATED  production-servers

🔐 Policies:
  ✓ CREATED  allow-qa-access
  ✗ CONFLICT staging-policy (already exists, use --update)

================================================
📊 Import Summary
================================================

✓ Created:  2 resources
✓ Updated:  1 resource
⚠ Skipped:  1 resource
✗ Failed:   1 resource

Errors:
  1. Policy staging-policy: policy already exists

⚠ Fix errors and re-run with --skip-existing
```

### Notes

- **Dry-run by default** - Always preview before applying
- Flags must come **before** the filename: `netbird-manage import --apply config.yml`
- Partial failures are OK - successfully imported resources remain
- Use `--skip-existing` to re-import after fixing errors
- **Peers cannot be imported** - use `netbird-manage migrate` to move peers

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
| [Migrate](migrate.md) | Migration between NetBird accounts |

---

[Home](../README.md) | [Ingress Ports](ingress-ports.md) | **Export & Import** | [Migrate](migrate.md)
