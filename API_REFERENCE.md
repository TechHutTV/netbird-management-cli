# NetBird API Reference

This document provides a quick reference for the NetBird Management API endpoints used by this CLI tool.

**Base URL:** `https://api.netbird.io/api` (default cloud)
**Authentication:** Bearer token via `Authorization: Token <token>` header
**Documentation:** https://docs.netbird.io/api

---

## Authentication

### Headers Required
```
Authorization: Token <your-api-token>
Accept: application/json
Content-Type: application/json (for POST/PUT requests)
```

### Getting a Token
1. Log into your NetBird dashboard
2. Navigate to Settings → Access Tokens
3. Create a Personal Access Token or Service User token
4. Use with: `netbird-manage connect --token <token>`

---

## Accounts

### GET /accounts
**Purpose:** Retrieve account information
**Response:** Account details including ID, settings, and metadata
**Used by:** N/A (not currently implemented in CLI)

---

## Users

### GET /users
**Purpose:** List all users in the account
**Response:** Array of user objects
**Used by:** N/A (not currently implemented in CLI)

### POST /users
**Purpose:** Create a new user (invite)
**Request Body:**
```json
{
  "email": "user@example.com",
  "name": "User Name",
  "role": "user"
}
```
**Used by:** N/A (planned feature)

---

## Peers

### GET /peers
**Purpose:** List all peers in the network
**Response:**
```json
[
  {
    "id": "peer-uuid",
    "name": "peer-name",
    "ip": "100.64.0.1",
    "connected": true,
    "last_seen": "2024-01-15T10:30:00Z",
    "os": "linux",
    "version": "0.24.0",
    "groups": [
      {"id": "group-uuid", "name": "group-name"}
    ],
    "hostname": "hostname"
  }
]
```
**Used by:** `netbird-manage peer --list`
**Implementation:** `client.go:listPeers()`, `peers.go:handlePeersCommand()`

### GET /peers/{peerId}
**Purpose:** Get detailed information for a specific peer
**Response:** Single peer object (same structure as above)
**Used by:** `netbird-manage peer --inspect <peer-id>`
**Implementation:** `client.go:getPeerByID()`, `peers.go:inspectPeer()`

### DELETE /peers/{peerId}
**Purpose:** Remove a peer from the network
**Response:** 200 OK on success
**Used by:** `netbird-manage peer --remove <peer-id>`
**Implementation:** `client.go:removePeerByID()`, `peers.go:handlePeersCommand()`

**Note:** Peers cannot be directly updated via API. To modify peer groups, you must update the group membership instead.

---

## Groups

### GET /groups
**Purpose:** List all groups
**Response:**
```json
[
  {
    "id": "group-uuid",
    "name": "group-name",
    "peers_count": 5,
    "resources_count": 2,
    "issued": "api"
  }
]
```
**Used by:** `netbird-manage group`
**Implementation:** `groups.go:listGroups()`

### GET /groups/{groupId}
**Purpose:** Get full group details including members
**Response:**
```json
{
  "id": "group-uuid",
  "name": "group-name",
  "peers_count": 5,
  "resources_count": 2,
  "issued": "api",
  "peers": [
    {
      "id": "peer-uuid",
      "name": "peer-name",
      "ip": "100.64.0.1",
      "connected": true,
      "last_seen": "2024-01-15T10:30:00Z",
      "os": "linux",
      "version": "0.24.0",
      "groups": [],
      "hostname": "hostname"
    }
  ],
  "resources": [
    {
      "id": "resource-uuid",
      "type": "host"
    }
  ]
}
```
**Used by:** `netbird-manage peer --edit <id> --add-group <name>`
**Implementation:** `groups.go:getGroupByID()`, `groups.go:getGroupByName()`

### PUT /groups/{groupId}
**Purpose:** Update group (modify members or resources)
**Request Body:**
```json
{
  "name": "group-name",
  "peers": ["peer-uuid-1", "peer-uuid-2"],
  "resources": [
    {
      "id": "resource-uuid",
      "type": "host"
    }
  ]
}
```
**Response:** Updated group object
**Used by:** `netbird-manage peer --edit <id> --add-group <name>`, `--remove-group <name>`
**Implementation:** `groups.go:updateGroup()`, `peers.go:modifyPeerGroup()`

**Important:** PUT requires the complete group object. To add/remove a peer:
1. GET /groups/{groupId} to fetch current state
2. Modify the peers array
3. PUT /groups/{groupId} with updated data

### POST /groups
**Purpose:** Create a new group
**Request Body:**
```json
{
  "name": "new-group-name",
  "peers": [],
  "resources": []
}
```
**Used by:** N/A (planned feature)

### DELETE /groups/{groupId}
**Purpose:** Delete a group
**Response:** 200 OK on success
**Used by:** N/A (planned feature)

---

## Networks

### GET /networks
**Purpose:** List all networks
**Response:**
```json
[
  {
    "id": "network-uuid",
    "name": "network-name",
    "description": "Network description",
    "routers": ["peer-uuid-1", "peer-uuid-2"],
    "routing_peers_count": 2,
    "resources": ["resource-uuid"],
    "policies": ["policy-uuid"]
  }
]
```
**Used by:** `netbird-manage networks`
**Implementation:** `networks.go:listNetworks()`

### GET /networks/{networkId}
**Purpose:** Get detailed network information
**Response:** Single network object
**Used by:** N/A (not currently implemented)

### POST /networks
**Purpose:** Create a new network
**Request Body:**
```json
{
  "name": "network-name",
  "description": "Description",
  "routers": [],
  "resources": []
}
```
**Used by:** N/A (planned feature)

### PUT /networks/{networkId}
**Purpose:** Update network configuration
**Used by:** N/A (planned feature)

### DELETE /networks/{networkId}
**Purpose:** Delete a network
**Used by:** N/A (planned feature)

---

## Policies

### GET /policies
**Purpose:** List all access control policies
**Response:**
```json
[
  {
    "id": "policy-uuid",
    "name": "policy-name",
    "description": "Policy description",
    "enabled": true,
    "rules": [
      {
        "id": "rule-uuid",
        "name": "rule-name",
        "enabled": true,
        "action": "accept",
        "protocol": "tcp",
        "sources": [
          {"id": "group-uuid", "name": "source-group"}
        ],
        "destinations": [
          {"id": "group-uuid", "name": "dest-group"}
        ]
      }
    ]
  }
]
```
**Used by:** `netbird-manage policy`
**Implementation:** `policies.go:listPolicies()`

**Rule Actions:**
- `accept` - Allow traffic
- `drop` - Block traffic

**Protocols:** `tcp`, `udp`, `icmp`, `all`

### GET /policies/{policyId}
**Purpose:** Get detailed policy information
**Response:** Single policy object
**Used by:** N/A (not currently implemented)

### POST /policies
**Purpose:** Create a new policy
**Request Body:**
```json
{
  "name": "policy-name",
  "description": "Description",
  "enabled": true,
  "rules": [
    {
      "name": "rule-name",
      "enabled": true,
      "action": "accept",
      "protocol": "tcp",
      "sources": ["group-uuid"],
      "destinations": ["group-uuid"]
    }
  ]
}
```
**Used by:** N/A (planned feature)

### PUT /policies/{policyId}
**Purpose:** Update policy configuration
**Used by:** N/A (planned feature)

### DELETE /policies/{policyId}
**Purpose:** Delete a policy
**Used by:** N/A (planned feature)

---

## Setup Keys

### GET /setup-keys
**Purpose:** List all setup keys for device onboarding
**Response:** Array of setup key objects
**Used by:** N/A (not currently implemented)

### POST /setup-keys
**Purpose:** Create a new setup key
**Request Body:**
```json
{
  "name": "key-name",
  "type": "reusable",
  "expires_in": 86400,
  "auto_groups": ["group-uuid"]
}
```
**Used by:** N/A (planned feature)

---

## DNS

### GET /dns/nameservers
**Purpose:** List DNS nameserver groups
**Response:** Array of nameserver group objects
**Used by:** N/A (not currently implemented)

### GET /dns/settings
**Purpose:** Get DNS settings
**Response:** DNS configuration object
**Used by:** N/A (not currently implemented)

---

## Events

### GET /events
**Purpose:** Retrieve audit logs and activity records
**Query Parameters:**
- `?limit=100` - Limit number of results
- `?offset=0` - Pagination offset
**Response:** Array of event objects
**Used by:** N/A (not currently implemented)

---

## Posture Checks

### GET /posture-checks
**Purpose:** List device compliance checks
**Response:** Array of posture check objects
**Used by:** N/A (not currently implemented)

### POST /posture-checks
**Purpose:** Create a new posture check
**Used by:** N/A (planned feature)

---

## Error Responses

All endpoints may return these error codes:

### 400 Bad Request
```json
{
  "message": "Invalid request parameters",
  "code": 400
}
```

### 401 Unauthorized
```json
{
  "message": "Invalid or missing authentication token",
  "code": 401
}
```

### 403 Forbidden
```json
{
  "message": "Insufficient permissions",
  "code": 403
}
```

### 404 Not Found
```json
{
  "message": "Resource not found",
  "code": 404
}
```

### 500 Internal Server Error
```json
{
  "message": "Internal server error",
  "code": 500
}
```

---

## Rate Limiting

The NetBird API implements rate limiting to prevent abuse. If you exceed the rate limit, you'll receive a 429 Too Many Requests response.

**Best Practices:**
- Implement exponential backoff for retries
- Cache responses when appropriate
- Batch operations when possible

---

## API Versioning

The current API version is included in the base URL. Breaking changes will be released under a new version path.

**Current Version:** v1 (implicit in `/api` path)

---

## Common Patterns

### Adding a Peer to a Group

This is the most common operation in the CLI. The pattern is:

1. **Fetch the group details:**
   ```
   GET /groups/{groupId}
   ```

2. **Modify the peers array:**
   ```go
   newPeers := append(group.Peers, newPeerID)
   ```

3. **Update the group:**
   ```
   PUT /groups/{groupId}
   {
     "name": "group-name",
     "peers": ["peer-1", "peer-2", "new-peer"],
     "resources": [...]
   }
   ```

**Implementation Reference:** `peers.go:modifyPeerGroup()`

### Removing a Peer from a Group

Same as above, but filter out the peer ID from the array:

```go
var newPeers []string
for _, p := range group.Peers {
    if p.ID != peerToRemove {
        newPeers = append(newPeers, p.ID)
    }
}
```

**Implementation Reference:** `peers.go:modifyPeerGroup()`

---

## Self-Hosted API URL

For self-hosted NetBird instances, use the `--management-url` flag:

```bash
netbird-manage connect --token <token> --management-url https://your-server.com/api
```

The management URL should point to your NetBird management server's API endpoint.

---

## Authentication Flow

1. **User generates token** in NetBird dashboard
2. **User connects:** `netbird-manage connect --token <token>`
3. **CLI validates token:** Makes test request to `GET /peers`
4. **CLI saves config:** Stores token in `~/.netbird-manage.json` with `0600` permissions
5. **Subsequent requests:** Load token from config, add to `Authorization` header

**Implementation Reference:** `config.go:testAndSaveConfig()`, `client.go:makeRequest()`

---

## Future API Endpoints (Planned)

The following endpoints exist in the NetBird API but are not yet implemented in this CLI:

- [ ] Full CRUD for Groups (create, delete)
- [ ] Full CRUD for Networks (create, update, delete)
- [ ] Full CRUD for Policies (create, update, delete)
- [ ] User Management (invite, remove, update roles)
- [ ] Setup Keys (create, list, delete)
- [ ] Routes (create, update, delete)
- [ ] DNS Management (nameservers, settings)
- [ ] Events/Audit Logs (list, filter)
- [ ] Posture Checks (create, update, delete)

See `README.md` for the full roadmap.

---

## Additional Resources

- **Official API Documentation:** https://docs.netbird.io/api
- **NetBird GitHub:** https://github.com/netbirdio/netbird
- **NetBird Website:** https://netbird.io/
- **OpenAPI Spec:** Available at your management server `/api/openapi.json`

---

**Last Updated:** 2025-11-15
**API Version:** v1
**CLI Version:** Corresponds to codebase at commit 5033e93
