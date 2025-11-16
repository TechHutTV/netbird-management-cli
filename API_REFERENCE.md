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

#### Peers
| CLI Command | API Endpoint | Implementation File |
|-------------|--------------|---------------------|
| `peer --list` | `GET /peers` | `peers.go` |
| `peer --inspect <id>` | `GET /peers/{id}` | `peers.go` |
| `peer --update <id>` | `PUT /peers/{id}` | `peers.go` |
| `peer --remove <id>` | `DELETE /peers/{id}` | `peers.go` |
| `peer --accessible-peers <id>` | `GET /peers/{id}/accessible-peers` | `peers.go` |
| `peer --edit --add/remove-group` | `PUT /groups/{id}` | `groups.go` |

#### Groups
| CLI Command | API Endpoint | Implementation File |
|-------------|--------------|---------------------|
| `group --list` | `GET /groups` | `groups.go` |
| `group --inspect <id>` | `GET /groups/{id}` | `groups.go` |
| `group --create` | `POST /groups` | `groups.go` |
| `group --delete <id>` | `DELETE /groups/{id}` | `groups.go` |
| `group --rename <id>` | `PUT /groups/{id}` | `groups.go` |
| `group --add-peers/--remove-peers` | `PUT /groups/{id}` | `groups.go` |

#### Networks
| CLI Command | API Endpoint | Implementation File |
|-------------|--------------|---------------------|
| `network --list` | `GET /networks` | `networks.go` |
| `network --inspect <id>` | `GET /networks/{id}` | `networks.go` |
| `network --create` | `POST /networks` | `networks.go` |
| `network --delete <id>` | `DELETE /networks/{id}` | `networks.go` |
| `network --rename/--update` | `PUT /networks/{id}` | `networks.go` |
| `network --add/update/remove-resource` | Network resource endpoints | `networks.go` |
| `network --add/update/remove-router` | Network router endpoints | `networks.go` |
| `network --list-routers/--list-all-routers` | Router listing endpoints | `networks.go` |

#### Policies
| CLI Command | API Endpoint | Implementation File |
|-------------|--------------|---------------------|
| `policy --list` | `GET /policies` | `policies.go` |
| `policy --inspect <id>` | `GET /policies/{id}` | `policies.go` |
| `policy --create` | `POST /policies` | `policies.go` |
| `policy --delete <id>` | `DELETE /policies/{id}` | `policies.go` |
| `policy --enable/--disable` | `PUT /policies/{id}` | `policies.go` |
| `policy --add-rule` | `PUT /policies/{id}` | `policies.go` |
| `policy --edit-rule` | `PUT /policies/{id}` | `policies.go` |
| `policy --remove-rule` | `PUT /policies/{id}` | `policies.go` |

#### Setup Keys
| CLI Command | API Endpoint | Implementation File |
|-------------|--------------|---------------------|
| `setup-key --list` | `GET /setup-keys` | `setup-keys.go` |
| `setup-key --inspect <id>` | `GET /setup-keys/{id}` | `setup-keys.go` |
| `setup-key --create/--quick` | `POST /setup-keys` | `setup-keys.go` |
| `setup-key --delete <id>` | `DELETE /setup-keys/{id}` | `setup-keys.go` |
| `setup-key --revoke/--enable` | `PUT /setup-keys/{id}` | `setup-keys.go` |
| `setup-key --update-groups` | `PUT /setup-keys/{id}` | `setup-keys.go` |

#### Users
| CLI Command | API Endpoint | Implementation File |
|-------------|--------------|---------------------|
| `user --list` | `GET /users` | `users.go` |
| `user --me` | `GET /users/current` | `users.go` |
| `user --invite` | `POST /users` | `users.go` |
| `user --update <id>` | `PUT /users/{id}` | `users.go` |
| `user --remove <id>` | `DELETE /users/{id}` | `users.go` |
| `user --resend-invite <id>` | `POST /users/{id}/invite` | `users.go` |

#### Tokens
| CLI Command | API Endpoint | Implementation File |
|-------------|--------------|---------------------|
| `token --list` | `GET /users/{userId}/tokens` | `tokens.go` |
| `token --inspect <id>` | `GET /users/{userId}/tokens/{id}` | `tokens.go` |
| `token --create` | `POST /users/{userId}/tokens` | `tokens.go` |
| `token --revoke <id>` | `DELETE /users/{userId}/tokens/{id}` | `tokens.go` |

#### Routes
| CLI Command | API Endpoint | Implementation File |
|-------------|--------------|---------------------|
| `route --list` | `GET /routes` | `routes.go` |
| `route --inspect <id>` | `GET /routes/{id}` | `routes.go` |
| `route --create` | `POST /routes` | `routes.go` |
| `route --update <id>` | `PUT /routes/{id}` | `routes.go` |
| `route --delete <id>` | `DELETE /routes/{id}` | `routes.go` |
| `route --enable/--disable <id>` | `PUT /routes/{id}` | `routes.go` |

#### DNS
| CLI Command | API Endpoint | Implementation File |
|-------------|--------------|---------------------|
| `dns --list` | `GET /dns/nameservers` | `dns.go` |
| `dns --inspect <id>` | `GET /dns/nameservers/{id}` | `dns.go` |
| `dns --create` | `POST /dns/nameservers` | `dns.go` |
| `dns --update <id>` | `PUT /dns/nameservers/{id}` | `dns.go` |
| `dns --delete <id>` | `DELETE /dns/nameservers/{id}` | `dns.go` |
| `dns --enable/--disable <id>` | `PUT /dns/nameservers/{id}` | `dns.go` |
| `dns --get-settings` | `GET /dns/settings` | `dns.go` |
| `dns --update-settings` | `PUT /dns/settings` | `dns.go` |

#### Posture Checks
| CLI Command | API Endpoint | Implementation File |
|-------------|--------------|---------------------|
| `posture-check --list` | `GET /posture-checks` | `posture-checks.go` |
| `posture-check --inspect <id>` | `GET /posture-checks/{id}` | `posture-checks.go` |
| `posture-check --create` | `POST /posture-checks` | `posture-checks.go` |
| `posture-check --update <id>` | `PUT /posture-checks/{id}` | `posture-checks.go` |
| `posture-check --delete <id>` | `DELETE /posture-checks/{id}` | `posture-checks.go` |

### 📋 Not Yet Implemented

**Monitoring & Analytics:**
- Events (audit logs and activity monitoring)
- Geo-Locations (location data for access control)

**Account Management:**
- Accounts (account settings and configuration)

**Cloud-Only Features:**
- Ingress Ports (port forwarding - NetBird Cloud only)

**See README.md** for complete feature roadmap and implementation details.

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
| PUT | `/peers/{id}` | Update peer settings | ✅ `peer --update` |
| DELETE | `/peers/{id}` | Remove peer | ✅ `peer --remove` |
| GET | `/peers/{id}/accessible-peers` | List accessible peers | ✅ `peer --accessible-peers` |

**Docs:** [`docs/api/resources/peers.md`](docs/api/resources/peers.md)

### Groups

| Method | Endpoint | Purpose | CLI Support |
|--------|----------|---------|-------------|
| GET | `/groups` | List all groups | ✅ `group --list` |
| POST | `/groups` | Create group | ✅ `group --create` |
| GET | `/groups/{id}` | Get group details | ✅ `group --inspect` |
| PUT | `/groups/{id}` | Update group | ✅ `group --rename`, `group --add/remove-peers`, `peer --edit` |
| DELETE | `/groups/{id}` | Delete group | ✅ `group --delete` |

**Docs:** [`docs/api/resources/groups.md`](docs/api/resources/groups.md)

### Networks

| Method | Endpoint | Purpose | CLI Support |
|--------|----------|---------|-------------|
| GET | `/networks` | List all networks | ✅ `network --list` |
| POST | `/networks` | Create network | ✅ `network --create` |
| GET | `/networks/{id}` | Get network details | ✅ `network --inspect` |
| PUT | `/networks/{id}` | Update network | ✅ `network --rename`, `network --update` |
| DELETE | `/networks/{id}` | Delete network | ✅ `network --delete` |

Plus 7 additional endpoints for network resources and routers - **ALL FULLY IMPLEMENTED** via `network --add/update/remove-resource` and `network --add/update/remove-router` commands.

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
| GET | `/users` | List all users | ✅ `user --list` |
| POST | `/users` | Create/invite user | ✅ `user --invite` |
| PUT | `/users/{id}` | Update user | ✅ `user --update` |
| DELETE | `/users/{id}` | Delete user | ✅ `user --remove` |
| POST | `/users/{id}/invite` | Resend invitation | ✅ `user --resend-invite` |
| GET | `/users/current` | Get current user | ✅ `user --me` |

**Docs:** [`docs/api/resources/users.md`](docs/api/resources/users.md)

### Other Resources

**✅ Fully Implemented:**
- **Tokens** (4 endpoints) - [`docs/api/resources/tokens.md`](docs/api/resources/tokens.md) - See `token` commands
- **DNS** (6 endpoints) - [`docs/api/resources/dns.md`](docs/api/resources/dns.md) - See `dns` commands
- **Routes** (5 endpoints) - [`docs/api/resources/routes.md`](docs/api/resources/routes.md) - See `route` commands
- **Setup Keys** (5 endpoints) - [`docs/api/resources/setup-keys.md`](docs/api/resources/setup-keys.md) - See `setup-key` commands
- **Posture Checks** (5 endpoints) - [`docs/api/resources/posture-checks.md`](docs/api/resources/posture-checks.md) - See `posture-check` commands

**❌ Not Yet Implemented:**
- **Accounts** (3 endpoints) - [`docs/api/resources/accounts.md`](docs/api/resources/accounts.md) - Account settings
- **Events** (2 endpoints) - [`docs/api/resources/events.md`](docs/api/resources/events.md) - Audit logs and monitoring
- **Geo-Locations** (2 endpoints) - [`docs/api/resources/geo-locations.md`](docs/api/resources/geo-locations.md) - Location data
- **Ingress Ports** (10 endpoints) - [`docs/api/resources/ingress-ports.md`](docs/api/resources/ingress-ports.md) - Port forwarding (Cloud only)

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
