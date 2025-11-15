# NetBird API - Authentication Guide

## Overview

NetBird provides two distinct authentication mechanisms for API requests:
1. **OAuth2 with Bearer Tokens**
2. **Personal Access Tokens (PATs)**

Both methods are secure and suitable for different use cases. Choose the method that best fits your workflow.

---

## Method 1: OAuth2 Authentication

OAuth2 authentication uses bearer tokens from your identity provider (IDP).

### When to Use OAuth2

Use OAuth2 when:
- You're building applications that integrate with existing identity providers
- You need user-context authentication
- You want to leverage existing SSO infrastructure
- You require short-lived, revocable tokens

### Getting an OAuth2 Token

Retrieve your access token from your IDP manager. The exact process depends on your identity provider (Auth0, Okta, Azure AD, etc.).

### Using OAuth2 Tokens

Include the token in the `Authorization` header with the `Bearer` prefix:

```bash
curl https://api.netbird.io/api/users \
  -H "Authorization: Bearer {token}" \
  -H "Accept: application/json"
```

**Example with actual request:**
```bash
curl -X GET https://api.netbird.io/api/peers \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Accept: application/json"
```

---

## Method 2: Personal Access Tokens (PATs)

Personal Access Tokens are API-specific tokens generated directly in the NetBird dashboard.

### When to Use PATs

Use Personal Access Tokens when:
- You're writing scripts or automation tools
- You need long-lived tokens for continuous access
- You're building CLI tools or integrations
- You don't have an external identity provider

### Creating a Personal Access Token

1. Log into the NetBird dashboard at [app.netbird.io](https://app.netbird.io)
2. Navigate to **Users** → **Me** (or your user settings)
3. Find the **Personal Access Tokens** section
4. Click **Create Token**
5. Provide a name and expiration period
6. Copy the token immediately (it won't be shown again)

### Using Personal Access Tokens

Include the token in the `Authorization` header with the `Token` prefix:

```bash
curl https://api.netbird.io/api/users \
  -H "Authorization: Token {token}" \
  -H "Accept: application/json"
```

**Example with actual request:**
```bash
curl -X GET https://api.netbird.io/api/peers \
  -H "Authorization: Token nb_pat_1234567890abcdef" \
  -H "Accept: application/json"
```

**Note:** The format is `Token` (not `Bearer`) for PATs.

---

## Service Users for Organization-Wide Operations

For automation and organization-level API access, NetBird recommends using **service users**.

### What are Service Users?

Service users are:
- Non-human user accounts designed for API access
- Not tied to a specific person
- Ideal for CI/CD pipelines, automation scripts, and integrations
- Can be managed independently of individual user accounts

### Creating a Service User

1. Navigate to **Users** in the NetBird dashboard
2. Click **Add User**
3. Select **Service User** as the user type
4. Configure permissions and auto-groups
5. Generate a Personal Access Token for the service user

### Best Practices for Service Users

- Create separate service users for different applications/purposes
- Use descriptive names (e.g., `ci-pipeline-bot`, `monitoring-service`)
- Assign minimal required permissions (principle of least privilege)
- Rotate tokens regularly
- Monitor service user activity in audit logs

---

## Security Best Practices

### Token Storage

**DO:**
- ✓ Store tokens in environment variables
- ✓ Use secure secret management systems (Vault, AWS Secrets Manager, etc.)
- ✓ Encrypt tokens at rest
- ✓ Use secure file permissions (0600) for config files

**DON'T:**
- ✗ Hardcode tokens in source code
- ✗ Commit tokens to version control
- ✗ Share tokens via email or messaging
- ✗ Log tokens in application logs

### Token Rotation

- Rotate tokens regularly (recommended: every 90 days)
- Immediately revoke tokens if:
  - You suspect compromise
  - An employee leaves
  - A service is decommissioned
  - Tokens appear in logs or version control

### Token Revocation

To revoke a token:
1. Go to NetBird dashboard → Users → [User] → Personal Access Tokens
2. Find the token to revoke
3. Click **Delete** or **Revoke**
4. Confirm the action

**Important:** Always keep your token safe and reset it if you suspect it has been compromised.

---

## Authentication Headers

### Required Headers

All authenticated requests must include:

```
Authorization: Token <YOUR_TOKEN>  (for PATs)
             or
Authorization: Bearer <YOUR_TOKEN> (for OAuth2)
```

### Recommended Headers

For best compatibility, also include:

```
Accept: application/json
Content-Type: application/json  (for POST/PUT requests)
```

### Complete Example

```bash
curl -X GET https://api.netbird.io/api/peers \
  -H "Authorization: Token nb_pat_abc123xyz" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json"
```

---

## Testing Authentication

### Verify Token Validity

Test your token by making a simple API request:

```bash
# For Personal Access Tokens
curl -X GET https://api.netbird.io/api/peers \
  -H "Authorization: Token YOUR_TOKEN" \
  -H "Accept: application/json"

# Expected: 200 OK with list of peers
# If invalid: 401 Unauthorized
```

### Common Authentication Errors

| Status Code | Error | Solution |
|-------------|-------|----------|
| 401 Unauthorized | Invalid or missing token | Verify token is correct and not expired |
| 403 Forbidden | Insufficient permissions | Check user/service user has required permissions |
| 429 Too Many Requests | Rate limit exceeded | Implement backoff and retry logic |

---

## Example: CLI Tool Authentication

Here's how to implement authentication in a CLI tool:

### Using Environment Variables

```bash
#!/bin/bash

# Set token as environment variable
export NETBIRD_TOKEN="nb_pat_your_token_here"

# Use in requests
curl -X GET https://api.netbird.io/api/peers \
  -H "Authorization: Token $NETBIRD_TOKEN" \
  -H "Accept: application/json"
```

### Using Config File

```bash
# Store in ~/.netbird-cli.conf (permissions: 0600)
{
  "token": "nb_pat_your_token_here",
  "api_url": "https://api.netbird.io/api"
}
```

```bash
# Load and use
TOKEN=$(jq -r '.token' ~/.netbird-cli.conf)

curl -X GET https://api.netbird.io/api/peers \
  -H "Authorization: Token $TOKEN" \
  -H "Accept: application/json"
```

---

## Self-Hosted Authentication

For self-hosted NetBird installations:

### Custom Management URL

Point to your self-hosted management server:

```bash
curl -X GET https://your-netbird-server.com/api/peers \
  -H "Authorization: Token YOUR_TOKEN" \
  -H "Accept: application/json"
```

### Port Configuration

Self-hosted installations may use port 33073:

```bash
curl -X GET https://your-netbird-server.com:33073/api/peers \
  -H "Authorization: Token YOUR_TOKEN" \
  -H "Accept: application/json"
```

---

## Next Steps

- **[Quickstart Guide](quickstart.md)** - Make your first API request
- **[Error Handling](errors.md)** - Understanding API errors
- **[Users API](../resources/users.md)** - Manage users and permissions
- **[Tokens API](../resources/tokens.md)** - Programmatic token management

---

## Additional Resources

- **Creating Access Tokens:** https://docs.netbird.io/how-to/access-netbird-public-api
- **Managing PATs:** Dashboard at app.netbird.io/users
- **Service Users Guide:** https://docs.netbird.io/how-to/access-netbird-public-api#service-users
