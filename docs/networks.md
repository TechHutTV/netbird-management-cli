# Networks

[Home](../README.md) | [Getting Started](getting-started.md) | [Groups](groups.md) | **Networks** | [Policies](policies.md) | [Routes](routes.md) | [More...](#documentation)

---

Manage networks, resources, and routers. Running `netbird-manage network` by itself will display the help menu.

## Network Operations

### Query Operations

```bash
# List all networks
netbird-manage network --list

# Filter networks by name (supports wildcards)
netbird-manage network --list --filter-name "prod-*"

# Inspect a specific network (shows routers and resources)
netbird-manage network --inspect <network-id>
```

### Modification Operations

```bash
# Create a new network
netbird-manage network --create "Production-Network"

# Create a network with description
netbird-manage network --create "Staging-Network" --description "Staging environment network"

# Delete a network
netbird-manage network --delete <network-id>

# Rename a network
netbird-manage network --rename <network-id> --new-name "New-Network-Name"

# Update network description
netbird-manage network --update <network-id> --description "Updated description"
```

## Resource Management

Resources are hosts, subnets, or domains assigned to groups within a network.

### Query Operations

```bash
# List all resources in a network
netbird-manage network --list-resources <network-id>

# Inspect a specific resource
netbird-manage network --inspect-resource --network-id <network-id> --resource-id <resource-id>
```

### Modification Operations

```bash
# Add a resource (host IP)
netbird-manage network --add-resource <network-id> \
  --name "Web Server" \
  --address "192.168.1.100" \
  --groups "group-id-1,group-id-2" \
  --description "Production web server"

# Add a resource (subnet)
netbird-manage network --add-resource <network-id> \
  --name "Office Network" \
  --address "10.0.0.0/24" \
  --groups "group-id-1"

# Add a resource (domain with wildcard)
netbird-manage network --add-resource <network-id> \
  --name "API Services" \
  --address "*.api.example.com" \
  --groups "group-id-1"

# Add a disabled resource
netbird-manage network --add-resource <network-id> \
  --name "Maintenance Server" \
  --address "192.168.1.200" \
  --groups "group-id-1" \
  --disabled

# Update a resource
netbird-manage network --update-resource \
  --network-id <network-id> \
  --resource-id <resource-id> \
  --name "Updated Name" \
  --address "192.168.1.101" \
  --groups "new-group-id"

# Remove a resource
netbird-manage network --remove-resource \
  --network-id <network-id> \
  --resource-id <resource-id>
```

## Router Management

Routers are routing peers that enable traffic flow with configurable metrics and masquerading (NAT).

### Query Operations

```bash
# List all routers in a specific network
netbird-manage network --list-routers <network-id>

# List all routers across all networks
netbird-manage network --list-all-routers

# Inspect a specific router
netbird-manage network --inspect-router \
  --network-id <network-id> \
  --router-id <router-id>
```

### Modification Operations

```bash
# Add a router using a single peer
netbird-manage network --add-router <network-id> \
  --peer <peer-id> \
  --metric 100 \
  --masquerade

# Add a router using peer groups
netbird-manage network --add-router <network-id> \
  --peer-groups "group-id-1,group-id-2" \
  --metric 50 \
  --masquerade

# Add a router with custom settings
netbird-manage network --add-router <network-id> \
  --peer <peer-id> \
  --metric 200 \
  --no-masquerade \
  --disabled

# Update a router's configuration
netbird-manage network --update-router \
  --network-id <network-id> \
  --router-id <router-id> \
  --metric 75 \
  --masquerade \
  --enabled

# Change router to use peer groups instead of single peer
netbird-manage network --update-router \
  --network-id <network-id> \
  --router-id <router-id> \
  --peer-groups "new-group-id-1,new-group-id-2"

# Remove a router
netbird-manage network --remove-router \
  --network-id <network-id> \
  --router-id <router-id>
```

## Notes

- Resource addresses can be: direct hosts (`1.1.1.1` or `1.1.1.1/32`), subnets (`192.168.0.0/24`), or domains (`example.com`, `*.example.com`)
- Router metrics range from 1-9999 (lower = higher priority)
- Routers can use either a single `--peer` OR `--peer-groups`, but not both
- Masquerading enables NAT for traffic routed through the peer

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

[Home](../README.md) | [Groups](groups.md) | **Networks** | [Policies](policies.md) | [Routes](routes.md) | [DNS](dns.md)
