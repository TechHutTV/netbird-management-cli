# NetBird API - Error Handling Guide

## Overview

Understanding HTTP error types helps diagnose and resolve API request issues. The NetBird API uses standard HTTP status codes to indicate the success or failure of requests.

**Success Indicator:** Check the response status code to determine if your request succeeded.

---

## HTTP Status Code Categories

### 2xx - Success

**A 2xx status code indicates a successful response.**

| Code | Status | Meaning |
|------|--------|---------|
| 200 | OK | Request succeeded, response body contains data |
| 201 | Created | Resource successfully created (POST requests) |
| 204 | No Content | Request succeeded, no response body (DELETE requests) |

**Example Success Response:**

```bash
$ curl -I -X GET https://api.netbird.io/api/peers \
    -H 'Authorization: Token <TOKEN>'

HTTP/2 200 OK
content-type: application/json
```

---

### 4xx - Client Errors

**A 4xx status code indicates a client error**, typically caused by:
- Insufficient permissions
- Incorrect request parameters
- Invalid authentication
- Missing required fields
- Resource not found

| Code | Status | Common Causes | Solution |
|------|--------|---------------|----------|
| 400 | Bad Request | Invalid JSON, missing required fields | Check request body syntax and required parameters |
| 401 | Unauthorized | Invalid or missing token | Verify authentication token is correct and not expired |
| 403 | Forbidden | Insufficient permissions | Check user/service user has required permissions |
| 404 | Not Found | Resource doesn't exist | Verify the resource ID or endpoint path |
| 409 | Conflict | Resource already exists or state conflict | Check for duplicate resources or concurrent modifications |
| 422 | Unprocessable Entity | Validation errors | Review field values against API requirements |
| 429 | Too Many Requests | Rate limit exceeded | Implement backoff and retry with delays |

### Common 4xx Error Examples

#### 400 Bad Request

**Cause:** Invalid request body or parameters

```json
{
  "message": "Invalid request parameters",
  "code": 400,
  "details": "field 'name' is required"
}
```

**Solution:**
```bash
# Incorrect (missing required field)
curl -X POST https://api.netbird.io/api/groups \
  -H 'Authorization: Token <TOKEN>' \
  -d '{}'

# Correct
curl -X POST https://api.netbird.io/api/groups \
  -H 'Authorization: Token <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"name": "my-group", "peers": [], "resources": []}'
```

#### 401 Unauthorized

**Cause:** Invalid, expired, or missing authentication token

```json
{
  "message": "Invalid or missing authentication token",
  "code": 401
}
```

**Solutions:**
- Verify token hasn't expired
- Check token format: `Authorization: Token <TOKEN>` (not `Bearer`)
- Generate a new token if necessary
- Ensure Authorization header is present

```bash
# Incorrect (missing auth)
curl -X GET https://api.netbird.io/api/peers

# Correct
curl -X GET https://api.netbird.io/api/peers \
  -H 'Authorization: Token nb_pat_abc123'
```

#### 403 Forbidden

**Cause:** Valid token but insufficient permissions

```json
{
  "message": "Insufficient permissions",
  "code": 403
}
```

**Solutions:**
- Check user role has required permissions
- Verify service user permissions in dashboard
- Contact account admin for permission escalation

#### 404 Not Found

**Cause:** Resource doesn't exist or invalid endpoint

```json
{
  "message": "Resource not found",
  "code": 404
}
```

**Solutions:**
- Verify resource ID is correct
- Check endpoint path for typos
- Ensure resource wasn't deleted
- List resources first to get valid IDs

```bash
# Get valid peer IDs first
curl -X GET https://api.netbird.io/api/peers \
  -H 'Authorization: Token <TOKEN>' | jq '.[].id'

# Then use correct ID
curl -X GET https://api.netbird.io/api/peers/valid-peer-id \
  -H 'Authorization: Token <TOKEN>'
```

#### 429 Too Many Requests

**Cause:** Rate limit exceeded

```json
{
  "message": "Rate limit exceeded",
  "code": 429
}
```

**Solutions:**
- Implement exponential backoff
- Reduce request frequency
- Batch operations where possible
- Cache responses to minimize requests

**Example: Exponential Backoff (Python)**

```python
import time
import requests

def make_request_with_retry(url, headers, max_retries=5):
    for attempt in range(max_retries):
        response = requests.get(url, headers=headers)

        if response.status_code == 429:
            wait_time = 2 ** attempt  # Exponential: 1s, 2s, 4s, 8s, 16s
            print(f"Rate limited. Waiting {wait_time}s...")
            time.sleep(wait_time)
            continue

        return response

    raise Exception("Max retries exceeded")
```

---

### 5xx - Server Errors

**A 5xx status code indicates a server error** on NetBird's side.

| Code | Status | Meaning | Action |
|------|--------|---------|--------|
| 500 | Internal Server Error | Unexpected server error | Retry request, contact support if persistent |
| 502 | Bad Gateway | Upstream service unavailable | Wait and retry, check NetBird status |
| 503 | Service Unavailable | Service temporarily down | Wait and retry with backoff |
| 504 | Gateway Timeout | Request timeout | Retry request or reduce payload size |

### Common 5xx Error Examples

#### 500 Internal Server Error

**Cause:** Unexpected error on the server

```json
{
  "message": "Internal server error",
  "code": 500
}
```

**Actions:**
1. Retry the request after a brief delay
2. Check NetBird status page or community channels
3. Report issue via GitHub or Slack if persistent
4. Include request details when reporting:
   - Endpoint and method
   - Request timestamp
   - Request ID (if provided in response headers)

---

## Error Response Structure

### Standard Error Format

```json
{
  "message": "Human-readable error description",
  "code": 400,
  "details": "Additional context (optional)"
}
```

### Fields

- **message** (string): Human-readable error description
- **code** (integer): HTTP status code
- **details** (string, optional): Additional context or field-specific errors
- **type** (string, optional): Error type categorization

### Example Error Responses

**Missing Required Field:**
```json
{
  "message": "Validation failed",
  "code": 400,
  "details": "field 'name' is required"
}
```

**Resource Not Found:**
```json
{
  "message": "Peer not found",
  "code": 404,
  "details": "peer with ID 'invalid-id' does not exist"
}
```

**Permission Denied:**
```json
{
  "message": "Permission denied",
  "code": 403,
  "details": "user does not have permission to delete this resource"
}
```

---

## Best Practices for Error Handling

### 1. Always Check Status Codes

```bash
# Use curl's -w flag to show status code
curl -w "\nHTTP Status: %{http_code}\n" \
  -X GET https://api.netbird.io/api/peers \
  -H 'Authorization: Token <TOKEN>'
```

```python
# Python: Always check response status
response = requests.get(url, headers=headers)

if response.status_code == 200:
    data = response.json()
    # Process data
elif response.status_code == 401:
    print("Authentication failed - check token")
elif response.status_code >= 500:
    print("Server error - retry later")
else:
    print(f"Error {response.status_code}: {response.text}")
```

### 2. Implement Retry Logic

**Retry-Worthy Status Codes:**
- 429 (Too Many Requests) - with backoff
- 500 (Internal Server Error) - with limited retries
- 502, 503, 504 (Service issues) - with backoff

**Don't Retry:**
- 400 (Bad Request) - fix the request first
- 401 (Unauthorized) - fix authentication first
- 403 (Forbidden) - requires permission changes
- 404 (Not Found) - resource doesn't exist

**Example Retry Logic:**

```python
import time

def api_request_with_retry(url, headers, max_retries=3):
    retry_codes = [429, 500, 502, 503, 504]

    for attempt in range(max_retries):
        response = requests.get(url, headers=headers)

        if response.status_code == 200:
            return response.json()

        if response.status_code in retry_codes:
            wait = min(2 ** attempt, 60)  # Max 60 seconds
            print(f"Retry {attempt + 1}/{max_retries} after {wait}s")
            time.sleep(wait)
            continue

        # Don't retry client errors (4xx except 429)
        raise Exception(f"Request failed: {response.status_code} - {response.text}")

    raise Exception("Max retries exceeded")
```

### 3. Parse and Log Error Details

```go
// Go: Structured error handling
type APIError struct {
    Message string `json:"message"`
    Code    int    `json:"code"`
    Details string `json:"details"`
}

func handleResponse(resp *http.Response) error {
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        return nil
    }

    var apiError APIError
    if err := json.NewDecoder(resp.Body).Decode(&apiError); err == nil {
        return fmt.Errorf("API error %d: %s - %s",
            apiError.Code, apiError.Message, apiError.Details)
    }

    return fmt.Errorf("API error %d: %s", resp.StatusCode, resp.Status)
}
```

### 4. Provide User-Friendly Messages

```javascript
// JavaScript: User-friendly error messages
function handleAPIError(error) {
    if (error.response) {
        const status = error.response.status;

        switch (status) {
            case 400:
                return "Invalid request. Please check your input.";
            case 401:
                return "Authentication failed. Please log in again.";
            case 403:
                return "You don't have permission for this action.";
            case 404:
                return "Resource not found.";
            case 429:
                return "Too many requests. Please wait and try again.";
            case 500:
                return "Server error. Please try again later.";
            default:
                return `An error occurred (${status}). Please try again.`;
        }
    }

    return "Network error. Please check your connection.";
}
```

---

## Beta API Considerations

**The NetBird API is currently in Beta**, which means:

- Error handling may not be fully comprehensive yet
- Error messages are being refined
- New error codes may be introduced
- Error response formats may evolve

### Recommendations for Beta API

1. **Handle unexpected errors gracefully** - Don't rely solely on documented error codes
2. **Validate inputs client-side** - Catch errors before API calls
3. **Log detailed error information** - Helps with debugging and reporting
4. **Provide feedback** - Report confusing or missing error messages
5. **Check for updates** - Error handling improvements are ongoing

---

## Debugging Tips

### Use Verbose cURL

```bash
# Show full request and response headers
curl -v -X GET https://api.netbird.io/api/peers \
  -H 'Authorization: Token <TOKEN>'
```

### Validate JSON

```bash
# Pipe request body through jq to validate JSON syntax
echo '{"name": "test-group", "peers": []}' | jq .

# Then use in request
curl -X POST https://api.netbird.io/api/groups \
  -H 'Authorization: Token <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d "$(echo '{"name": "test-group", "peers": [], "resources": []}' | jq -c .)"
```

### Check Request Format

Common issues:
- Missing `Content-Type: application/json` header for POST/PUT
- Incorrect JSON syntax (trailing commas, unquoted keys)
- Wrong HTTP method (GET vs POST)
- Missing required fields

---

## Getting Help

If you encounter persistent errors:

1. **Check the documentation** - [docs.netbird.io/api](https://docs.netbird.io/api)
2. **Search existing issues** - [GitHub Issues](https://github.com/netbirdio/netbird/issues)
3. **Ask the community** - NetBird Slack workspace
4. **Report bugs** - Create detailed issue on GitHub

### When Reporting Errors

Include:
- HTTP method and endpoint
- Request headers and body (redact token!)
- Response status and body
- Timestamp of request
- Expected vs actual behavior

---

## Next Steps

- **[Authentication Guide](authentication.md)** - Ensure proper authentication
- **[Quickstart Guide](quickstart.md)** - Learn API basics
- **[API Resources](../resources/)** - Explore all endpoints

---

**Remember:** Examining both the status code and error details (type and message) is key to effective API troubleshooting!
