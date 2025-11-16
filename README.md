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

# Inspect a specific user
netbird-manage user --inspect <user-id>

# Get current authenticated user information
netbird-manage user --me
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

# Inspect a specific token
netbird-manage token --inspect <token-id>
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

**Important:**
- ⚠️ **Token values are only shown once during creation** - save them immediately!
- Tokens are used for API authentication via `Authorization: Token <token>` header
- Revoked tokens cannot be recovered - you must create new ones
- Use tokens for CI/CD pipelines, automation, and programmatic access

**Examples:**
```bash
# Create a token for automation
netbird-manage token --create --name "Terraform Token" --expires-in 365

# List all tokens to check expiration dates
netbird-manage token --list

# Revoke a compromised token
netbird-manage token --revoke tok-abc123xyz

# Check token details
netbird-manage token --inspect tok-abc123xyz
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

**API Coverage:** 10/14 NetBird API resource types fully implemented (71%)

### 📋 Planned Features

**Monitoring & Analytics (Phase 3):**
- ❌ **Events** - Audit logs and activity monitoring
- ❌ **Peer Update** - Modify peer properties (SSH, login expiration, IP assignment)
- ❌ **JSON Output** - Machine-readable output for scripting (`--output json`)
- ❌ **Accounts** - Account settings and configuration
- ❌ **Geo-Locations** - Location data for access control
- ❌ **Ingress Ports** - Port forwarding and ingress peers (Cloud only)

### 🎯 Enhancement Features

**GitOps & Automation:**
- ❌ **YAML Export/Import** - Infrastructure as Code workflows
  ```bash
  netbird-manage policy export > policies.yml
  netbird-manage policy apply -f policies.yml
  ```

**Interactive CLI:**
- ❌ **Confirmation Prompts** - Safety for destructive operations
- ❌ **Interactive Selection** - User-friendly resource picking with [bubbletea](https://github.com/charmbracelet/bubbletea)
- ❌ **TUI Mode** - Full-screen terminal interface with real-time updates
- ❌ **Shell Completion** - Tab completion for bash/zsh/fish

**Quality of Life:**
- ❌ **Batch Operations** - Process multiple resources at once
- ❌ **Colorized Output** - Improve readability with color coding
- ❌ **Debug Mode** - Verbose output showing HTTP requests/responses

For detailed implementation notes and architecture guidance, see [CLAUDE.md](CLAUDE.md).
