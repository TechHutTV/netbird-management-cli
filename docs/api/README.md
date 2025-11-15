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

### Implemented in CLI

These resources have partial or full implementation in the netbird-manage CLI:

| Resource | Documentation | CLI Commands | Endpoints Used |
|----------|---------------|--------------|----------------|
| **Peers** | [peers.md](resources/peers.md) | `peer --list`, `--inspect`, `--remove`, `--edit` | GET /peers, GET /peers/{id}, DELETE /peers/{id} |
| **Groups** | [groups.md](resources/groups.md) | `group` (list), peer `--add-group`, `--remove-group` | GET /groups, GET /groups/{id}, PUT /groups/{id} |
| **Networks** | [networks.md](resources/networks.md) | `networks` (list) | GET /networks |
| **Policies** | [policies.md](resources/policies.md) | `policy` (list) | GET /policies |

### Not Yet Implemented

These resources are documented but not yet implemented in the CLI (planned for future releases):

| Resource | Documentation | Planned Features |
|----------|---------------|------------------|
| **Users** | [users.md](resources/users.md) | User management, invitations, role assignment |
| **Tokens** | [tokens.md](resources/tokens.md) | Token creation, management, revocation |
| **Accounts** | [accounts.md](resources/accounts.md) | Account settings, configuration |
| **DNS** | [dns.md](resources/dns.md) | Nameserver groups, DNS settings |
| **Routes** | [routes.md](resources/routes.md) | Network routing management |
| **Setup Keys** | [setup-keys.md](resources/setup-keys.md) | Device onboarding key management |
| **Posture Checks** | [posture-checks.md](resources/posture-checks.md) | Compliance validation rules |
| **Events** | [events.md](resources/events.md) | Audit logs, traffic monitoring |
| **Geo-Locations** | [geo-locations.md](resources/geo-locations.md) | Location data for policies |
| **Ingress Ports** | [ingress-ports.md](resources/ingress-ports.md) | Port forwarding (Cloud only) |

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
GET /peers                    → client.listPeers()           → peers.go:79
GET /peers/{id}              → client.getPeerByID()         → peers.go:122
DELETE /peers/{id}           → client.removePeerByID()      → peers.go:144
GET /groups                  → client.listGroups()          → groups.go:49
GET /groups/{id}             → client.getGroupByID()        → groups.go:80
PUT /groups/{id}             → client.updateGroup()         → groups.go:102
GET /networks                → client.listNetworks()        → networks.go:15
GET /policies                → client.listPolicies()        → policies.go:15
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
