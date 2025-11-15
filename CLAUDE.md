# CLAUDE.md - AI Assistant Guide for NetBird Management CLI

This document provides comprehensive guidance for AI assistants working on the NetBird Management CLI project.

## Project Overview

**NetBird Management CLI** (`netbird-manage`) is an unofficial command-line tool written in Go that provides terminal-based management for NetBird networks. It interfaces with the NetBird REST API to manage peers, groups, networks, and access control policies.

**Key Characteristics:**
- **Language:** Go 1.25.4+ (requires minimum 1.18)
- **Dependencies:** Zero external dependencies (stdlib only)
- **Architecture:** Single-binary CLI with flat file structure
- **API:** RESTful HTTP client with Bearer token authentication
- **Lines of Code:** ~942 lines across 9 Go files

**Project Links:**
- NetBird API Documentation: https://docs.netbird.io/api
- NetBird Website: https://netbird.io/

**Local API References:**
- `API_REFERENCE.md` - Quick navigation and endpoint reference
- `docs/api/` - Complete API documentation directory:
  - `README.md` - Main documentation index
  - `introduction.md` - API overview and getting started
  - `guides/` - Authentication, quickstart, error handling guides
  - `resources/` - Detailed documentation for all API endpoints
- `.claude/commands/api-docs.md` - Slash command to fetch live API docs (use `/api-docs`)

---

## Codebase Structure

### File Organization

```
netbird-management-cli/
├── netbird-manage.go    # Main entry point and command router (125 lines)
├── client.go            # HTTP API client with authentication (62 lines)
├── config.go            # Configuration file management (98 lines)
├── models.go            # Data type definitions (88 lines)
├── helpers.go           # Utility functions and formatters (56 lines)
├── peers.go             # Peer command handlers (248 lines)
├── groups.go            # Group command handlers (121 lines)
├── networks.go          # Network command handlers (61 lines)
├── policies.go          # Policy command handlers (83 lines)
├── go.mod               # Go module definition
├── README.md            # User-facing documentation
├── API_REFERENCE.md     # Quick API navigation and reference
├── LICENSE              # MIT/Apache dual license
├── CLAUDE.md            # This file - AI assistant guide
├── docs/
│   └── api/             # Complete NetBird API documentation
│       ├── README.md    # API documentation index
│       ├── introduction.md
│       ├── guides/
│       │   ├── authentication.md
│       │   ├── quickstart.md
│       │   └── errors.md
│       └── resources/   # Per-resource endpoint documentation
│           └── README.md
└── .claude/
    └── commands/
        └── api-docs.md  # Slash command for fetching API docs
```

### Module Responsibilities

| File | Purpose | Key Functions |
|------|---------|---------------|
| `netbird-manage.go` | Entry point, command routing | `main()`, `printUsage()`, `handleConnectCommand()` |
| `client.go` | HTTP client, API requests | `NewClient()`, `makeRequest()` |
| `config.go` | Config persistence, loading | `loadConfig()`, `testAndSaveConfig()`, `getConfigPath()` |
| `models.go` | Data structures | `Peer`, `Group`, `Network`, `Policy`, `Config` |
| `helpers.go` | Formatting, utilities | `formatOS()`, `printConnectStatus()` |
| `peers.go` | Peer operations | `handlePeersCommand()`, `listPeers()`, `modifyPeerGroup()` |
| `groups.go` | Group operations | `handleGroupsCommand()`, `listGroups()`, `getGroupByName()` |
| `networks.go` | Network operations | `handleNetworksCommand()`, `listNetworks()` |
| `policies.go` | Policy operations | `handlePoliciesCommand()`, `listPolicies()` |

---

## Architecture & Design Patterns

### Command Flow

```
User Input (CLI args)
    ↓
main() - Command Router
    ↓
[Special Case: connect command - no config required]
    ↓
loadConfig() - Load credentials
    ↓
NewClient() - Create HTTP client with token
    ↓
Command Handler (handlePeersCommand, handleGroupsCommand, etc.)
    ↓
Client.makeRequest() - HTTP API call with Bearer auth
    ↓
JSON Response → Model Deserialization
    ↓
Formatted Output (tabwriter for tables)
```

### Design Patterns Used

**1. Client Pattern**
```go
type Client struct {
    Token         string
    ManagementURL string
    HTTPClient    *http.Client
}

client := NewClient(config.Token, config.ManagementURL)
```

**2. Command Handler Pattern**
Each domain has a dedicated handler function:
- `handlePeersCommand(client, args)`
- `handleGroupsCommand(client, args)`
- `handleNetworksCommand(client, args)`
- `handlePoliciesCommand(client, args)`

**3. Repository Pattern**
Methods on Client act as repositories:
- `client.getPeerByID(id)`
- `client.getGroupByName(name)`
- `client.listPeers()`

**4. Configuration Fallback Pattern**
```
1. Try: $HOME/.netbird-manage.json
2. Fallback: NETBIRD_API_TOKEN env var
3. Fail: Return error (not connected)
```

---

## Development Workflows

### Building the Project

```bash
# Initialize module (if not already done)
go mod init netbird-manage

# Build binary
go build

# Result: ./netbird-manage (or netbird-manage.exe on Windows)
```

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o netbird-manage-linux-amd64

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o netbird-manage-windows-amd64.exe

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o netbird-manage-darwin-arm64

# macOS AMD64 (Intel)
GOOS=darwin GOARCH=amd64 go build -o netbird-manage-darwin-amd64
```

### Testing the CLI

```bash
# Connect to NetBird API
./netbird-manage connect --token "your-api-token"

# List peers
./netbird-manage peer --list

# Inspect specific peer
./netbird-manage peer --inspect "peer-id"

# Add peer to group
./netbird-manage peer --edit "peer-id" --add-group "group-name"

# List groups
./netbird-manage group

# List networks
./netbird-manage networks

# List policies
./netbird-manage policy
```

### Configuration File Location

- **Path:** `$HOME/.netbird-manage.json`
- **Format:** JSON
- **Permissions:** `0600` (owner read/write only)
- **Structure:**
```json
{
  "token": "your-api-token-here",
  "management_url": "https://api.netbird.io/api"
}
```

---

## Key Conventions & Code Style

### Naming Conventions

**Functions & Methods:**
- Use camelCase: `handlePeersCommand()`, `listPeers()`
- Action verbs: `get`, `list`, `remove`, `modify`, `update`, `test`
- Handler prefix for command handlers: `handle*Command()`

**Constants:**
- camelCase: `configFileName`, `defaultCloudURL`

**Variables:**
- Short names for receivers: `c` for Client, `p` for Peer
- Descriptive names for complex variables: `peerID`, `groupName`, `managementURL`

**Flags:**
- Kebab-case: `--add-group`, `--remove-group`, `--management-url`
- Boolean flags: `--list`, `--enabled`

### Error Handling

**Standard Pattern:**
```go
result, err := someFunction()
if err != nil {
    return fmt.Errorf("context: %v", err)
}
```

**API Errors:**
```go
if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    var apiError struct {
        Message string `json:"message"`
        Code    int    `json:"code"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&apiError); err == nil {
        return nil, fmt.Errorf("api request failed: %d %s", apiError.Code, apiError.Message)
    }
    return nil, fmt.Errorf("api request failed: %s", resp.Status)
}
```

**User-Facing Errors:**
- Send to stderr: `fmt.Fprintf(os.Stderr, "Error: %v\n", err)`
- Exit with code 1: `os.Exit(1)`

### JSON Serialization

**Always use struct tags:**
```go
type Model struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

**Encoding:**
```go
data, err := json.MarshalIndent(obj, "", "  ")
```

**Decoding:**
```go
var result Model
err := json.NewDecoder(resp.Body).Decode(&result)
```

### Output Formatting

**Tables - Use tabwriter:**
```go
w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
fmt.Fprintln(w, "HEADER1\tHEADER2\tHEADER3")
fmt.Fprintln(w, "-------\t-------\t-------")
for _, item := range items {
    fmt.Fprintf(w, "%s\t%s\t%s\n", item.Field1, item.Field2, item.Field3)
}
w.Flush()
```

**Status Messages:**
- Informational: `fmt.Println("message")` → stdout
- Errors: `fmt.Fprintln(os.Stderr, "Error: message")` → stderr
- Success: `fmt.Println("✓ Operation successful")` → stdout

---

## Data Models

### Core Models

**Config** (`config.go`)
```go
type Config struct {
    Token         string `json:"token"`
    ManagementURL string `json:"management_url"`
}
```

**Peer** (`models.go`)
```go
type Peer struct {
    ID        string        `json:"id"`
    Name      string        `json:"name"`
    IP        string        `json:"ip"`
    Connected bool          `json:"connected"`
    LastSeen  string        `json:"last_seen"`
    OS        string        `json:"os"`
    Version   string        `json:"version"`
    Groups    []PolicyGroup `json:"groups"`
    Hostname  string        `json:"hostname"`
}
```

**GroupDetail** (Full group representation)
```go
type GroupDetail struct {
    ID             string          `json:"id"`
    Name           string          `json:"name"`
    PeersCount     int             `json:"peers_count"`
    ResourcesCount int             `json:"resources_count"`
    Issued         string          `json:"issued"`
    Peers          []Peer          `json:"peers"`
    Resources      []GroupResource `json:"resources"`
}
```

**PolicyGroup** (Lightweight group reference)
```go
type PolicyGroup struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

**Network** (`models.go`)
```go
type Network struct {
    ID                string   `json:"id"`
    Name              string   `json:"name"`
    Routers           []string `json:"routers"`
    RoutingPeersCount int      `json:"routing_peers_count"`
    Resources         []string `json:"resources"`
    Policies          []string `json:"policies"`
    Description       string   `json:"description"`
}
```

**Policy** (`models.go`)
```go
type Policy struct {
    ID          string       `json:"id"`
    Name        string       `json:"name"`
    Description string       `json:"description"`
    Enabled     bool         `json:"enabled"`
    Rules       []PolicyRule `json:"rules"`
}

type PolicyRule struct {
    ID           string        `json:"id"`
    Name         string        `json:"name"`
    Enabled      bool          `json:"enabled"`
    Action       string        `json:"action"` // "accept" or "drop"
    Protocol     string        `json:"protocol"`
    Sources      []PolicyGroup `json:"sources"`
    Destinations []PolicyGroup `json:"destinations"`
}
```

### Model Relationships

```
Peer
  └── Groups: []PolicyGroup (many-to-many)

GroupDetail
  ├── Peers: []Peer (members)
  └── Resources: []GroupResource

Policy
  └── Rules: []PolicyRule
      ├── Sources: []PolicyGroup
      └── Destinations: []PolicyGroup

Network
  ├── Routers: []string (peer IDs)
  ├── Resources: []string (resource IDs)
  └── Policies: []string (policy IDs)
```

---

## API Integration

### Authentication

**Bearer Token Authentication:**
```go
req.Header.Set("Authorization", "Token "+c.Token)
```

**Token Sources:**
1. `--token` flag during `connect` command
2. `$HOME/.netbird-manage.json` config file
3. `NETBIRD_API_TOKEN` environment variable

### API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/peers` | List all peers |
| GET | `/peers/{id}` | Get peer details |
| DELETE | `/peers/{id}` | Remove peer |
| GET | `/groups` | List all groups |
| GET | `/groups/{id}` | Get group details |
| PUT | `/groups/{id}` | Update group (members, resources) |
| GET | `/networks` | List all networks |
| GET | `/policies` | List all policies |

### Making API Requests

**Pattern:**
```go
func (c *Client) exampleMethod() error {
    // 1. Make request
    resp, err := c.makeRequest("GET", "/endpoint", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 2. Decode response
    var result []Model
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return fmt.Errorf("failed to decode response: %v", err)
    }

    // 3. Process and display
    for _, item := range result {
        fmt.Println(item.Name)
    }

    return nil
}
```

**PUT Request Example:**
```go
reqBody := GroupPutRequest{
    Name:      group.Name,
    Peers:     updatedPeerIDs,
    Resources: resources,
}

bodyBytes, _ := json.Marshal(reqBody)
resp, err := c.makeRequest("PUT", "/groups/"+groupID, bytes.NewReader(bodyBytes))
```

### Management URL Configuration

**Default Cloud URL:**
```go
const defaultCloudURL = "https://api.netbird.io/api"
```

**Custom Self-Hosted:**
```bash
netbird-manage connect --token "token" --management-url "https://your-server.com/api"
```

---

## Common Tasks

### Adding a New Command

**1. Define the command handler function:**
```go
// In new file: example.go
func handleExampleCommand(client *Client, args []string) error {
    exampleCmd := flag.NewFlagSet("example", flag.ContinueOnError)
    exampleCmd.SetOutput(os.Stderr)

    listFlag := exampleCmd.Bool("list", false, "List examples")

    if err := exampleCmd.Parse(args[1:]); err != nil {
        return err
    }

    if *listFlag {
        return client.listExamples()
    }

    exampleCmd.Usage()
    return nil
}
```

**2. Add Client method:**
```go
func (c *Client) listExamples() error {
    resp, err := c.makeRequest("GET", "/examples", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var examples []Example
    if err := json.NewDecoder(resp.Body).Decode(&examples); err != nil {
        return fmt.Errorf("failed to decode response: %v", err)
    }

    // Display output
    w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
    fmt.Fprintln(w, "ID\tNAME\tSTATUS")
    fmt.Fprintln(w, "--\t----\t------")
    for _, ex := range examples {
        fmt.Fprintf(w, "%s\t%s\t%s\n", ex.ID, ex.Name, ex.Status)
    }
    w.Flush()

    return nil
}
```

**3. Register in main router (netbird-manage.go):**
```go
switch command {
case "example":
    if err := handleExampleCommand(client, args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
// ... existing cases
}
```

**4. Update printUsage():**
```go
func printUsage() {
    fmt.Println("Usage: netbird-manage <command> [options]")
    fmt.Println("\nCommands:")
    fmt.Println("  connect              Authenticate and save API token")
    fmt.Println("  example              Manage examples")
    // ... existing commands
}
```

**5. Add model if needed (models.go):**
```go
type Example struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Status string `json:"status"`
}
```

### Adding a New Flag to Existing Command

**Example: Add `--name` filter to peer list**

**In peers.go:**
```go
func handlePeersCommand(client *Client, args []string) error {
    peerCmd := flag.NewFlagSet("peer", flag.ContinueOnError)

    listFlag := peerCmd.Bool("list", false, "List all peers")
    nameFilter := peerCmd.String("name", "", "Filter by peer name") // NEW

    // ... parse flags

    if *listFlag {
        return client.listPeers(*nameFilter) // Pass filter
    }
}
```

**Update Client method:**
```go
func (c *Client) listPeers(nameFilter string) error {
    // ... fetch peers

    for _, peer := range peers {
        // Apply filter if provided
        if nameFilter != "" && !strings.Contains(peer.Name, nameFilter) {
            continue
        }

        fmt.Fprintf(w, "%s\t%s\t...\n", peer.ID, peer.Name)
    }
}
```

### Fixing Bugs

**Common bug locations:**
1. **Flag parsing errors** → Check `flag.NewFlagSet` and `Parse()` calls
2. **API errors** → Check `makeRequest()` error handling
3. **JSON decoding errors** → Verify struct tags match API response
4. **Config loading issues** → Check `loadConfig()` fallback logic
5. **Output formatting** → Verify tabwriter usage and column alignment

**Debugging tips:**
```go
// Add debug output (remove before commit)
fmt.Fprintf(os.Stderr, "DEBUG: Value = %+v\n", variable)

// Check API response
bodyBytes, _ := io.ReadAll(resp.Body)
fmt.Fprintf(os.Stderr, "DEBUG: API Response = %s\n", string(bodyBytes))
```

### Refactoring

**When refactoring:**
1. Maintain backward compatibility for commands
2. Keep config file format unchanged
3. Preserve API client interface
4. Update README.md if user-facing changes
5. Test all commands after changes

**Safe refactoring areas:**
- Internal helper functions
- Output formatting
- Error messages
- Code organization (moving functions between files)

**Risky refactoring areas:**
- Config file structure (breaks existing users)
- Command names or flags (breaks scripts)
- API endpoint paths (breaks API integration)

---

## Testing Strategy

### Current Testing Approach

**The project currently has NO automated tests.** Testing is manual via CLI commands.

**Manual testing checklist:**
```bash
# 1. Test connection
netbird-manage connect --token "test-token"
netbird-manage connect  # Check status

# 2. Test peer operations
netbird-manage peer --list
netbird-manage peer --inspect "peer-id"
netbird-manage peer --edit "peer-id" --add-group "group-name"
netbird-manage peer --remove "peer-id"

# 3. Test group operations
netbird-manage group

# 4. Test network operations
netbird-manage networks

# 5. Test policy operations
netbird-manage policy

# 6. Test error cases
netbird-manage peer --list  # Without connection
netbird-manage peer --edit "invalid-id" --add-group "test"
```

### Recommended Testing Approach (for AI assistants adding tests)

**Unit Tests:**
```go
// client_test.go
func TestMakeRequest(t *testing.T) {
    // Mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify auth header
        if r.Header.Get("Authorization") != "Token test-token" {
            t.Error("Missing or incorrect auth header")
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode([]Peer{{ID: "test", Name: "Test Peer"}})
    }))
    defer server.Close()

    client := NewClient("test-token", server.URL)
    resp, err := client.makeRequest("GET", "/peers", nil)

    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }

    if resp.StatusCode != 200 {
        t.Errorf("Expected status 200, got %d", resp.StatusCode)
    }
}
```

**Integration Tests:**
```go
// integration_test.go (requires API token)
func TestListPeers(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    token := os.Getenv("NETBIRD_API_TOKEN")
    if token == "" {
        t.Skip("No API token provided")
    }

    client := NewClient(token, defaultCloudURL)
    err := client.listPeers("")

    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
}
```

**Running tests:**
```bash
# Unit tests only
go test -short

# All tests (including integration)
NETBIRD_API_TOKEN="your-token" go test

# With coverage
go test -cover
```

---

## Important Notes & Gotchas

### Security Considerations

**1. Token Storage Security**
- Config file uses `0600` permissions (owner-only read/write)
- Never log or print tokens in error messages
- Tokens stored in plaintext (acceptable for local CLI)

**2. API Token Validation**
- Token validated on `connect` via test API call (GET /peers)
- Invalid token → user sees API error message
- No token expiration handling (user must reconnect)

**3. HTTPS Enforcement**
- Default cloud URL uses HTTPS
- Self-hosted URLs should use HTTPS (not enforced by CLI)

### Common Pitfalls

**1. Group Modifications Require Full Object**
When updating a group (adding/removing peer), you must:
- Fetch full group details (GET /groups/{id})
- Modify the peer list
- Send complete updated group (PUT /groups/{id})

```go
// INCORRECT - This won't work
client.makeRequest("POST", "/groups/"+groupID+"/peers/"+peerID, nil)

// CORRECT - Full group update required
group := client.getGroupByID(groupID)
group.Peers = append(group.Peers, newPeer)
client.updateGroup(groupID, group)
```

**2. Flag Parsing Order Matters**
```go
// CORRECT - Parse before checking flags
peerCmd.Parse(args[1:])
if *listFlag {
    // Use flag
}

// INCORRECT - Check before parsing
if *listFlag { // Always false!
    // Never executes
}
peerCmd.Parse(args[1:])
```

**3. Response Body Must Be Closed**
```go
// ALWAYS defer close after checking error
resp, err := c.makeRequest("GET", "/endpoint", nil)
if err != nil {
    return err
}
defer resp.Body.Close() // IMPORTANT - prevents leak
```

**4. Environment Variable Fallback**
- `NETBIRD_API_TOKEN` only used if config file doesn't exist
- Changing env var won't affect existing config
- To use env var: delete `~/.netbird-manage.json` first

### API Quirks

**1. Peer IDs vs Names**
- API uses UUIDs for peer IDs
- Peer names are user-friendly but not unique
- Always use IDs for operations, display names for UX

**2. Group Name Lookup**
- No API endpoint to search groups by name
- Must fetch all groups and filter locally
- See `getGroupByName()` in `groups.go`

**3. Connected vs LastSeen**
- `Connected` is boolean (current status)
- `LastSeen` is timestamp string (ISO 8601)
- Use `Connected` for status, `LastSeen` for troubleshooting

**4. Policy Actions**
- `action` field: `"accept"` or `"drop"`
- Case-sensitive in API responses
- Used in firewall rules (allow/block traffic)

### Development Environment

**Go Version Requirements:**
- Minimum: Go 1.18 (uses generics? No, but go.mod specifies 1.25.4)
- Recommended: Latest stable Go version
- No CGO required (pure Go)

**IDE Recommendations:**
- VS Code with Go extension
- GoLand
- Vim/Neovim with gopls

**Useful Go Commands:**
```bash
# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run

# Check for issues
go vet ./...

# Build for current platform
go build

# Show dependencies
go list -m all

# Tidy dependencies
go mod tidy
```

---

## Future Development Areas

### Planned Features (from README)

**1. Full CRUD Operations**
- Groups: Create, update, delete (currently read-only)
- Networks: Full management (currently read-only)
- Policies: Create, update, delete rules (currently read-only)
- Users: Invite, remove, update roles (not implemented)
- Setup Keys, Routes, DNS (not implemented)

**2. YAML-based Policy Management**
```bash
# Export policies to YAML
netbird-manage policy export > my-policies.yml

# Apply policies from YAML (GitOps workflow)
netbird-manage policy apply -f my-policies.yml
```

**3. Interactive CLI Features**
- Confirmation prompts for destructive operations
- Interactive peer/group selection (using bubbletea library)
- TUI mode for browsing resources

### Implementation Guidance for New Features

**Adding CRUD for Groups:**

1. **Create Group:**
```go
func (c *Client) createGroup(name string, description string) error {
    reqBody := map[string]interface{}{
        "name": name,
        "description": description,
        "peers": []string{},
        "resources": []interface{}{},
    }

    bodyBytes, _ := json.Marshal(reqBody)
    resp, err := c.makeRequest("POST", "/groups", bytes.NewReader(bodyBytes))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    fmt.Println("Group created successfully")
    return nil
}
```

2. **Delete Group:**
```go
func (c *Client) deleteGroup(groupID string) error {
    resp, err := c.makeRequest("DELETE", "/groups/"+groupID, nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    fmt.Println("Group deleted successfully")
    return nil
}
```

3. **Add flags to handleGroupsCommand:**
```go
createFlag := groupCmd.String("create", "", "Create new group")
deleteFlag := groupCmd.String("delete", "", "Delete group by ID")
descFlag := groupCmd.String("description", "", "Group description")

if *createFlag != "" {
    return client.createGroup(*createFlag, *descFlag)
}
if *deleteFlag != "" {
    return client.deleteGroup(*deleteFlag)
}
```

**Adding YAML Export/Apply:**

1. **Install YAML library:**
```bash
go get gopkg.in/yaml.v3
```

2. **Export implementation:**
```go
func (c *Client) exportPolicies(outputPath string) error {
    // Fetch all policies
    resp, err := c.makeRequest("GET", "/policies", nil)
    // ...

    var policies []Policy
    json.NewDecoder(resp.Body).Decode(&policies)

    // Convert to YAML
    yamlData, err := yaml.Marshal(policies)
    if err != nil {
        return fmt.Errorf("failed to marshal YAML: %v", err)
    }

    // Write to file
    if err := os.WriteFile(outputPath, yamlData, 0644); err != nil {
        return fmt.Errorf("failed to write file: %v", err)
    }

    fmt.Printf("Policies exported to %s\n", outputPath)
    return nil
}
```

3. **Apply implementation:**
```go
func (c *Client) applyPolicies(inputPath string) error {
    // Read YAML file
    yamlData, err := os.ReadFile(inputPath)
    if err != nil {
        return fmt.Errorf("failed to read file: %v", err)
    }

    // Parse YAML
    var policies []Policy
    if err := yaml.Unmarshal(yamlData, &policies); err != nil {
        return fmt.Errorf("failed to parse YAML: %v", err)
    }

    // Apply each policy (PUT or POST depending on ID)
    for _, policy := range policies {
        if policy.ID != "" {
            // Update existing
            reqBody, _ := json.Marshal(policy)
            c.makeRequest("PUT", "/policies/"+policy.ID, bytes.NewReader(reqBody))
        } else {
            // Create new
            reqBody, _ := json.Marshal(policy)
            c.makeRequest("POST", "/policies", bytes.NewReader(reqBody))
        }
    }

    fmt.Printf("Applied %d policies from %s\n", len(policies), inputPath)
    return nil
}
```

---

## Debugging & Troubleshooting

### Common Issues

**Issue: "Error: Not connected"**
- **Cause:** No config file and no `NETBIRD_API_TOKEN` env var
- **Fix:** Run `netbird-manage connect --token "your-token"`

**Issue: "api request failed: 401 Unauthorized"**
- **Cause:** Invalid or expired token
- **Fix:** Generate new token and reconnect

**Issue: "api request failed: 404 Not Found"**
- **Cause:** Invalid peer/group/network ID, or API endpoint changed
- **Fix:** Verify ID exists with list command, check API docs

**Issue: "Group not found"**
- **Cause:** Group name doesn't match exactly (case-sensitive)
- **Fix:** List groups to see exact names

**Issue: Peer not added to group**
- **Cause:** Group update failed silently, or peer already in group
- **Fix:** Check response status, verify with `peer --inspect`

### Debug Mode Implementation

**Add debug flag to client:**
```go
type Client struct {
    Token         string
    ManagementURL string
    HTTPClient    *http.Client
    Debug         bool // NEW
}

func (c *Client) makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
    url := c.ManagementURL + endpoint

    if c.Debug {
        fmt.Fprintf(os.Stderr, "DEBUG: %s %s\n", method, url)
        if body != nil {
            bodyBytes, _ := io.ReadAll(body)
            fmt.Fprintf(os.Stderr, "DEBUG: Request Body = %s\n", string(bodyBytes))
            body = bytes.NewReader(bodyBytes) // Recreate reader
        }
    }

    // ... rest of method

    if c.Debug && resp != nil {
        fmt.Fprintf(os.Stderr, "DEBUG: Response Status = %s\n", resp.Status)
    }
}
```

**Add --debug flag to main:**
```go
if len(os.Args) > 1 && os.Args[1] == "--debug" {
    client.Debug = true
    os.Args = append(os.Args[:1], os.Args[2:]...) // Remove debug flag
}
```

---

## Git & Version Control

### Branch Strategy

**Main Branch:**
- `main` - Stable releases
- No direct commits to main (use PRs for changes)

**Feature Branches:**
- Naming: `feature/description` or `fix/description`
- Example: `feature/yaml-export`, `fix/group-update-bug`

**AI Assistant Branches:**
- Follow the Claude-specific branch naming: `claude/claude-md-mi0n3hu27wdac3cg-017fio3VaK5aELmeEYqABU2G`
- Always develop on the designated Claude branch
- Push to remote with `-u` flag: `git push -u origin <branch-name>`

### Commit Message Guidelines

**Format:**
```
<type>: <short description>

<optional longer description>
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code restructuring
- `docs:` - Documentation changes
- `style:` - Formatting changes
- `test:` - Adding tests
- `chore:` - Maintenance tasks

**Examples:**
```
feat: Add YAML export for policies

Implements policy export to YAML file for GitOps workflows.
Users can now run: netbird-manage policy export > policies.yml

fix: Handle missing group name in peer edit

Previously crashed when group name didn't exist. Now returns
clear error message to user.

docs: Update README with YAML export instructions
```

### Pull Request Checklist

When creating a PR:
- [ ] Code builds successfully (`go build`)
- [ ] No linting errors (`go vet ./...`, `go fmt ./...`)
- [ ] Manually tested all affected commands
- [ ] Updated README.md if user-facing changes
- [ ] Updated CLAUDE.md if architectural changes
- [ ] Commit messages follow guidelines
- [ ] No debug code or commented-out code
- [ ] No hardcoded tokens or credentials

---

## Resources & References

### API Documentation

**Local Documentation:**
- **API Documentation Hub:** `docs/api/README.md` (comprehensive documentation index)
- **Quick Reference:** `API_REFERENCE.md` (quick navigation and CLI mappings)
- **Introduction:** `docs/api/introduction.md` (API overview and getting started)
- **Guides:**
  - `docs/api/guides/authentication.md` - OAuth2 and PAT setup
  - `docs/api/guides/quickstart.md` - Your first API request
  - `docs/api/guides/errors.md` - Error handling and troubleshooting
- **Resources:** `docs/api/resources/` (detailed per-endpoint documentation)
- **Slash Command:** `/api-docs` (fetch live API documentation on demand)

**External Resources:**
- **Live API Docs:** https://docs.netbird.io/api (official, always up-to-date)
- **OpenAPI Spec:** https://api.netbird.io/api/openapi.json
- **Source Documentation:** https://github.com/netbirdio/docs/tree/main/src/pages/ipa

### Official Documentation
- **NetBird API Docs:** https://docs.netbird.io/api
- **NetBird Website:** https://netbird.io/
- **Go Documentation:** https://go.dev/doc/

### Go Standard Library Packages Used
- `encoding/json` - JSON encoding/decoding
- `flag` - Command-line flag parsing
- `fmt` - Formatted I/O
- `io` - I/O primitives
- `net/http` - HTTP client/server
- `os` - Operating system interface
- `path/filepath` - File path manipulation
- `strings` - String utilities
- `text/tabwriter` - Tabular output formatting

### Useful Go Resources
- **Effective Go:** https://go.dev/doc/effective_go
- **Go Code Review Comments:** https://go.dev/wiki/CodeReviewComments
- **Standard Library:** https://pkg.go.dev/std

---

## Questions & Support

### For AI Assistants

If you encounter unclear requirements:
1. Check this CLAUDE.md file first
2. Review README.md for user-facing documentation
3. Examine similar existing code patterns
4. Check NetBird API documentation
5. Ask the user for clarification if still unclear

### For Developers

- **Issues:** Report bugs and request features via GitHub Issues
- **License:** MIT/Apache dual license (see LICENSE file)
- **Contributing:** Follow conventions in this document

---

**Last Updated:** 2025-11-15
**Document Version:** 1.0
**Codebase Version:** ~942 lines of Go code across 9 files
