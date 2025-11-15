# NetBird Management CLI

netbird-manage is an unofficial command-line tool written in Go for interacting with the [NetBird](https://netbird.io/) API. It allows you to quickly manage peers, groups, policies, and other network resources directly from your terminal. This tool is built based on the official [NetBird REST API documentation](https://docs.netbird.io/api).

![](https://github.com/TechHutTV/netbird-management-cli/blob/main/demo.png)

## Setup & Installation

You must have the [Go toolchain](https://go.dev/doc/install) (version 1.18 or later) installed on your system. Place all .go files from this project into a new directory (e.g., netbird-manage). From inside that directory, initialize the Go module and run the go build command:  
```
go mod init netbird-manage
go build
```

## Current Commands & Functionality

### Connect
Before you can use the tool, you must authenticate. This tool stores your API token in a configuration file at `$HOME/.netbird-manage.conf`. Generate a Personal Access Token (PAT) or a Service User token from your NetBird dashboard. Then, run the connect command: 
```
netbird-manage connect --token <token>
```
If successful, you will see a "Connection successful" message. To check status or change api url see the flags below.

```
netbird-manage connect          Check current connection status
  connect [flags]               Connect and save your API token
    --token <token>             (Required) Your NetBird API token
    --management-url <url>      (Optional) Your self-hosted management URL
```

### Peer

Manage network peers. Running netbird-manage peer by itself will display the help menu.

#### Query Operations
```
netbird-manage peer --list                     List all peers in your network
  --filter-name <pattern>                      Filter by name (supports wildcards: ubuntu*)
  --filter-ip <pattern>                        Filter by IP address pattern

netbird-manage peer --inspect <peer-id>        View detailed information for a single peer

netbird-manage peer --accessible-peers <peer-id>  List peers accessible from the specified peer
```

#### Modification Operations
```
netbird-manage peer --remove <peer-id>         Remove a peer from your network

netbird-manage peer --edit <peer-id>           Edit peer group membership
  --add-group <group-id>                       Add peer to a specified group
  --remove-group <group-id>                    Remove peer from a specified group

netbird-manage peer --update <peer-id>         Update peer settings
  --rename <new-name>                          Change peer name
  --ssh-enabled <true|false>                   Enable/disable SSH access
  --login-expiration <true|false>              Enable/disable login expiration
  --inactivity-expiration <true|false>         Enable/disable inactivity expiration
  --approval-required <true|false>             Require approval (cloud-only)
  --ip <ip-address>                            Set IP (must be in 100.64.0.0/10 range)
```

**Examples:**
```bash
# List all peers with "ubuntu" in the name
netbird-manage peer --list --filter-name "ubuntu*"

# Rename a peer
netbird-manage peer --update d3mjakrl0ubs738ajj00 --rename "UbuntuServer"

# Enable SSH and disable login expiration
netbird-manage peer --update d3mjakrl0ubs738ajj00 --ssh-enabled true --login-expiration false

# Set a custom IP address
netbird-manage peer --update d3mjakrl0ubs738ajj00 --ip 100.64.1.50

# Check which peers a specific peer can access
netbird-manage peer --accessible-peers d3mjakrl0ubs738ajj00
```

### Group

Manage peer groups. Running netbird-manage group by itself will display the help menu.

#### Query Operations
```
netbird-manage group --list                    List all groups in your network
  --filter-name <pattern>                      Filter by name (supports wildcards: prod-*)

netbird-manage group --inspect <group-id>      View detailed information for a specific group
```

#### Modification Operations
```
netbird-manage group --create <group-name>     Create a new group
  --peers <id1,id2,...>                        (Optional) Add peers on creation

netbird-manage group --delete <group-id>       Delete a group

netbird-manage group --rename <group-id>       Rename a group
  --new-name <new-name>                        New name for the group

netbird-manage group --add-peers <group-id>    Add multiple peers to a group
  --peers <id1,id2,...>                        Comma-separated peer IDs

netbird-manage group --remove-peers <group-id> Remove multiple peers from a group
  --peers <id1,id2,...>                        Comma-separated peer IDs
```

**Examples:**
```bash
# Create a new group
netbird-manage group --create "Production-Servers"

# Create a group with initial peers
netbird-manage group --create "Dev-Team" --peers "peer-id-1,peer-id-2,peer-id-3"

# List all groups containing "prod" in the name
netbird-manage group --list --filter-name "prod*"

# Inspect a specific group
netbird-manage group --inspect d2l17grl0ubs73bh4vpg

# Rename a group
netbird-manage group --rename d2l17grl0ubs73bh4vpg --new-name "Production"

# Add multiple peers to a group at once
netbird-manage group --add-peers d2l17grl0ubs73bh4vpg --peers "peer1,peer2,peer3"

# Remove peers from a group
netbird-manage group --remove-peers d2l17grl0ubs73bh4vpg --peers "peer1,peer2"

# Delete a group
netbird-manage group --delete d2l17grl0ubs73bh4vpg
```

### Network

Manage networks, resources, and routers. Running `netbird-manage network` by itself will display the help menu.

#### Network Operations

##### Query Operations
```bash
# List all networks
netbird-manage network --list

# Filter networks by name (supports wildcards)
netbird-manage network --list --filter-name "prod-*"

# Inspect a specific network (shows routers and resources)
netbird-manage network --inspect <network-id>
```

##### Modification Operations
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

#### Resource Management

Resources are hosts, subnets, or domains assigned to groups within a network.

##### Query Operations
```bash
# List all resources in a network
netbird-manage network --list-resources <network-id>

# Inspect a specific resource
netbird-manage network --inspect-resource --network-id <network-id> --resource-id <resource-id>
```

##### Modification Operations
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

#### Router Management

Routers are routing peers that enable traffic flow with configurable metrics and masquerading (NAT).

##### Query Operations
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

##### Modification Operations
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

**Note:**
- Resource addresses can be: direct hosts (`1.1.1.1` or `1.1.1.1/32`), subnets (`192.168.0.0/24`), or domains (`example.com`, `*.example.com`)
- Router metrics range from 1-9999 (lower = higher priority)
- Routers can use either a single `--peer` OR `--peer-groups`, but not both
- Masquerading enables NAT for traffic routed through the peer

### Policy

Manage access control policies and firewall rules. Running `netbird-manage policy` by itself will display the help menu.

#### Query Operations
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

#### Policy Management
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

#### Rule Management

Add, edit, and remove rules within policies to control network traffic.

##### Add Rules
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

##### Edit Rules
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

##### Remove Rules
```bash
# Remove a rule by name
netbird-manage policy --remove-rule "web-access" --policy-id <policy-id>

# Remove a rule by ID
netbird-manage policy --remove-rule <rule-id> --policy-id <policy-id>
```

**Rule Configuration Options:**
- **`--action`**: `accept` or `drop` (default: accept)
- **`--protocol`**: `tcp`, `udp`, `icmp`, or `all` (default: all)
- **`--sources`**: Comma-separated group names or IDs
- **`--destinations`**: Comma-separated group names or IDs
- **`--ports`**: Comma-separated port list (e.g., `80,443,8080`)
- **`--port-range`**: Port range (e.g., `6000-6100`)
- **`--bidirectional`**: Apply rule in both directions (flag)
- **`--rule-description`**: Rule description text
- **`--rule-enabled`**: Enable/disable the rule (default: true)

**Note:**
- Group names are automatically resolved to IDs, so you can use friendly names
- Rules can be identified by either name or ID for editing/removal
- Bidirectional rules apply the same action in both source→destination and destination→source directions

## 🚀 Future Plans

This tool is in active development. The goal is to build a comprehensive and easy-to-use CLI for all NetBird management tasks.

* **Full API Coverage:** Implement the entire NetBird API, including:
  * ✅ Full CRUD (Create, Read, Update, Delete) for **Groups** - **COMPLETE**
  * ✅ Full CRUD for **Networks** and **Network Resources** - **COMPLETE**
  * ✅ Full CRUD for **Policies** and **Rules** - **COMPLETE**
  * Full **User Management** (invite, remove, update roles)
  * Management for **Setup Keys**, **Routes**, and **DNS**
* **YAML-based Policy Management:**
  * Add `netbird-manage policy export > my-policies.yml` to save all policies to a file
  * Add `netbird-manage policy apply -f my-policies.yml` to apply policy changes from a YAML file, enabling GitOps workflows
* **Interactive CLI Features:**
  * Implement interactive prompts for complex operations (e.g., `netbird-manage peer --remove <id>` asking for confirmation)
  * Use interactive selectors (like [bubbletea](https://github.com/charmbracelet/bubbletea)) for picking peers or groups from a list
