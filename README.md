# NetBird Management CLI

netbird-manage is an unofficial command-line tool written in Go for interacting with the [NetBird](https://netbird.io/) API. It allows you to quickly manage peers, groups, policies, and other network resources directly from your terminal. This tool is built based on the official [NetBird REST API documentation](https://docs.netbird.io/api).

![](https://github.com/TechHutTV/netbird-management-cli/blob/main/demo.png)

## Setup & Installation

You must have the [Go toolchain](https://go.dev/doc/install) (version 1.18 or later) installed on your system. Clone this repository and build the binary:
```bash
git clone https://github.com/TechHutTV/netbird-management-cli.git
cd netbird-management-cli
go build ./cmd/netbird-manage
```

This creates a `netbird-manage` binary in the current directory. Optionally, move it to a directory in your PATH:
```bash
sudo mv netbird-manage /usr/local/bin/
```

## Safety Features

### Confirmation Prompts

To prevent accidental data loss, all destructive operations (delete, remove, revoke) now require explicit confirmation before executing. When you attempt to delete a resource, you'll see detailed information about what will be removed:

```bash
$ netbird-manage peer --remove abc123

About to remove peer:
  Name:      laptop-001
  ID:        abc123
  IP:        100.64.0.5
  Hostname:  laptop-001.local
  OS:        Linux
  Connected: true
  Groups:    2 (developers, ssh-access)

⚠️  This action cannot be undone. Continue? [y/N]: _
```

**For bulk operations** (like `--delete-unused` or `--delete-all`), you'll need to type a confirmation phrase:

```bash
$ netbird-manage group --delete-unused

🔴 This will delete 3 groups:
  - old-servers (ID: def456)
  - test-group (ID: ghi789)
  - unused-group (ID: jkl012)

Type 'delete 3 groups' to confirm: _
```

### Automation Mode

For scripts and automation, use the `--yes` (or `-y`) flag to skip all confirmation prompts:

```bash
# Skip confirmation for automation
netbird-manage --yes peer --remove abc123

# Also works with short flag
netbird-manage -y group --delete def456
```

**Warning:** When using `--yes`, deletions happen immediately without any prompts. Use with caution!

## Quality of Life Features

### Debug Mode

Enable verbose debug output to see all HTTP requests and responses. This is invaluable for troubleshooting API issues or understanding what's happening under the hood:

```bash
# Enable debug mode with --debug or -d flag
netbird-manage --debug peer --list

# Debug output shows:
═══ DEBUG: HTTP REQUEST ═══
GET https://api.netbird.io/api/peers
Headers:
  Authorization: Token [REDACTED]
  Accept: application/json

═══ DEBUG: HTTP RESPONSE ═══
Status: 200 OK
Headers:
  Content-Type: application/json
Response Body:
[
  {
    "id": "abc123",
    "name": "laptop-001",
    ...
  }
]
═══════════════════════════
```

**Features:**
- Shows full HTTP method and URL
- Displays request headers (token redacted for security)
- Pretty-prints JSON request/response bodies
- All debug output goes to stderr (keeps stdout clean for scripting)

### Batch Operations

Process multiple resources at once for efficient bulk operations. All batch operations support the same confirmation prompts as single deletions:

```bash
# Remove multiple peers at once
netbird-manage peer --remove-batch abc123,def456,ghi789

# Delete multiple groups
netbird-manage group --delete-batch dev-team,test-group,old-servers

# Delete multiple setup keys
netbird-manage setup-key --delete-batch key1,key2,key3

# Combine with --yes for automation
netbird-manage --yes peer --remove-batch abc123,def456,ghi789
```

**Batch operation features:**
- Fetches and displays details for all resources before confirmation
- Shows progress indicator during processing (e.g., `[2/5] Removing peer...`)
- Continues processing even if some operations fail
- Provides summary at the end: `Completed: 4 succeeded, 1 failed`
- Supports type-to-confirm for safety (type `delete N resources` to proceed)

**Example batch removal:**
```bash
$ netbird-manage group --delete-batch old-servers,test-group,unused

Fetching group details...
🔴 This will delete 3 groups:
  - old-servers (ID: abc123, Peers: 0, Resources: 0)
  - test-group (ID: def456, Peers: 2, Resources: 1)
  - unused (ID: ghi789, Peers: 0, Resources: 0)

Type 'delete 3 groups' to confirm: delete 3 groups

[1/3] Deleting group 'old-servers'... ✓ Done
[2/3] Deleting group 'test-group'... ✓ Done
[3/3] Deleting group 'unused'... ✓ Done

✓ All 3 groups deleted successfully
```

## Current Commands & Functionality

### Connect
Before you can use the tool, you must authenticate. This tool stores your API token in a configuration file at `$HOME/.netbird-manage.json`. Generate a Personal Access Token (PAT) or a Service User token from your NetBird dashboard. Then, run the connect command: 
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
netbird-manage peer --remove-batch <id1,id2,...>  Remove multiple peers (comma-separated IDs)

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

### Setup Key

Manage device registration and onboarding keys. Setup keys are used to register new peers to your NetBird network. Running `netbird-manage setup-key` by itself will display the help menu.

#### Query Operations
```bash
# List all setup keys
netbird-manage setup-key --list

# Filter by name (supports wildcards)
netbird-manage setup-key --list --filter-name "office-*"

# Filter by type (one-off or reusable)
netbird-manage setup-key --list --filter-type reusable

# Show only valid (non-revoked, non-expired) keys
netbird-manage setup-key --list --valid-only

# Inspect a specific setup key
netbird-manage setup-key --inspect <key-id>
```

#### Create Operations
```bash
# Quick create a one-off key (7d expiration, single use)
netbird-manage setup-key --quick "office-laptop"

# Create a one-off key with custom settings
netbird-manage setup-key --create "temp-access" \
  --type one-off \
  --expires-in 1d \
  --usage-limit 1

# Create a reusable key for team onboarding
netbird-manage setup-key --create "team-onboarding" \
  --type reusable \
  --expires-in 30d \
  --usage-limit 10 \
  --auto-groups "group-id-1,group-id-2"

# Create an ephemeral peer key (peer deleted when disconnected)
netbird-manage setup-key --create "ephemeral-test" \
  --expires-in 7d \
  --ephemeral
```

#### Update Operations
```bash
# Revoke a setup key (prevent new device registrations)
netbird-manage setup-key --revoke <key-id>

# Enable a previously revoked key
netbird-manage setup-key --enable <key-id>

# Update auto-groups for a key
netbird-manage setup-key --update-groups <key-id> \
  --groups "new-group-1,new-group-2"
```

#### Delete Operations
```bash
# Delete a setup key
netbird-manage setup-key --delete <key-id>

# Delete multiple setup keys at once
netbird-manage setup-key --delete-batch <key-id-1,key-id-2,key-id-3>

# Delete all setup keys (with confirmation)
netbird-manage setup-key --delete-all
```

**Examples:**
```bash
# Create a quick one-off key for a new office computer
netbird-manage setup-key --quick "johns-laptop"

# Create a reusable key that expires in 90 days for contractor onboarding
netbird-manage setup-key --create "contractor-key" \
  --type reusable \
  --expires-in 90d \
  --usage-limit 5 \
  --auto-groups "contractors,limited-access"

# List all valid keys
netbird-manage setup-key --list --valid-only

# Revoke a compromised key immediately
netbird-manage setup-key --revoke 12345

# Inspect a key to check usage statistics
netbird-manage setup-key --inspect 12345
```

**Key Configuration Options:**
- **`--type`**: `one-off` (single use) or `reusable` (multiple uses) - default: one-off
- **`--expires-in`**: Human-readable duration: `1d`, `7d`, `30d`, `90d`, `1y` - default: 7d
- **`--usage-limit`**: Maximum number of uses, `0` = unlimited - default: 0
- **`--auto-groups`**: Comma-separated group IDs for automatic peer assignment
- **`--ephemeral`**: Mark peers registered with this key as ephemeral (deleted when offline)
- **`--allow-extra-dns-labels`**: Allow additional DNS labels for registered peers

**Note:**
- Setup keys are displayed **only once** during creation - save them immediately!
- Expiration must be between 1 day and 1 year (API constraint)
- Revoked keys cannot be used to register new devices but existing devices remain active
- One-off keys are automatically revoked after first use

### User

Manage users and user invitations. Running `netbird-manage user` by itself will display the help menu.

#### Query Operations
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

#### Invite/Create Operations
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

#### Update Operations
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

#### Delete Operations
```bash
# Remove a user
netbird-manage user --remove <user-id>

# Resend invitation to a user
netbird-manage user --resend-invite <user-id>
```

**User Role Options:**
- **`admin`**: Full administrative access
- **`user`**: Standard user access
- **`owner`**: Account owner (highest privileges)

**Note:**
- Service users are designed for API access and automation
- Blocked users cannot access the system but their configuration is preserved
- Auto-groups automatically assign new peers to specified groups

### Token

Manage personal access tokens for API authentication. Running `netbird-manage token` by itself will display the help menu.

#### Query Operations
```bash
# List all personal access tokens
netbird-manage token --list

# List tokens when using service user token (requires --user-id)
netbird-manage token --list --user-id <user-id>

# Inspect a specific token
netbird-manage token --inspect <token-id>

# Inspect token when using service user token (requires --user-id)
netbird-manage token --inspect <token-id> --user-id <user-id>
```

#### Create Operations
```bash
# Create a token with default 90-day expiration
netbird-manage token --create --name "My CLI Token"

# Create a token with custom expiration
netbird-manage token --create \
  --name "CI/CD Token" \
  --expires-in 365

# Create a short-lived token
netbird-manage token --create \
  --name "Testing Token" \
  --expires-in 7
```

#### Delete Operations
```bash
# Revoke/delete a token
netbird-manage token --revoke <token-id>
```

**Token Options:**
- **`--expires-in`**: Expiration in days (1-365, default: 90)
- **`--user-id`**: User ID for token operations (defaults to current user)
  - **Required for service user tokens** since they cannot access `/users/current` endpoint
  - Use `netbird-manage user --list` to find your user ID

**Important:**
- ⚠️ **Token values are only shown once during creation** - save them immediately!
- Tokens are used for API authentication via `Authorization: Token <token>` header
- Revoked tokens cannot be recovered - you must create new ones
- Use tokens for CI/CD pipelines, automation, and programmatic access
- **Service user tokens**: Must provide `--user-id` flag for all token operations

**Examples:**
```bash
# Create a token for automation
netbird-manage token --create --name "Terraform Token" --expires-in 365

# List all tokens to check expiration dates
netbird-manage token --list

# List tokens when using service user token
netbird-manage token --list --user-id ef3799d6-3891-4769-9690-e6798258d5f6

# Revoke a compromised token
netbird-manage token --revoke tok-abc123xyz

# Check token details
netbird-manage token --inspect tok-abc123xyz

# Revoke token when using service user token
netbird-manage token --revoke tok-abc123xyz --user-id ef3799d6-3891-4769-9690-e6798258d5f6
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
netbird-manage group --delete-batch <id1,id2,...>  Delete multiple groups (comma-separated IDs)
netbird-manage group --delete-unused           Delete all unused groups (no peers, resources, or references)

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

# Delete all unused groups (scans for groups with no peers, resources, or references)
netbird-manage group --delete-unused
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

### Route

Manage network routes and routing configuration. Routes define how traffic flows through your NetBird network. Running `netbird-manage route` by itself will display the help menu.

#### Query Operations
```bash
# List all routes
netbird-manage route --list

# Filter routes by network CIDR pattern
netbird-manage route --list --filter-network "10.0"

# Filter by routing peer
netbird-manage route --list --filter-peer <peer-id>

# Show only enabled routes
netbird-manage route --list --enabled-only

# Show only disabled routes
netbird-manage route --list --disabled-only

# Inspect a specific route
netbird-manage route --inspect <route-id>
```

#### Modification Operations
```bash
# Create a route for 10.0.0.0/16 network
netbird-manage route --create "10.0.0.0/16" \
  --network-id <network-id> \
  --peer <peer-id> \
  --groups <group-id> \
  --metric 100 \
  --masquerade

# Create a route using peer groups instead of single peer
netbird-manage route --create "192.168.0.0/16" \
  --network-id <network-id> \
  --peer-groups "router-group-1,router-group-2" \
  --groups <group-id> \
  --metric 50

# Create a disabled route with description
netbird-manage route --create "172.16.0.0/12" \
  --network-id <network-id> \
  --peer <peer-id> \
  --groups <group-id> \
  --description "Private network route" \
  --disabled

# Update route metric (priority)
netbird-manage route --update <route-id> --metric 50

# Enable/disable a route
netbird-manage route --enable <route-id>
netbird-manage route --disable <route-id>

# Delete a route
netbird-manage route --delete <route-id>
```

**Route Configuration Options:**
- **`--network-id`**: Target network ID (required)
- **`--peer`**: Single routing peer ID (use OR `--peer-groups`)
- **`--peer-groups`**: Peer group IDs for high-availability routing (use OR `--peer`)
- **`--metric`**: Route priority (1-9999, lower = higher priority, default: 100)
- **`--masquerade`**: Enable masquerading/NAT (flag)
- **`--no-masquerade`**: Disable masquerading (default)
- **`--groups`**: Access group IDs (required, comma-separated)
- **`--description`**: Route description text

**Note:**
- Network must be in valid CIDR notation (e.g., `10.0.0.0/16`)
- Lower metric values have higher priority (metric 10 > metric 100)
- Masquerading enables NAT for outbound traffic
- Routes can use either a single peer or peer groups for redundancy

### DNS

Manage DNS nameserver groups and settings. DNS groups control domain resolution for specific peer groups. Running `netbird-manage dns` by itself will display the help menu.

#### Query Operations
```bash
# List all DNS nameserver groups
netbird-manage dns --list

# Filter by name pattern
netbird-manage dns --list --filter-name "corp-*"

# Show only primary DNS groups
netbird-manage dns --list --primary-only

# Show only enabled groups
netbird-manage dns --list --enabled-only

# Inspect a specific DNS group
netbird-manage dns --inspect <group-id>

# Get DNS settings for the account
netbird-manage dns --get-settings
```

#### Modification Operations
```bash
# Create a DNS group with Google and Cloudflare DNS
netbird-manage dns --create "corp-dns" \
  --nameservers "8.8.8.8:53,1.1.1.1:53" \
  --groups <group-id>

# Create a DNS group with domain matching
netbird-manage dns --create "internal-dns" \
  --nameservers "10.0.0.53:53" \
  --groups <group-id> \
  --domains "example.com,internal.local" \
  --search-domains \
  --primary

# Create DNS group with description
netbird-manage dns --create "public-dns" \
  --nameservers "1.1.1.1:53,8.8.8.8:53" \
  --groups <group-id> \
  --description "Public DNS resolvers for external access"

# Update nameservers for a group
netbird-manage dns --update <group-id> \
  --nameservers "9.9.9.9:53,149.112.112.112:53"

# Set a group as primary
netbird-manage dns --update <group-id> --primary

# Enable/disable a DNS group
netbird-manage dns --enable <group-id>
netbird-manage dns --disable <group-id>

# Update DNS settings (disable management for specific groups)
netbird-manage dns --update-settings \
  --disabled-groups <group-id-1>,<group-id-2>

# Delete a DNS group
netbird-manage dns --delete <group-id>
```

**DNS Configuration Options:**
- **`--nameservers`**: DNS servers in `IP:port` format (default port: 53)
- **`--groups`**: Target peer group IDs (required, comma-separated)
- **`--domains`**: Match specific domains (optional, comma-separated)
- **`--search-domains`**: Enable search domains (flag)
- **`--primary`**: Set as primary DNS group (flag)
- **`--description`**: DNS group description

**Note:**
- Nameserver format: `8.8.8.8:53` or just `8.8.8.8` (defaults to port 53)
- Primary DNS group is used when no domain-specific match is found
- Search domains append the domain to short hostnames
- Only one primary DNS group should be active at a time

### Posture Check

Manage device posture checks for zero-trust security. Posture checks validate device compliance before granting network access. Running `netbird-manage posture-check` by itself will display the help menu.

#### Query Operations
```bash
# List all posture checks
netbird-manage posture-check --list

# Filter by name pattern
netbird-manage posture-check --list --filter-name "version-*"

# Filter by check type
netbird-manage posture-check --list --filter-type "nb-version"

# Inspect a specific posture check
netbird-manage posture-check --inspect <check-id>
```

#### Check Types & Creation

##### NetBird Version Check
Ensure peers run a minimum NetBird version.
```bash
netbird-manage posture-check --create "min-nb-version" \
  --type nb-version \
  --min-version "0.28.0" \
  --description "Require NetBird 0.28.0 or newer"
```

##### OS Version Check
Validate minimum OS or kernel versions.
```bash
# Require macOS 13.0 or newer
netbird-manage posture-check --create "macos-13+" \
  --type os-version \
  --os darwin \
  --min-os-version "13.0"

# Require Linux kernel 5.10 or newer
netbird-manage posture-check --create "linux-kernel-5.10+" \
  --type os-version \
  --os linux \
  --min-kernel "5.10"

# Require Android 12 or newer
netbird-manage posture-check --create "android-12+" \
  --type os-version \
  --os android \
  --min-os-version "12"

# Supported OS types: android, darwin, ios, linux, windows
```

##### Geo-Location Check
Restrict or allow access based on geographic location.
```bash
# Allow access only from US
netbird-manage posture-check --create "us-only" \
  --type geo-location \
  --locations "US" \
  --action allow

# Allow specific cities
netbird-manage posture-check --create "office-locations" \
  --type geo-location \
  --locations "US:NewYork,GB:London,DE:Berlin" \
  --action allow

# Deny access from specific countries
netbird-manage posture-check --create "geo-block" \
  --type geo-location \
  --locations "RU,CN,KP" \
  --action deny
```

##### Network Range Check
Require peers to be on specific networks.
```bash
# Allow only corporate network ranges
netbird-manage posture-check --create "corporate-net" \
  --type network-range \
  --ranges "192.168.0.0/16,10.0.0.0/8" \
  --action allow

# Deny access from public networks
netbird-manage posture-check --create "block-public" \
  --type network-range \
  --ranges "0.0.0.0/0" \
  --action deny
```

##### Process Check
Verify required security software is running.
```bash
# Check for antivirus on multiple platforms
netbird-manage posture-check --create "av-required" \
  --type process \
  --linux-path "/usr/bin/clamav" \
  --mac-path "/Applications/Antivirus.app" \
  --windows-path "C:\\Program Files\\Antivirus\\av.exe"

# Check for VPN client
netbird-manage posture-check --create "vpn-client" \
  --type process \
  --windows-path "C:\\Program Files\\VPN\\client.exe"
```

#### Modification Operations
```bash
# Update a posture check
netbird-manage posture-check --update <check-id> \
  --type nb-version \
  --min-version "0.29.0" \
  --description "Updated minimum version"

# Delete a posture check
netbird-manage posture-check --delete <check-id>
```

**Posture Check Types:**
- **`nb-version`**: NetBird version check
- **`os-version`**: Operating system version check
- **`geo-location`**: Geographic location check
- **`network-range`**: Peer network range check
- **`process`**: Running process check

**Note:**
- Posture checks are evaluated before granting network access
- Multiple checks can be combined for defense-in-depth security
- Location checks use ISO 3166-1 alpha-2 country codes (e.g., US, GB, DE)
- Process checks support platform-specific paths for cross-platform security

### Event

Monitor audit logs and network traffic events. Events provide visibility into network activity and user actions. Running `netbird-manage event` by itself will display the help menu.

#### Audit Events
```bash
# List all audit events
netbird-manage event --audit

# Filter by user ID
netbird-manage event --audit --user-id <user-id>

# Filter by target resource
netbird-manage event --audit --target-id <resource-id>

# Filter by activity type
netbird-manage event --audit --activity-code peer.create

# Filter by date range
netbird-manage event --audit --start-date "2025-01-01T00:00:00Z" --end-date "2025-01-31T23:59:59Z"

# Search in initiator/target names
netbird-manage event --audit --search "laptop"

# Export to JSON
netbird-manage event --audit --output json > audit.json
```

#### Network Traffic Events (Cloud-only)
```bash
# List network traffic events
netbird-manage event --traffic

# Filter by protocol (6=TCP, 17=UDP, 1=ICMP)
netbird-manage event --traffic --protocol 6

# Filter by direction
netbird-manage event --traffic --direction incoming

# Filter by reporting peer
netbird-manage event --traffic --reporter-id <peer-id>

# Pagination
netbird-manage event --traffic --page 2 --page-size 50

# Export to JSON
netbird-manage event --traffic --output json > traffic.json
```

**Examples:**
```bash
# View recent peer creation events
netbird-manage event --audit --activity-code peer.create

# Monitor TCP traffic from the last 24 hours
netbird-manage event --traffic --protocol 6 --start-date "2025-01-15T00:00:00Z"

# Search for events related to a specific user
netbird-manage event --audit --search "admin@example.com"
```

**Note:**
- Audit events track all management actions (create, update, delete)
- Traffic events are an experimental feature available only on NetBird Cloud
- Events support both table and JSON output formats for integration with other tools

### Geo-Location

Retrieve geographic location data for use in posture checks and access policies. Running `netbird-manage geo` by itself will display the help menu.

#### Query Operations
```bash
# List all country codes
netbird-manage geo --countries

# List cities in a specific country
netbird-manage geo --cities --country DE
netbird-manage geo --cities --country US

# Export to JSON
netbird-manage geo --countries --output json
netbird-manage geo --cities --country FR --output json
```

**Examples:**
```bash
# Get all available country codes
netbird-manage geo --countries

# Find cities in Germany for geo-location posture checks
netbird-manage geo --cities --country DE

# Export US cities to JSON for automation
netbird-manage geo --cities --country US --output json > us-cities.json
```

**Note:**
- Country codes follow ISO 3166-1 alpha-2 standard (e.g., US, GB, DE, FR)
- City data includes geoname IDs for precise location matching
- Use geo-location data when creating posture checks with `--type geo-location`

### Account

Manage account settings and configuration. Running `netbird-manage account` by itself will display the help menu.

#### Query Operations
```bash
# List all accounts (returns current user's account)
netbird-manage account --list

# Inspect account details
netbird-manage account --inspect <account-id>
```

#### Update Operations
```bash
# Update peer login expiration
netbird-manage account --update <account-id> --peer-login-expiration 48h

# Update peer inactivity expiration
netbird-manage account --update <account-id> --peer-inactivity-expiration 30d

# Update DNS domain
netbird-manage account --update <account-id> --dns-domain nb.local

# Update network range
netbird-manage account --update <account-id> --network-range 100.64.0.0/10

# Enable JWT groups
netbird-manage account --update <account-id> --jwt-groups-enabled true

# Update multiple settings at once
netbird-manage account --update <account-id> \
  --peer-login-expiration 24h \
  --dns-domain company.local \
  --jwt-groups-enabled true
```

#### Delete Operations
```bash
# Delete account (requires confirmation, deletes ALL resources)
netbird-manage account --delete <account-id> --confirm
```

**Examples:**
```bash
# View current account settings
netbird-manage account --list

# Set peer login expiration to 2 days
netbird-manage account --update d10vfhbl0ubs73e6p8ig --peer-login-expiration 48h

# Configure JWT group claims
netbird-manage account --update d10vfhbl0ubs73e6p8ig \
  --jwt-groups-enabled true \
  --jwt-groups-claim groups \
  --jwt-allow-groups "engineering,ops,security"
```

**Note:**
- Duration format: `24h` (hours), `7d` (days), `30d` (days)
- Some settings like `peer-approval-enabled` and `traffic-logging` are Cloud-only
- Deleting an account is permanent and removes ALL associated resources

### Ingress Port

Manage port forwarding and ingress peers. **Cloud-only feature** - only available on NetBird Cloud. Running `netbird-manage ingress-port` by itself will display the help menu.

#### Port Allocation Operations
```bash
# List port allocations for a peer
netbird-manage ingress-port --list --peer <peer-id>

# Inspect a port allocation
netbird-manage ingress-port --inspect <allocation-id> --peer <peer-id>

# Create a port allocation
netbird-manage ingress-port --create --peer <peer-id> \
  --target-port 8080 \
  --protocol tcp \
  --description "Web Server"

# Update a port allocation
netbird-manage ingress-port --update <allocation-id> --peer <peer-id> \
  --target-port 8443

# Delete a port allocation
netbird-manage ingress-port --delete <allocation-id> --peer <peer-id>
```

#### Ingress Peer Operations
```bash
# List all ingress peers
netbird-manage ingress-peer --list

# Inspect an ingress peer
netbird-manage ingress-peer --inspect <ingress-peer-id>

# Create an ingress peer
netbird-manage ingress-peer --create \
  --name "US West" \
  --location us-west-1

# Update an ingress peer
netbird-manage ingress-peer --update <ingress-peer-id> \
  --enabled false

# Delete an ingress peer
netbird-manage ingress-peer --delete <ingress-peer-id>
```

**Examples:**
```bash
# Forward port 8080 on a peer to a public port
netbird-manage ingress-port --create --peer d41uqobl0ubs73bkuhqg \
  --target-port 8080 \
  --protocol tcp \
  --description "Production Web Server"

# List all port allocations for a specific peer
netbird-manage ingress-port --list --peer d41uqobl0ubs73bkuhqg

# Create a new ingress peer for EU region
netbird-manage ingress-peer --create \
  --name "EU Central" \
  --location eu-central-1

# Disable an ingress peer
netbird-manage ingress-peer --update ing-001 --enabled false
```

**Note:**
- Ingress ports are **Cloud-only** - not available on self-hosted instances
- Public ports are automatically assigned by NetBird Cloud
- Protocol options: `tcp` (default) or `udp`
- Target ports must be between 1-65535

### Export

Export your NetBird configuration to YAML files for GitOps workflows, backup, and infrastructure-as-code. Running `netbird-manage export` by itself will display the help menu.

#### Export Operations
```bash
# Export to single YAML file (default)
netbird-manage export
netbird-manage export --full

# Export to single file in specific directory
netbird-manage export --full ./backups

# Export to split files (one file per resource type)
netbird-manage export --split

# Export to split files in specific directory
netbird-manage export --split ~/exports
```

**Output Formats:**

**Single File** (`netbird-manage-export-YYMMDD.yml`):
- All resources in one file
- Clean map-based YAML structure
- Perfect for backups and small deployments
- Easy to review entire configuration

**Split Files** (`netbird-manage-export-YYMMDD/`):
```
├── config.yml          # Metadata and import order
├── groups.yml          # Group definitions with peer names
├── policies.yml        # Access control policies with rules
├── networks.yml        # Networks with resources and routers
├── routes.yml          # Custom network routes
├── dns.yml             # DNS nameserver groups
├── posture-checks.yml  # Device compliance checks
└── setup-keys.yml      # Device onboarding keys
```

**Examples:**
```bash
# Quick backup to current directory
netbird-manage export

# Backup to specific folder
netbird-manage export --full ~/netbird-backups

# GitOps-friendly split export
netbird-manage export --split

# Export for version control
netbird-manage export --split ./git-repo/netbird-config
```

**YAML Structure Features:**
- **Human-readable**: Uses resource names as keys instead of UUIDs
- **Map-based**: Minimal use of list markers (`-`) for clean structure
- **Name resolution**: Group IDs converted to group names automatically
- **Dependency-aware**: Split mode includes proper import order
- **Version controlled**: Ideal for Git workflows and change tracking

**Exported Resources:**
- ✅ Groups (with peer names - **for reference only, not imported**)
- ✅ Policies (with rules and group references)
- ✅ Networks (with resources and routers)
- ✅ Routes (with routing configuration)
- ✅ DNS (nameserver groups and domains)
- ✅ Posture Checks (all 5 check types)
- ✅ Setup Keys (with auto-groups)

**Sample YAML Structure:**
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

**Note:**
- Export uses current API configuration from `~/.netbird-manage.json`
- Timestamp format: YYMMDD (e.g., `251117` for November 17, 2025)
- Exported YAML is designed for import functionality
- Split mode is recommended for large configurations and team collaboration
- **Peer names in groups are for reference/backup only** - use `netbird-manage migrate` to move peers

### Import

Import NetBird configuration from YAML files for declarative infrastructure-as-code management. Running `netbird-manage import` by itself will display the help menu.

> ⚠️ **IMPORTANT: Peers cannot be imported via YAML.** Groups will be created/updated WITHOUT their peers. Peer data in the exported YAML is for reference and backup purposes only. To migrate peers between accounts, use the `migrate` command - see [Migrate](#migrate) section below.

#### Import Operations
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

#### Conflict Resolution

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

#### Conflict Resolution Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| **Default** | Fail on existing resources | Safe mode, requires manual resolution |
| `--update` | Update existing resources with YAML values | Apply configuration changes |
| `--skip-existing` | Skip resources that already exist | Import only new resources |
| `--force` | Create new or update existing (upsert) | Full declarative sync |

#### Examples

```bash
# Preview what would be imported (dry-run, default)
netbird-manage import config.yml

# Import new resources, skip existing
netbird-manage import --apply --skip-existing config.yml

# Update all resources from YAML
netbird-manage import --apply --update config.yml

# Import only groups
netbird-manage import --apply --groups-only groups.yml

# Import from split directory
netbird-manage import --apply ./netbird-export-dir/

# Verbose output
netbird-manage import --apply --verbose config.yml
```

**Import Process:**
1. **Parse YAML** - Validate syntax and structure
2. **Fetch Current State** - Load existing resources from API
3. **Resolve References** - Convert names to IDs (e.g., group names → group IDs)
4. **Detect Conflicts** - Check for existing resources
5. **Validate** - Verify all references exist and data is valid
6. **Execute** - Apply changes in dependency order (groups → policies → networks)
7. **Report** - Show created/updated/skipped/failed resources

**Dependency Order:**
Resources are imported in the correct order to satisfy dependencies:
1. Groups (no dependencies)
2. Posture Checks (no dependencies)
3. Policies (depends on groups, posture checks)
4. Routes (depends on groups)
5. DNS (depends on groups)
6. Networks (depends on groups, policies)
7. Setup Keys (depends on groups)

**Smart Name Resolution:**
- Group names automatically resolved to IDs
- Missing groups referenced in policies cause errors
- **Peers are NOT imported** - use `netbird-manage migrate` instead

**Validation:**
- ✅ Valid YAML syntax
- ✅ Required fields present
- ✅ Valid field values (ports, CIDRs, protocols)
- ✅ Referenced resources exist
- ✅ No circular dependencies

**Example Output:**
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

**Note:**
- **Dry-run by default** - Always preview before applying
- Flags must come **before** the filename: `netbird-manage import --apply config.yml`
- Partial failures are OK - successfully imported resources remain
- Use `--skip-existing` to re-import after fixing errors
- **Peers cannot be imported** - groups are created empty; use `netbird-manage migrate` to move peers between accounts
- Policy rules reference groups by name (automatically resolved)

### Migrate

Migrate peers between NetBird accounts. This command generates migration commands that can be run on peer devices to move them from one NetBird account to another. Running `netbird-manage migrate` by itself will display the help menu.

#### Use Cases

- **Cloud to Self-Hosted Migration**: Move peers from NetBird Cloud to your own self-hosted instance
- **Self-Hosted to Cloud**: Move peers from self-hosted to NetBird Cloud
- **Account Consolidation**: Merge multiple NetBird accounts into one
- **Organization Transfers**: Move network infrastructure between organizations

#### Single Peer Migration

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

#### Batch Migration (by Group)

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

#### Example Output

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

#### Batch Migration Output

```
$ netbird-manage migrate \
    --source-token "nbp_Sxxxx..." \
    --dest-token "nbp_Dxxxx..." \
    --group "old-servers"

Fetching peers in group 'old-servers' from source...
  Source: https://api.netbird.io/api

Found 3 peers to migrate.

Connecting to destination account...
  Destination: https://api.netbird.io/api

Peer 1/3: server-web-01
  Creating setup key... Done
Peer 2/3: server-db-01
  Creating setup key... Done
Peer 3/3: server-cache-01
  Creating setup key... Done

========================================================================
MIGRATION COMMANDS - Run on each peer:
========================================================================

# server-web-01 (100.64.0.20)
sudo netbird down && sudo netbird up --setup-key AAAA-AAAA-AAAA-AAAA --hostname server-web-01

# server-db-01 (100.64.0.21)
sudo netbird down && sudo netbird up --setup-key BBBB-BBBB-BBBB-BBBB --hostname server-db-01

# server-cache-01 (100.64.0.22)
sudo netbird down && sudo netbird up --setup-key CCCC-CCCC-CCCC-CCCC --hostname server-cache-01

========================================================================
```

#### Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--source-token` | Yes | - | API token for source account |
| `--source-url` | No | `https://api.netbird.io/api` | Management URL for source |
| `--dest-token` | Yes | - | API token for destination account |
| `--dest-url` | No | `https://api.netbird.io/api` | Management URL for destination |
| `--peer` | Yes* | - | Peer ID to migrate |
| `--group` | Yes* | - | Migrate all peers in group |
| `--create-groups` | No | `true` | Create missing groups in destination |
| `--key-expiry` | No | `24h` | Setup key expiration (e.g., 1h, 24h, 7d) |

*Either `--peer` or `--group` is required.

**What Gets Migrated:**
- ✅ Hostname (preserved via `--hostname` flag)
- ✅ Group memberships (groups created in destination if missing)
- ❌ Peer ID (new one assigned by destination)
- ❌ IP address (new one assigned by destination network)
- ❌ Connection history/statistics

**Important Notes:**
- The peer must be accessible to run the migration command on it
- Setup keys are single-use and expire after the specified duration
- The old peer entry in the source account must be manually removed after migration
- Groups are automatically created in the destination if they don't exist (unless `--create-groups=false`)

## 🚀 Roadmap

This tool is in active development. The goal is to build a comprehensive and easy-to-use CLI for all NetBird management tasks.

### ✅ Completed Features

**Core Resource Management:**
- ✅ **Peers** - List, inspect, remove, and group assignment
- ✅ **Groups** - Full CRUD operations with bulk peer management
- ✅ **Networks** - Full CRUD operations including resource and router management
- ✅ **Policies** - Full CRUD operations with advanced rule management
- ✅ **Setup Keys** - Full CRUD operations for device registration and onboarding
- ✅ **Users** - Full user management including invites, roles, and permissions
- ✅ **Tokens** - Personal access token management for secure API access

**Network Services (Phase 2 - COMPLETED):**
- ✅ **Routes** - Network routing configuration with metrics and masquerading
- ✅ **DNS** - DNS nameserver groups with domain matching and settings
- ✅ **Posture Checks** - Device compliance validation with 5 check types

**Monitoring & Analytics (Phase 3 - COMPLETED):**
- ✅ **Events** - Audit logs and network traffic monitoring with filtering and pagination
- ✅ **Geo-Locations** - Country/city location data for posture checks and policies
- ✅ **JSON Output** - Machine-readable output for events and geo-location commands

**Account Management (Phase 4 - COMPLETED):**
- ✅ **Accounts** - Account settings and configuration management
- ✅ **Ingress Ports** - Port forwarding and ingress peer management (Cloud only)
- ✅ **Peer Update** - Modify peer properties (SSH, login expiration, IP address)
- ✅ **Accessible Peers** - Query peer connectivity and reachability

**GitOps & Infrastructure-as-Code (Phase 5 - COMPLETED):**
- ✅ **YAML Export** - Export all configuration to YAML files (single or split mode)
- ✅ **YAML Import** - Apply YAML configuration to NetBird with conflict resolution

**Safety & UX Enhancements (Phase 6 - COMPLETED):**
- ✅ **Confirmation Prompts** - Prevent accidental deletions with detailed resource info and Y/N prompts
- ✅ **Bulk Deletion Confirmations** - Type-to-confirm for operations affecting multiple resources
- ✅ **Automation Mode** - Global `--yes` flag to skip confirmations for scripts and automation

**Quality of Life (Phase 7 - COMPLETED):**
- ✅ **Batch Operations** - Process multiple resources at once with progress indicators
- ✅ **Debug Mode** - Verbose HTTP request/response logging with `--debug` or `-d` flag

**Migration Tools (Phase 8 - COMPLETED):**
- ✅ **Peer Migration** - Migrate peers between NetBird accounts (cloud-to-cloud, cloud-to-self-hosted, or self-hosted-to-cloud)
- ✅ **Batch Migration** - Migrate all peers in a group with a single command
- ✅ **Auto Group Creation** - Automatically create missing groups in destination account

**API Coverage:** 14/14 NetBird API resource types fully implemented (100%) 🎉

### 📋 Planned Features

**Developer Experience:**
- ❌ **Shell Completion** - Tab completion for bash/zsh/fish

For detailed implementation notes and architecture guidance, see [CLAUDE.md](CLAUDE.md).
