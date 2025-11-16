# NetBird API Resources

This directory contains detailed documentation for all NetBird API endpoints, organized by resource type.

## Core Resources

### Network Management
- **[Peers](peers.md)** - Manage network peer devices
- **[Groups](groups.md)** - Organize peers and resources into groups
- **[Networks](networks.md)** - Network infrastructure and routing configuration
- **[Policies](policies.md)** - Access control rules and firewall policies

### User & Access Management
- **[Users](users.md)** - User account management and invitations
- **[Tokens](tokens.md)** - Personal access token management
- **[Accounts](accounts.md)** - Account settings and configuration

### Network Services
- **[DNS](dns.md)** - DNS nameserver groups and settings
- **[Routes](routes.md)** - Network routing configuration
- **[Ingress Ports](ingress-ports.md)** - Port forwarding and ingress peers (Cloud only)

### Security & Monitoring
- **[Posture Checks](posture-checks.md)** - Device compliance validation
- **[Events](events.md)** - Audit logs and network traffic monitoring
- **[Geo-Locations](geo-locations.md)** - Location-based data for access control

### Onboarding
- **[Setup Keys](setup-keys.md)** - Device registration and onboarding keys

## Quick Reference

| Resource | Endpoints | CLI Support |
|----------|-----------|-------------|
| Peers | 5 endpoints | ✅ **Full** (list, get, update, delete, accessible-peers) |
| Groups | 5 endpoints | ✅ **Full** (list, get, create, update, delete) |
| Networks | 12 endpoints | ✅ **Full** (CRUD + resources + routers) |
| Policies | 5 endpoints | ✅ **Full** (CRUD + rule management) |
| Setup Keys | 5 endpoints | ✅ **Full** (list, get, create, update, delete, revoke) |
| Users | 6 endpoints | ✅ **Full** (list, me, invite, update, remove, resend) |
| Tokens | 4 endpoints | ✅ **Full** (list, get, create, revoke) |
| Routes | 5 endpoints | ✅ **Full** (list, get, create, update, delete) |
| DNS | 6 endpoints | ✅ **Full** (CRUD + settings) |
| Posture Checks | 5 endpoints | ✅ **Full** (list, get, create, update, delete) |
| Accounts | 3 endpoints | ❌ Not implemented |
| Events | 2 endpoints | ❌ Not implemented |
| Geo-Locations | 2 endpoints | ❌ Not implemented |
| Ingress Ports | 10 endpoints | ❌ Not implemented |

**Total Coverage: 10/14 resource types (71%) - 58 endpoints fully implemented**

## Authentication

All API endpoints require authentication via the `Authorization` header:

**Personal Access Token:**
```
Authorization: Token <YOUR_TOKEN>
```

**OAuth2 Bearer Token:**
```
Authorization: Bearer <YOUR_TOKEN>
```

## Common Patterns

### List Resources
```bash
GET /api/{resource}
```

### Create Resource
```bash
POST /api/{resource}
Content-Type: application/json

{request body}
```

### Get Resource Details
```bash
GET /api/{resource}/{id}
```

### Update Resource
```bash
PUT /api/{resource}/{id}
Content-Type: application/json

{request body}
```

### Delete Resource
```bash
DELETE /api/{resource}/{id}
```

## Response Formats

All responses use JSON format with consistent schemas per resource type.

**Success (2xx):**
- 200 OK - Request successful, data in response body
- 201 Created - Resource created successfully
- 204 No Content - Request successful, no response body

**Errors (4xx/5xx):**
```json
{
  "message": "Error description",
  "code": 400,
  "details": "Additional context"
}
```

## Code Examples

Each resource documentation includes complete code examples in:
- cURL (bash)
- JavaScript (Axios)
- Python (requests)
- Go (net/http)
- Ruby
- Java
- PHP

## Next Steps

1. Choose a resource from the list above
2. Read the detailed endpoint documentation
3. Try the examples with your own NetBird token
4. Build your integration!

---

*For getting started, see the [Quickstart Guide](../guides/quickstart.md)*
