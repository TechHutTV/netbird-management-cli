# Policies

[Home](../README.md) | [Getting Started](getting-started.md) | [Groups](groups.md) | [Networks](networks.md) | **Policies** | [Routes](routes.md) | [More...](#documentation)

---

Manage access control policies and firewall rules. Running `netbird-manage policy` by itself will display the help menu.

## Query Operations

```bash
# List all policies
netbird-manage policy --list

# List only enabled policies
netbird-manage policy --list --enabled

# List only disabled policies
netbird-manage policy --list --disabled

# Filter policies by name
netbird-manage policy --list --name "dev"

# Inspect a specific policy (shows detailed rule information)
netbird-manage policy --inspect <policy-id>
```

## Policy Management

```bash
# Create a new policy
netbird-manage policy --create "dev-access" --description "Developer access policy"

# Create a disabled policy
netbird-manage policy --create "staging-policy" --description "Staging access" --active false

# Enable a policy
netbird-manage policy --enable <policy-id>

# Disable a policy
netbird-manage policy --disable <policy-id>

# Delete a policy
netbird-manage policy --delete <policy-id>
```

## Rule Management

Add, edit, and remove rules within policies to control network traffic.

### Add Rules

```bash
# Add a rule allowing TCP traffic on specific ports
netbird-manage policy --add-rule "web-access" \
  --policy-id <policy-id> \
  --action accept \
  --protocol tcp \
  --sources "developers,qa-team" \
  --destinations "web-servers" \
  --ports "80,443" \
  --bidirectional

# Add a rule with port range
netbird-manage policy --add-rule "app-ports" \
  --policy-id <policy-id> \
  --protocol tcp \
  --sources "app-servers" \
  --destinations "database" \
  --port-range "6000-6100"

# Add a rule blocking all ICMP traffic
netbird-manage policy --add-rule "block-ping" \
  --policy-id <policy-id> \
  --action drop \
  --protocol icmp \
  --sources "external-network" \
  --destinations "internal-servers"

# Add a rule with description
netbird-manage policy --add-rule "ssh-access" \
  --policy-id <policy-id> \
  --protocol tcp \
  --sources "admins" \
  --destinations "all-servers" \
  --ports "22" \
  --rule-description "Allow SSH access for administrators"
```

### Edit Rules

```bash
# Update ports on an existing rule
netbird-manage policy --edit-rule "web-access" \
  --policy-id <policy-id> \
  --ports "80,443,8443"

# Change rule action from accept to drop
netbird-manage policy --edit-rule "web-access" \
  --policy-id <policy-id> \
  --action drop

# Update source and destination groups
netbird-manage policy --edit-rule "web-access" \
  --policy-id <policy-id> \
  --sources "developers" \
  --destinations "web-servers,api-servers"

# Rename a rule
netbird-manage policy --edit-rule "old-rule-name" \
  --policy-id <policy-id> \
  --rule-name "new-rule-name"
```

### Remove Rules

```bash
# Remove a rule by name
netbird-manage policy --remove-rule "web-access" --policy-id <policy-id>

# Remove a rule by ID
netbird-manage policy --remove-rule <rule-id> --policy-id <policy-id>
```

## Rule Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--action` | `accept` or `drop` | accept |
| `--protocol` | `tcp`, `udp`, `icmp`, or `all` | all |
| `--sources` | Comma-separated group names or IDs | - |
| `--destinations` | Comma-separated group names or IDs | - |
| `--ports` | Comma-separated port list (e.g., `80,443,8080`) | - |
| `--port-range` | Port range (e.g., `6000-6100`) | - |
| `--bidirectional` | Apply rule in both directions | false |
| `--rule-description` | Rule description text | - |
| `--rule-enabled` | Enable/disable the rule | true |

## Notes

- Group names are automatically resolved to IDs, so you can use friendly names
- Rules can be identified by either name or ID for editing/removal
- Bidirectional rules apply the same action in both source→destination and destination→source directions

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

[Home](../README.md) | [Networks](networks.md) | **Policies** | [Routes](routes.md) | [DNS](dns.md) | [Posture Checks](posture-checks.md)
