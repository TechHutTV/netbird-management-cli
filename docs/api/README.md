# NetBird API Documentation

Complete reference documentation for the NetBird Management API, organized for AI assistants and developers working with the netbird-management-cli project.

## Documentation Structure

### Getting Started
- **[Introduction](introduction.md)** - API overview, authentication basics, and quick examples
- **[Quickstart Guide](guides/quickstart.md)** - Make your first API request in minutes
- **[Authentication](guides/authentication.md)** - OAuth2 and Personal Access Token setup
- **[Error Handling](guides/errors.md)** - Understanding and handling API errors

### API Resources
- **[Resources Index](resources/README.md)** - Overview of all API endpoints
- **[Individual Resources](resources/)** - Detailed documentation per resource type

## Quick Navigation

### ✅ Fully Implemented in CLI

These resources have complete implementation in the netbird-manage CLI:

| Resource | Documentation | CLI Commands | Coverage |
|----------|---------------|--------------|----------|
| **Peers** | [peers.md](resources/peers.md) | `peer --list`, `--inspect`, `--update`, `--remove`, `--accessible-peers`, `--edit` | 5/5 endpoints |
| **Groups** | [groups.md](resources/groups.md) | `group --list`, `--inspect`, `--create`, `--delete`, `--rename`, `--add-peers`, `--remove-peers` | 5/5 endpoints |
| **Networks** | [networks.md](resources/networks.md) | `network` (full CRUD + resources + routers) | 12/12 endpoints |
| **Policies** | [policies.md](resources/policies.md) | `policy` (full CRUD + rule management) | 5/5 endpoints |
| **Setup Keys** | [setup-keys.md](resources/setup-keys.md) | `setup-key --list`, `--inspect`, `--create`, `--quick`, `--delete`, `--revoke`, `--enable`, `--update-groups` | 5/5 endpoints |
| **Users** | [users.md](resources/users.md) | `user --list`, `--me`, `--invite`, `--update`, `--remove`, `--resend-invite` | 6/6 endpoints |
| **Tokens** | [tokens.md](resources/tokens.md) | `token --list`, `--inspect`, `--create`, `--revoke` | 4/4 endpoints |
| **Routes** | [routes.md](resources/routes.md) | `route --list`, `--inspect`, `--create`, `--update`, `--delete`, `--enable`, `--disable` | 5/5 endpoints |
| **DNS** | [dns.md](resources/dns.md) | `dns --list`, `--inspect`, `--create`, `--update`, `--delete`, `--enable`, `--disable`, `--get-settings`, `--update-settings` | 6/6 endpoints |
| **Posture Checks** | [posture-checks.md](resources/posture-checks.md) | `posture-check --list`, `--inspect`, `--create`, `--update`, `--delete` (5 check types) | 5/5 endpoints |

**Total API Coverage: 10/14 resource types (71%)** - 58 total endpoints implemented

### ❌ Not Yet Implemented

These resources are documented but not yet implemented in the CLI (planned for future releases):

| Resource | Documentation | Planned Features |
|----------|---------------|------------------|
| **Accounts** | [accounts.md](resources/accounts.md) | Account settings, configuration (3 endpoints) |
| **Events** | [events.md](resources/events.md) | Audit logs, traffic monitoring (2 endpoints) |
| **Geo-Locations** | [geo-locations.md](resources/geo-locations.md) | Location data for policies (2 endpoints) |
| **Ingress Ports** | [ingress-ports.md](resources/ingress-ports.md) | Port forwarding - NetBird Cloud only (10 endpoints) |

## API Base URLs

**NetBird Cloud (Default):**
```
https://api.netbird.io/api
```

**Self-Hosted:**
```
https://your-server.com/api
```

## Authentication

All requests require authentication via the `Authorization` header:

**Personal Access Token (Recommended for CLI):**
```bash
Authorization: Token <YOUR_PAT>
```

**OAuth2 Bearer Token:**
```bash
Authorization: Bearer <YOUR_TOKEN>
```

See [Authentication Guide](guides/authentication.md) for complete details.

## Common Request Examples

### List Peers
```bash
curl -X GET https://api.netbird.io/api/peers \
  -H 'Authorization: Token <TOKEN>' \
  -H 'Accept: application/json'
```

### Create a Group
```bash
curl -X POST https://api.netbird.io/api/groups \
  -H 'Authorization: Token <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "my-group",
    "peers": [],
    "resources": []
  }'
```

### Update Peer Groups
```bash
# Get group details first
curl -X GET https://api.netbird.io/api/groups/{groupId} \
  -H 'Authorization: Token <TOKEN>'

# Update group with modified peer list
curl -X PUT https://api.netbird.io/api/groups/{groupId} \
  -H 'Authorization: Token <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "group-name",
    "peers": ["peer-1", "peer-2", "new-peer-3"],
    "resources": []
  }'
```

## For AI Assistants

### Using This Documentation

When working on the netbird-management-cli project:

1. **For existing features:** Check the implemented resources above and reference their documentation files
2. **For new features:** Review the "Not Yet Implemented" resources for API endpoint details
3. **For debugging:** Use the [Error Handling Guide](guides/errors.md)
4. **For authentication issues:** Refer to the [Authentication Guide](guides/authentication.md)

### CLI Implementation Patterns

The CLI follows these patterns (see `CLAUDE.md` for complete details):

**Data Flow:**
```
CLI Command → handleCommand() → client.makeRequest() → API Endpoint → JSON Response → Display
```

**Example: Adding Peer to Group**
1. User runs: `netbird-manage peer --edit <peer-id> --add-group <group-name>`
2. CLI calls: `client.getGroupByName(groupName)` → GET /groups + filter
3. CLI modifies peer list in memory
4. CLI calls: `client.updateGroup(groupID, updatedGroup)` → PUT /groups/{id}

See individual resource documentation for endpoint-specific patterns.

### Quick Reference: API → CLI Mapping

```
# Peers (5/5 endpoints)
GET /peers                        → client.listPeers()           → peers.go
GET /peers/{id}                   → client.getPeerByID()         → peers.go
PUT /peers/{id}                   → client.updatePeer()          → peers.go
DELETE /peers/{id}                → client.removePeerByID()      → peers.go
GET /peers/{id}/accessible-peers  → client.getAccessiblePeers()  → peers.go

# Groups (5/5 endpoints)
GET /groups                   → client.listGroups()          → groups.go
POST /groups                  → client.createGroup()         → groups.go
GET /groups/{id}              → client.getGroupByID()        → groups.go
PUT /groups/{id}              → client.updateGroup()         → groups.go
DELETE /groups/{id}           → client.deleteGroup()         → groups.go

# Networks (12/12 endpoints - full CRUD + resources + routers)
GET /networks                 → client.listNetworks()        → networks.go
POST /networks                → client.createNetwork()       → networks.go
GET /networks/{id}            → client.getNetwork()          → networks.go
PUT /networks/{id}            → client.updateNetwork()       → networks.go
DELETE /networks/{id}         → client.deleteNetwork()       → networks.go
# ... plus 7 resource/router endpoints

# Policies (5/5 endpoints + rule management)
GET /policies                 → client.listPolicies()        → policies.go
POST /policies                → client.createPolicy()        → policies.go
GET /policies/{id}            → client.getPolicy()           → policies.go
PUT /policies/{id}            → client.updatePolicy()        → policies.go
DELETE /policies/{id}         → client.deletePolicy()        → policies.go

# Plus: Setup Keys, Users, Tokens, Routes, DNS, Posture Checks
# See individual resource docs for details
```

## External Resources

- **Official NetBird API Docs:** https://docs.netbird.io/api
- **OpenAPI Specification:** https://api.netbird.io/api/openapi.json
- **NetBird GitHub:** https://github.com/netbirdio/netbird
- **NetBird Website:** https://netbird.io

## Document Status

- **Last Updated:** 2025-11-15
- **API Version:** v1 (Beta)
- **Documentation Version:** 1.0
- **CLI Compatibility:** netbird-manage v0.1 (current codebase)

## Contributing

When updating this documentation:

1. Keep CLI implementation mappings current
2. Add examples for new endpoints
3. Update the "Implemented in CLI" table when adding features
4. Test all code examples before committing
5. Link to official docs for comprehensive details

---

**Ready to get started?** Begin with the [Quickstart Guide](guides/quickstart.md) or explore [API Resources](resources/README.md).
