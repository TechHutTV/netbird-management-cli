# NetBird REST API - Introduction

## Overview

The NetBird Public API enables developers to manage users, peers, network rules and more from inside your application or scripts to automate the setup of your mesh network.

## Purpose

Use the NetBird API to:
- Manage network peers programmatically
- Configure access control policies
- Create and manage user accounts
- Set up network routes and DNS
- Monitor network activity through events
- Automate network setup and configuration

## Getting Started

### Prerequisites

Before using the NetBird API, you need to:

1. **Create a service user** for API communication
2. **Generate an authentication token** using one of these methods:
   - **Bearer token** from your identity provider (OAuth2)
   - **Personal access token** generated in the NetBird dashboard

### Base URL

**Cloud (Default):**
```
https://api.netbird.io/api
```

**Self-Hosted:**
```
https://your-server.com/api
```

*Note: Self-hosted installations may require port 33073 when connecting to the management server.*

## Authentication

All API requests must include authentication. NetBird supports two authentication methods:

### Method 1: OAuth2 Bearer Token
```bash
curl https://api.netbird.io/api/users \
  -H "Authorization: Bearer {token}"
```

### Method 2: Personal Access Token
```bash
curl https://api.netbird.io/api/users \
  -H "Authorization: Token {token}"
```

**Security Note:** Always keep your token safe and reset it if you suspect it has been compromised.

## Available Resources

The NetBird API provides endpoints for managing:

| Resource | Description |
|----------|-------------|
| **Accounts** | Account information and settings management |
| **Users** | User account creation, updates, and invitations |
| **Tokens** | Personal access token management |
| **Peers** | Network peer devices and their configurations |
| **Groups** | Organization of peers and resources |
| **Networks** | Network infrastructure and routing |
| **Policies** | Access control rules and firewall policies |
| **Setup Keys** | Device onboarding and registration |
| **DNS** | DNS nameserver groups and settings |
| **Routes** | Network routing configuration |
| **Events** | Audit logs and network traffic monitoring |
| **Posture Checks** | Device compliance and security validation |
| **Geo-Locations** | Location-based access control data |
| **Ingress Ports** | Port forwarding and ingress peer management (Cloud only) |

## Quick Example

Here's a simple example to get started - listing all peers in your network:

```bash
curl -X GET https://api.netbird.io/api/peers \
  -H 'Accept: application/json' \
  -H 'Authorization: Token <YOUR_TOKEN>'
```

Replace `<YOUR_TOKEN>` with your actual Personal Access Token.

## API Status

The NetBird API is currently in **Beta**. This means:
- Core functionality is stable and production-ready
- Some features may be experimental
- Error handling continues to be refined
- Breaking changes may occur with advance notice

## Next Steps

1. **[Quickstart Guide](guides/quickstart.md)** - Get up and running quickly
2. **[Authentication Details](guides/authentication.md)** - Learn about auth methods
3. **[Error Handling](guides/errors.md)** - Understand API responses
4. **[Resources Documentation](resources/)** - Explore all available endpoints

## Support

- **Documentation:** https://docs.netbird.io/api
- **GitHub Issues:** https://github.com/netbirdio/netbird/issues
- **Slack Community:** Contact via NetBird Slack workspace

## Rate Limiting

The NetBird API implements rate limiting to prevent abuse. Best practices:
- Implement exponential backoff for retries
- Cache responses when appropriate
- Batch operations when possible
- Handle 429 (Too Many Requests) responses gracefully

## API Versioning

The current API version is included in the base URL path. Breaking changes will be released under a new version with advance notice to developers.

---

*For detailed endpoint documentation, see the [Resources](resources/) section.*
