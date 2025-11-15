# NetBird API Reference - Quick Navigation

This file provides quick-reference information for the NetBird Management API. For comprehensive documentation, see the `docs/api/` directory.

## 📚 Complete Documentation

**Main Documentation:** [`docs/api/README.md`](docs/api/README.md)

### Getting Started
- **[Introduction](docs/api/introduction.md)** - API overview and quick examples
- **[Quickstart Guide](docs/api/guides/quickstart.md)** - Your first API request
- **[Authentication](docs/api/guides/authentication.md)** - OAuth2 and PAT setup
- **[Error Handling](docs/api/guides/errors.md)** - Understanding API responses

### API Resources
- **[All Resources](docs/api/resources/README.md)** - Complete endpoint catalog
- **Individual Resources:** See `docs/api/resources/` for detailed per-resource documentation

---

## Quick Reference

### Base URLs

**Cloud (Default):**
```
https://api.netbird.io/api
```

**Self-Hosted:**
```
https://your-server.com/api
```

### Authentication

```bash
# Personal Access Token (Recommended for CLI)
Authorization: Token <YOUR_PAT>

# OAuth2 Bearer Token
Authorization: Bearer <YOUR_TOKEN>
```

---

## CLI Implementation Status

### ✅ Implemented Endpoints

| Resource | CLI Command | API Endpoint | Implementation File |
|----------|-------------|--------------|---------------------|
| List Peers | `peer --list` | `GET /peers` | `peers.go:79` |
| Inspect Peer | `peer --inspect <id>` | `GET /peers/{id}` | `peers.go:122` |
| Remove Peer | `peer --remove <id>` | `DELETE /peers/{id}` | `peers.go:144` |
| List Groups | `group --list` | `GET /groups` | `groups.go:49` |
| Get Group | (internal) | `GET /groups/{id}` | `groups.go:80` |
| Update Group | `peer --edit --add/remove-group` | `PUT /groups/{id}` | `groups.go:102` |
| List Networks | `network --list` | `GET /networks` | `networks.go:15` |
| List Policies | `policy --list` | `GET /policies` | `policies.go:181` |
| **Create Policy** | `policy --create` | `POST /policies` | `policies.go:315` |
| **Inspect Policy** | `policy --inspect <id>` | `GET /policies/{id}` | `policies.go:252` |
| **Update Policy** | `policy --enable/--disable` | `PUT /policies/{id}` | `policies.go:359` |
| **Delete Policy** | `policy --delete <id>` | `DELETE /policies/{id}` | `policies.go:347` |
| **Add Rule** | `policy --add-rule` | `PUT /policies/{id}` | `policies.go:404` |
| **Edit Rule** | `policy --edit-rule` | `PUT /policies/{id}` | `policies.go:451` |
| **Remove Rule** | `policy --remove-rule` | `PUT /policies/{id}` | `policies.go:544` |

### 📋 Planned Endpoints

See [`docs/api/resources/README.md`](docs/api/resources/README.md) for full catalog of available endpoints.

**High Priority:**
- ✅ Full CRUD for Groups (create, delete) - **COMPLETE**
- ✅ Full CRUD for Networks (create, update, delete) - **COMPLETE**
- ✅ Full CRUD for Policies (create, update, delete) - **COMPLETE**
- User Management (list, create, delete)
- Setup Keys (create, list, delete)

**See README.md** for complete feature roadmap.

---

## Common API Patterns

### Listing Resources

```bash
curl -X GET https://api.netbird.io/api/{resource} \
  -H 'Authorization: Token <TOKEN>' \
  -H 'Accept: application/json'
```

**Example: List Peers**
```bash
curl -X GET https://api.netbird.io/api/peers \
  -H 'Authorization: Token nb_pat_abc123' \
  -H 'Accept: application/json'
```

### Creating Resources

```bash
curl -X POST https://api.netbird.io/api/{resource} \
  -H 'Authorization: Token <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{ "key": "value" }'
```

**Example: Create Group**
```bash
curl -X POST https://api.netbird.io/api/groups \
  -H 'Authorization: Token nb_pat_abc123' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "developers",
    "peers": [],
    "resources": []
  }'
```

### Updating Resources

**Important:** PUT requires the complete object. To modify a resource:
1. GET the current state
2. Modify the data
3. PUT the complete updated object

**Example: Add Peer to Group**
```bash
# Step 1: Get current group state
curl -X GET https://api.netbird.io/api/groups/group-id \
  -H 'Authorization: Token <TOKEN>'

# Step 2: Modify peer list locally (add new peer ID)

# Step 3: Send complete updated group
curl -X PUT https://api.netbird.io/api/groups/group-id \
  -H 'Authorization: Token <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "developers",
    "peers": ["peer-1", "peer-2", "new-peer-3"],
    "resources": []
  }'
```

See `peers.go:modifyPeerGroup()` for implementation example.

### Deleting Resources

```bash
curl -X DELETE https://api.netbird.io/api/{resource}/{id} \
  -H 'Authorization: Token <TOKEN>'
```

**Example: Delete Peer**
```bash
curl -X DELETE https://api.netbird.io/api/peers/peer-id \
  -H 'Authorization: Token nb_pat_abc123'
```

---

## Quick Endpoint Reference

### Peers

| Method | Endpoint | Purpose | CLI Support |
|--------|----------|---------|-------------|
| GET | `/peers` | List all peers | ✅ `peer --list` |
| GET | `/peers/{id}` | Get peer details | ✅ `peer --inspect` |
| PUT | `/peers/{id}` | Update peer settings | ❌ Not implemented |
| DELETE | `/peers/{id}` | Remove peer | ✅ `peer --remove` |
| GET | `/peers/{id}/accessible-peers` | List accessible peers | ❌ Not implemented |

**Docs:** [`docs/api/resources/peers.md`](docs/api/resources/peers.md)

### Groups

| Method | Endpoint | Purpose | CLI Support |
|--------|----------|---------|-------------|
| GET | `/groups` | List all groups | ✅ `group` |
| POST | `/groups` | Create group | ❌ Planned |
| GET | `/groups/{id}` | Get group details | ✅ Internal |
| PUT | `/groups/{id}` | Update group | ✅ `peer --edit --add/remove-group` |
| DELETE | `/groups/{id}` | Delete group | ❌ Planned |

**Docs:** [`docs/api/resources/groups.md`](docs/api/resources/groups.md)

### Networks

| Method | Endpoint | Purpose | CLI Support |
|--------|----------|---------|-------------|
| GET | `/networks` | List all networks | ✅ `networks` |
| POST | `/networks` | Create network | ❌ Planned |
| GET | `/networks/{id}` | Get network details | ❌ Planned |
| PUT | `/networks/{id}` | Update network | ❌ Planned |
| DELETE | `/networks/{id}` | Delete network | ❌ Planned |

Plus 7 additional endpoints for network resources and routers.

**Docs:** [`docs/api/resources/networks.md`](docs/api/resources/networks.md)

### Policies

| Method | Endpoint | Purpose | CLI Support |
|--------|----------|---------|-------------|
| GET | `/policies` | List all policies | ✅ `policy --list` |
| POST | `/policies` | Create policy | ✅ `policy --create` |
| GET | `/policies/{id}` | Get policy details | ✅ `policy --inspect` |
| PUT | `/policies/{id}` | Update policy | ✅ `policy --enable/--disable/--add-rule/--edit-rule/--remove-rule` |
| DELETE | `/policies/{id}` | Delete policy | ✅ `policy --delete` |

**Features:**
- Full CRUD operations for policies
- Rule management (add, edit, remove)
- Protocol, port, and port range configuration
- Bidirectional traffic support
- Group name resolution (use friendly names instead of IDs)
- List filtering (enabled/disabled, name search)

**Docs:** [`docs/api/resources/policies.md`](docs/api/resources/policies.md)

### Users

| Method | Endpoint | Purpose | CLI Support |
|--------|----------|---------|-------------|
| GET | `/users` | List all users | ❌ Planned |
| POST | `/users` | Create/invite user | ❌ Planned |
| PUT | `/users/{id}` | Update user | ❌ Planned |
| DELETE | `/users/{id}` | Delete user | ❌ Planned |
| POST | `/users/{id}/invite` | Resend invitation | ❌ Planned |
| GET | `/users/current` | Get current user | ❌ Planned |

**Docs:** [`docs/api/resources/users.md`](docs/api/resources/users.md)

### Other Resources

- **Tokens:** [`docs/api/resources/tokens.md`](docs/api/resources/tokens.md)
- **Accounts:** [`docs/api/resources/accounts.md`](docs/api/resources/accounts.md)
- **DNS:** [`docs/api/resources/dns.md`](docs/api/resources/dns.md)
- **Routes:** [`docs/api/resources/routes.md`](docs/api/resources/routes.md)
- **Setup Keys:** [`docs/api/resources/setup-keys.md`](docs/api/resources/setup-keys.md)
- **Posture Checks:** [`docs/api/resources/posture-checks.md`](docs/api/resources/posture-checks.md)
- **Events:** [`docs/api/resources/events.md`](docs/api/resources/events.md)
- **Geo-Locations:** [`docs/api/resources/geo-locations.md`](docs/api/resources/geo-locations.md)
- **Ingress Ports:** [`docs/api/resources/ingress-ports.md`](docs/api/resources/ingress-ports.md) (Cloud only)

---

## Error Responses

All endpoints may return these error codes:

| Code | Status | Common Cause | Solution |
|------|--------|--------------|----------|
| 400 | Bad Request | Invalid parameters | Check request body and required fields |
| 401 | Unauthorized | Invalid token | Verify token is correct and not expired |
| 403 | Forbidden | Insufficient permissions | Check user has required permissions |
| 404 | Not Found | Resource doesn't exist | Verify resource ID is correct |
| 429 | Too Many Requests | Rate limit exceeded | Implement backoff and retry |
| 500 | Internal Server Error | Server error | Retry request, contact support if persistent |

**Detailed Error Guide:** [`docs/api/guides/errors.md`](docs/api/guides/errors.md)

---

## Configuration

### Environment Variables

```bash
# Store token securely
export NETBIRD_TOKEN="nb_pat_your_token_here"

# Use in CLI
netbird-manage connect --token "$NETBIRD_TOKEN"
```

### Config File

**Location:** `$HOME/.netbird-manage.json`

**Format:**
```json
{
  "token": "nb_pat_your_token_here",
  "management_url": "https://api.netbird.io/api"
}
```

**Permissions:** `0600` (owner read/write only)

---

## For AI Assistants

When working on this codebase:

1. **Comprehensive docs:** Start with [`docs/api/README.md`](docs/api/README.md)
2. **Implementation patterns:** See `CLAUDE.md` for code conventions
3. **API endpoint details:** Reference specific files in `docs/api/resources/`
4. **Examples:** All guides include working code examples

### CLI → API Mapping

```go
// peers.go
client.listPeers()       → GET /peers
client.getPeerByID(id)   → GET /peers/{id}
client.removePeerByID(id)→ DELETE /peers/{id}

// groups.go
client.listGroups()         → GET /groups
client.getGroupByID(id)     → GET /groups/{id}
client.getGroupByName(name) → GET /groups + filter
client.updateGroup(id, req) → PUT /groups/{id}

// networks.go
client.listNetworks()    → GET /networks

// policies.go
client.listPolicies()    → GET /policies
```

---

## External Resources

- **Official API Docs:** https://docs.netbird.io/api
- **OpenAPI Spec:** https://api.netbird.io/api/openapi.json
- **NetBird GitHub:** https://github.com/netbirdio/netbird
- **Source Docs:** https://github.com/netbirdio/docs/tree/main/src/pages/ipa

---

**Last Updated:** 2025-11-15
**API Version:** v1 (Beta)
**CLI Version:** netbird-manage v0.1

For detailed documentation, explore the [`docs/api/`](docs/api/) directory.
