# NetBird API - Quickstart Guide

This guide will help you make your first NetBird API request in minutes.

## Prerequisites

Before you begin, ensure you have:
- **cURL installed** - Download from [curl.se](https://curl.se/) if needed
- **NetBird account** - Sign up at [app.netbird.io](https://app.netbird.io)
- **Access to NetBird dashboard** - To generate your token

---

## Step 1: Generate Your Access Token

1. Log into the NetBird dashboard at [app.netbird.io](https://app.netbird.io)
2. Navigate to **Users** → **Me** (or click your profile)
3. Scroll to **Personal Access Tokens** section
4. Click **Create Token** or **Add Token**
5. Provide details:
   - **Name:** "My First API Token" (or any descriptive name)
   - **Expiration:** Choose duration (1-365 days)
6. Click **Generate** or **Create**
7. **Copy the token immediately** - it won't be shown again!

**Example token format:**
```
nb_pat_1a2b3c4d5e6f7g8h9i0j
```

**Important:** Store this token securely. You'll need it for all API requests.

---

## Step 2: Test Your Connection

Make your first API request to list all peers in your network:

```bash
curl -X GET https://api.netbird.io/api/peers \
  -H 'Accept: application/json' \
  -H 'Authorization: Token <TOKEN>'
```

**Replace `<TOKEN>` with your actual token:**

```bash
curl -X GET https://api.netbird.io/api/peers \
  -H 'Accept: application/json' \
  -H 'Authorization: Token nb_pat_1a2b3c4d5e6f7g8h9i0j'
```

### Expected Response

If successful, you'll receive a JSON response with a list of peers:

```json
[
  {
    "id": "abc123xyz",
    "name": "my-laptop",
    "ip": "100.64.0.1",
    "connected": true,
    "last_seen": "2024-01-15T10:30:00Z",
    "os": "linux",
    "version": "0.24.0",
    "groups": [
      {"id": "group-id", "name": "developers"}
    ],
    "hostname": "john-laptop"
  }
]
```

### Troubleshooting

**401 Unauthorized:**
```json
{"message": "Invalid or missing authentication token", "code": 401}
```
→ Check that your token is correct and hasn't expired

**Empty array `[]`:**
→ You don't have any peers registered yet (this is normal for new accounts)

---

## Step 3: Explore More Endpoints

Now that authentication works, try other common operations:

### List All Users

```bash
curl -X GET https://api.netbird.io/api/users \
  -H 'Accept: application/json' \
  -H 'Authorization: Token <TOKEN>'
```

### List All Groups

```bash
curl -X GET https://api.netbird.io/api/groups \
  -H 'Accept: application/json' \
  -H 'Authorization: Token <TOKEN>'
```

### Get Specific Peer Details

```bash
curl -X GET https://api.netbird.io/api/peers/{peerId} \
  -H 'Accept: application/json' \
  -H 'Authorization: Token <TOKEN>'
```

Replace `{peerId}` with an actual peer ID from your list.

### List All Policies

```bash
curl -X GET https://api.netbird.io/api/policies \
  -H 'Accept: application/json' \
  -H 'Authorization: Token <TOKEN>'
```

---

## Step 4: Make a POST Request

Create a new group to organize your peers:

```bash
curl -X POST https://api.netbird.io/api/groups \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Token <TOKEN>' \
  -d '{
    "name": "my-first-group",
    "peers": [],
    "resources": []
  }'
```

### Expected Response

```json
{
  "id": "ch8i54g6lnn4g9hqv7n0",
  "name": "my-first-group",
  "peers_count": 0,
  "resources_count": 0,
  "issued": "api",
  "peers": [],
  "resources": []
}
```

---

## Step 5: Update a Resource

Update the group you just created:

```bash
curl -X PUT https://api.netbird.io/api/groups/{groupId} \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Token <TOKEN>' \
  -d '{
    "name": "my-updated-group",
    "peers": [],
    "resources": []
  }'
```

Replace `{groupId}` with the ID from the create response.

---

## Step 6: Delete a Resource

Remove the test group:

```bash
curl -X DELETE https://api.netbird.io/api/groups/{groupId} \
  -H 'Authorization: Token <TOKEN>'
```

**Note:** DELETE requests typically return no content (204 status) on success.

---

## Working with Self-Hosted NetBird

If you're using a self-hosted NetBird installation:

### Using Custom Management URL

```bash
curl -X GET https://your-netbird-server.com/api/peers \
  -H 'Accept: application/json' \
  -H 'Authorization: Token <TOKEN>'
```

### With Custom Port (if required)

Self-hosted installations may use port 33073:

```bash
curl -X GET https://your-netbird-server.com:33073/api/peers \
  -H 'Accept: application/json' \
  -H 'Authorization: Token <TOKEN>'
```

---

## Using Environment Variables

For better security, store your token in an environment variable:

### Linux/macOS

```bash
# Set the token
export NETBIRD_TOKEN="nb_pat_your_token_here"

# Use in requests
curl -X GET https://api.netbird.io/api/peers \
  -H 'Accept: application/json' \
  -H "Authorization: Token $NETBIRD_TOKEN"
```

### Windows (PowerShell)

```powershell
# Set the token
$env:NETBIRD_TOKEN = "nb_pat_your_token_here"

# Use in requests
curl -X GET https://api.netbird.io/api/peers `
  -H 'Accept: application/json' `
  -H "Authorization: Token $env:NETBIRD_TOKEN"
```

---

## Common Request Patterns

### GET Request Template

```bash
curl -X GET https://api.netbird.io/api/{endpoint} \
  -H 'Accept: application/json' \
  -H 'Authorization: Token <TOKEN>'
```

### POST Request Template

```bash
curl -X POST https://api.netbird.io/api/{endpoint} \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Token <TOKEN>' \
  -d '{
    "key": "value"
  }'
```

### PUT Request Template

```bash
curl -X PUT https://api.netbird.io/api/{endpoint}/{id} \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Token <TOKEN>' \
  -d '{
    "key": "updated_value"
  }'
```

### DELETE Request Template

```bash
curl -X DELETE https://api.netbird.io/api/{endpoint}/{id} \
  -H 'Authorization: Token <TOKEN>'
```

---

## Using Other HTTP Clients

### Python (requests library)

```python
import requests

TOKEN = "nb_pat_your_token_here"
BASE_URL = "https://api.netbird.io/api"

headers = {
    "Authorization": f"Token {TOKEN}",
    "Accept": "application/json"
}

response = requests.get(f"{BASE_URL}/peers", headers=headers)
peers = response.json()
print(peers)
```

### JavaScript (Axios)

```javascript
const axios = require('axios');

const TOKEN = 'nb_pat_your_token_here';
const BASE_URL = 'https://api.netbird.io/api';

const headers = {
  'Authorization': `Token ${TOKEN}`,
  'Accept': 'application/json'
};

axios.get(`${BASE_URL}/peers`, { headers })
  .then(response => console.log(response.data))
  .catch(error => console.error(error));
```

### Go

```go
package main

import (
    "fmt"
    "io"
    "net/http"
)

func main() {
    token := "nb_pat_your_token_here"
    url := "https://api.netbird.io/api/peers"

    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Token "+token)
    req.Header.Set("Accept", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
}
```

---

## Next Steps

Now that you've made your first successful API requests:

1. **[Explore Authentication Methods](authentication.md)** - Learn about OAuth2 and service users
2. **[Understand Error Responses](errors.md)** - Handle errors gracefully
3. **[Browse API Resources](../resources/)** - Explore all available endpoints
4. **Build Your Integration** - Start automating your NetBird network!

---

## Quick Reference Card

| Action | Endpoint | Method |
|--------|----------|--------|
| List peers | `/api/peers` | GET |
| Get peer details | `/api/peers/{id}` | GET |
| Delete peer | `/api/peers/{id}` | DELETE |
| List groups | `/api/groups` | GET |
| Create group | `/api/groups` | POST |
| Update group | `/api/groups/{id}` | PUT |
| Delete group | `/api/groups/{id}` | DELETE |
| List users | `/api/users` | GET |
| List policies | `/api/policies` | GET |
| List networks | `/api/networks` | GET |

---

## Getting Help

- **Documentation:** [docs.netbird.io/api](https://docs.netbird.io/api)
- **GitHub Issues:** [github.com/netbirdio/netbird/issues](https://github.com/netbirdio/netbird/issues)
- **Community:** Join the NetBird Slack workspace

Happy coding! 🚀
