# CLAUDE.md - AI Assistant Guide for NetBird Management CLI

This document provides comprehensive guidance for AI assistants working on the NetBird Management CLI project.

## Project Overview

**NetBird Management CLI** (`netbird-manage`) is an unofficial command-line tool written in Go that provides terminal-based management for NetBird networks. It interfaces with the NetBird REST API to manage peers, groups, networks, and access control policies.

**Key Characteristics:**
- **Language:** Go 1.24+ (requires minimum 1.18)
- **Dependencies:** Minimal external dependencies (only `gopkg.in/yaml.v3` for YAML export/import)
- **Architecture:** Single-binary CLI with cmd/internal package structure
- **API:** RESTful HTTP client with Bearer token authentication
- **Packages:** Organized into `cmd/netbird-manage`, `internal/commands`, `internal/client`, `internal/config`, `internal/models`, `internal/helpers`

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
├── cmd/
│   └── netbird-manage/
│       └── main.go              # Main entry point and command router
├── internal/
│   ├── client/
│   │   └── client.go            # HTTP API client with authentication and debug logging
│   ├── config/
│   │   └── config.go            # Configuration file management
│   ├── models/
│   │   └── models.go            # Data type definitions
│   ├── helpers/
│   │   └── helpers.go           # Utility functions, formatters, and color output
│   └── commands/
│       ├── service.go           # Service struct that wraps client for commands
│       ├── usage.go             # Help text and usage functions
│       ├── peers.go             # Peer command handlers
│       ├── groups.go            # Group command handlers
│       ├── networks.go          # Network command handlers
│       ├── policies.go          # Policy command handlers
│       ├── setup_keys.go        # Setup key command handlers
│       ├── users.go             # User management handlers
│       ├── tokens.go            # Token management handlers
│       ├── routes.go            # Route management handlers
│       ├── dns.go               # DNS management handlers
│       ├── posture_checks.go    # Posture check handlers
│       ├── events.go            # Event/audit log handlers
│       ├── geo_locations.go     # Geographic location data handlers
│       ├── accounts.go          # Account management handlers
│       ├── ingress_ports.go     # Ingress port handlers (Cloud-only)
│       ├── export.go            # YAML export functionality
│       └── import.go            # YAML import functionality
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── README.md                    # User-facing documentation
├── API_REFERENCE.md             # Quick API navigation and reference
├── LICENSE                      # MIT/Apache dual license
├── CLAUDE.md                    # This file - AI assistant guide
├── docs/
│   └── api/                     # Complete NetBird API documentation
│       ├── README.md            # API documentation index
│       ├── introduction.md
│       ├── guides/
│       │   ├── authentication.md
│       │   ├── quickstart.md
│       │   └── errors.md
│       └── resources/           # Per-resource endpoint documentation
│           └── README.md
└── .claude/
    └── commands/
        └── api-docs.md          # Slash command for fetching API docs
```

### Module Responsibilities

| Package/File | Purpose | Key Functions/Types |
|--------------|---------|---------------------|
| `cmd/netbird-manage/main.go` | Entry point, command routing, global flags | `main()`, `handleConnectCommand()`, `debugMode` |
| `internal/client/client.go` | HTTP client, API requests, debug logging | `New()`, `MakeRequest()`, `Client` struct |
| `internal/config/config.go` | Config persistence, loading | `Load()`, `TestAndSave()`, `DefaultCloudURL` |
| `internal/models/models.go` | Data structures | `Peer`, `Group`, `Network`, `Policy`, `Config`, etc. |
| `internal/helpers/helpers.go` | Formatting, utilities, confirmations, colors | `ConfirmSingleDeletion()`, `ConfirmBulkDeletion()`, `Colorize()`, `IsTTY()` |
| `internal/commands/service.go` | Service wrapper for client | `Service` struct, `NewService()` |
| `internal/commands/usage.go` | Help text and usage functions | `PrintUsage()`, `PrintPeerUsage()`, etc. |
| `internal/commands/peers.go` | Peer operations | `HandlePeersCommand()`, list/inspect/remove/update peers |
| `internal/commands/groups.go` | Group operations | `HandleGroupsCommand()`, create/delete/rename groups |
| `internal/commands/networks.go` | Network operations | `HandleNetworkCommand()`, resources and routers |
| `internal/commands/policies.go` | Policy operations | `HandlePoliciesCommand()`, rules management |
| `internal/commands/export.go` | YAML export functionality | `HandleExportCommand()`, full/split export |
| `internal/commands/import.go` | YAML import functionality | `HandleImportCommand()`, conflict resolution |

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
// internal/client/client.go
type Client struct {
    Token         string
    ManagementURL string
    HTTPClient    *http.Client
    Debug         bool
}

client := client.New(config.Token, config.ManagementURL)
```

**2. Service Pattern**
Commands are organized via a Service struct that wraps the client:
```go
// internal/commands/service.go
type Service struct {
    Client *client.Client
}

svc := commands.NewService(c)
svc.HandlePeersCommand(args)
```

**3. Command Handler Pattern**
Each domain has a dedicated handler method on Service:
- `svc.HandlePeersCommand(args)`
- `svc.HandleGroupsCommand(args)`
- `svc.HandleNetworkCommand(args)`
- `svc.HandlePoliciesCommand(args)`
- `svc.HandleExportCommand(args)`
- `svc.HandleImportCommand(args)`

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
# Build binary
go build ./cmd/netbird-manage

# Result: ./netbird-manage (or netbird-manage.exe on Windows)
```

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o netbird-manage-linux-amd64 ./cmd/netbird-manage

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o netbird-manage-windows-amd64.exe ./cmd/netbird-manage

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o netbird-manage-darwin-arm64 ./cmd/netbird-manage

# macOS AMD64 (Intel)
GOOS=darwin GOARCH=amd64 go build -o netbird-manage-darwin-amd64 ./cmd/netbird-manage
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

### Confirmation Prompts

**All destructive operations require user confirmation** to prevent accidental data loss. The CLI implements two types of confirmation prompts:

**1. Single Resource Deletion** (Y/N prompt):
```go
details := map[string]string{
    "IP":        peer.IP,
    "Hostname":  peer.Hostname,
    "Connected": fmt.Sprintf("%t", peer.Connected),
}

if !confirmSingleDeletion("peer", peer.Name, peer.ID, details) {
    return nil // User cancelled
}
```

Output example:
```
About to remove peer:
  Name:      laptop-001
  ID:        abc123
  IP:        100.64.0.5
  Hostname:  laptop-001.local
  Connected: true

⚠️  This action cannot be undone. Continue? [y/N]:
```

**2. Bulk Operations** (Type-to-confirm):
```go
itemList := []string{
    "old-servers (ID: def456)",
    "test-group (ID: ghi789)",
}

if !confirmBulkDeletion("groups", itemList, len(itemList)) {
    return nil // User cancelled
}
```

Output example:
```
🔴 This will delete 2 groups:
  - old-servers (ID: def456)
  - test-group (ID: ghi789)

Type 'delete 2 groups' to confirm:
```

**Skipping Confirmations:**
- Global flag: `--yes` or `-y` (placed before command)
- Sets `skipConfirmation = true` (global variable in `helpers.go`)
- All confirmation functions check this flag first
- Usage: `netbird-manage --yes peer --remove abc123`

**Implementation Functions** (in `helpers.go`):
- `confirmSingleDeletion(resourceType, name, id string, details map[string]string) bool`
- `confirmBulkDeletion(resourceType string, items []string, count int) bool`
- `readYesNo() bool` (helper for Y/N input)

**All delete operations updated:**
1. Peers: `removePeerByID()`
2. Groups: `deleteGroup()`, `deleteUnusedGroups()`
3. Setup Keys: `deleteSetupKey()`, `deleteAllSetupKeys()`
4. Networks: `deleteNetwork()`, `removeNetworkResource()`, `removeNetworkRouter()`
5. Policies: `deletePolicy()`
6. Users: `removeUser()`
7. Tokens: `revokeToken()`
8. Routes: `deleteRoute()`
9. DNS: `deleteDNSGroup()`
10. Posture Checks: `deletePostureCheck()`
11. Accounts: `deleteAccount()`
12. Ingress Ports: `deleteIngressPort()`, `deleteIngressPeer()`

**Pattern for all deletions:**
```go
func (c *Client) deleteResource(id string) error {
    // 1. Fetch resource details (if not already available)
    resource, err := c.getResourceByID(id)
    if err != nil {
        return err
    }

    // 2. Build details map
    details := map[string]string{
        "Key1": "value1",
        "Key2": "value2",
    }

    // 3. Confirm deletion
    if !confirmSingleDeletion("resource", resource.Name, resource.ID, details) {
        return nil
    }

    // 4. Proceed with deletion
    resp, err := c.makeRequest("DELETE", "/resources/"+id, nil)
    // ... handle response
}
```

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

### Cleanup Operations

**Example: Delete unused groups**

The `--delete-unused` flag for the group command demonstrates how to safely delete resources that are no longer in use:

```bash
netbird-manage group --delete-unused
```

**Implementation approach (groups.go:522):**
1. **Scan all dependencies**: Fetch all resources that might reference groups (policies, setup keys, routes, DNS groups, users)
2. **Build reference map**: Create a map of all group IDs that are referenced
3. **Identify unused groups**: Find groups with no peers, no resources, and no references
4. **Show confirmation**: Display what will be deleted and require explicit confirmation
5. **Delete safely**: Delete groups one by one with error handling

**Key checks for "unused" groups:**
- `PeersCount == 0` (no peers in the group)
- `ResourcesCount == 0` (no resources in the group)
- Not referenced in any policy rules (Sources or Destinations)
- Not referenced in any setup keys (AutoGroups)
- Not referenced in any routes (Groups)
- Not referenced in any DNS nameserver groups (Groups)
- Not referenced in any users (AutoGroups)

This pattern can be applied to other resources that need cleanup based on dependency checking.

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

**4. Test Data & Credentials - DO NOT COMMIT**
- **NEVER** commit test results files (e.g., `TEST_RESULTS.md`, `test-output.txt`, etc.) to the repository
- **NEVER** commit API tokens or credentials in any files, including test files or documentation
- Test results should be kept locally or shared through other means (PR descriptions, issues, etc.)
- If test results need to be documented, redact all sensitive information (tokens, peer IDs, IP addresses, etc.)
- Add test result files to `.gitignore` if they are generated frequently
- Always review `git diff` before committing to ensure no sensitive data is included

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
- Minimum: Go 1.18
- Current: Go 1.24 (as specified in go.mod)
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
go build ./cmd/netbird-manage

# Show dependencies
go list -m all

# Tidy dependencies
go mod tidy
```

---

## Future Development Areas

### Feature Roadmap

This section tracks the implementation status of CLI features and planned enhancements.

#### ✅ Completed Features

**Core Resource Management:**
- ✅ **Peers** - Full read operations including list, inspect, remove, and group assignment
- ✅ **Groups** - Full CRUD operations including creation, deletion, renaming, and bulk peer management
- ✅ **Networks** - Full CRUD operations including resource and router management
- ✅ **Policies** - Full CRUD operations with advanced rule management including protocol/port configuration, bidirectional traffic, and group name resolution
- ✅ **Setup Keys** - Full CRUD operations for device registration and onboarding keys
  - Create, list, inspect, update, and delete setup keys
  - Support for one-off and reusable keys
  - Ephemeral peer support and auto-group assignment
  - Quick-create functionality for rapid key generation
  - **Implementation File:** `setup-keys.go`

**User & Access Management:**
- ✅ **Users** - Full user account management (6 API endpoints)
  - Invite users, manage roles and permissions
  - List, update, and remove users
  - Resend invitations and get current user info
  - Service user support for automation
  - Block/unblock user access
  - **Implementation File:** `users.go`

- ✅ **Tokens** - Personal access token management (4 API endpoints)
  - Create, list, revoke, and inspect API tokens
  - Essential for secure API access management
  - Automatic current user detection
  - Configurable expiration (1-365 days)
  - **Implementation File:** `tokens.go`

**Project Status:**
- **API Coverage:** 14/14 resource types fully implemented (100%) 🎉
- **Minimal Dependencies** - Only `gopkg.in/yaml.v3` for YAML export/import
- **Phase 1 Complete:** All high-priority user and access management features implemented
- **Phase 2 Complete:** Network services including routing, DNS, and posture checks
- **Phase 3 Complete:** Monitoring and analytics with audit logs, traffic events, and geo-location data
- **Phase 4 Complete:** Account management and ingress ports (Cloud-only) with peer updates
- **Phase 5 Complete:** GitOps with YAML export/import functionality

**Network Services (Phase 2 - COMPLETED):**
- ✅ **Routes** - Network routing configuration (5 API endpoints)
  - Define custom network routes, manage priorities, configure routing peers
  - Support for single peer or peer groups as routers
  - Metric-based priority control (1-9999, lower = higher priority)
  - Masquerading/NAT support for outbound traffic
  - **Implementation File:** `routes.go`

- ✅ **DNS** - DNS nameserver groups (6 API endpoints)
  - Create DNS nameserver groups, configure settings, manage domains
  - Domain-specific matching with wildcard support
  - Primary DNS group designation
  - Search domains configuration
  - Account-level DNS settings management
  - **Implementation File:** `dns.go`

**Security & Compliance (Phase 2 - COMPLETED):**
- ✅ **Posture Checks** - Device compliance validation (5 API endpoints)
  - Define compliance requirements (OS version, geolocation, network ranges, processes, NetBird version)
  - Enforce zero-trust security policies on peer groups
  - 5 check types: nb-version, os-version, geo-location, network-range, process
  - Platform-specific configuration for cross-platform support
  - **Implementation File:** `posture-checks.go`

#### 📋 Planned Features

**Developer Experience:**
- ❌ **Shell Completion** - Tab completion for bash/zsh/fish

#### 📊 Implementation Priority

**✅ Phase 1: Core Coverage (COMPLETED)**
1. ✅ Setup Keys (device onboarding - critical for operations)
2. ✅ Users management (critical for team management)
3. ✅ Tokens management (security and access control)

**✅ Phase 2: Network Services (COMPLETED)**
4. ✅ Routes management (network routing with metrics and masquerading)
5. ✅ DNS configuration (nameserver groups with domain matching)
6. ✅ Posture Checks (5 check types for zero-trust security)

**✅ Phase 3: Monitoring & Analytics (COMPLETED)**
7. ✅ Events (audit logs and network traffic monitoring)
8. ✅ Geo-Locations (country/city data for posture checks)
9. ✅ JSON output mode (implemented for events and geo-locations)

**✅ Phase 4: Account & Advanced Peer Features (COMPLETED)**
10. ✅ Peer update operations (SSH, login expiration, IP address)
11. ✅ Accessible peers query
12. ✅ Accounts management (full CRUD operations)
13. ✅ Ingress Ports (Cloud-only - port forwarding and ingress peers)

**✅ Phase 5: Safety & UX Enhancements (COMPLETED)**
14. ✅ Confirmation Prompts - Prevent accidental deletions with detailed resource info and Y/N prompts
15. ✅ Bulk deletion confirmations - Type-to-confirm for operations affecting multiple resources
16. ✅ Global `--yes` flag - Skip confirmations for automation and scripts
   - Implemented in `helpers.go`: `confirmSingleDeletion()`, `confirmBulkDeletion()`
   - Applied to all 16 delete operations across the codebase
   - Zero external dependencies (uses stdlib `bufio` and `fmt.Scanln`)

**✅ Phase 6: Quality of Life Enhancements (COMPLETED)**
17. ✅ Batch operations - Process multiple resources at once
   - Implemented `--remove-batch` for peers (peers.go:removePeersBatch)
   - Implemented `--delete-batch` for groups (groups.go:deleteGroupsBatch)
   - Implemented `--delete-batch` for setup keys (setup-keys.go:deleteSetupKeysBatch)
   - Features: Progress indicators, partial failure handling, summary reports
   - Uses same confirmation system (confirmBulkDeletion)
   - Zero external dependencies

18. ✅ Colorized output - Improve readability with ANSI color coding
   - Implemented in `colors.go` with automatic TTY detection
   - Color scheme: Headers (bold cyan), IDs (dim), Status (green/red), Success/Error/Warning indicators
   - Applied to peer and group list outputs
   - Pipe-friendly: Auto-disables when output is not a TTY
   - Zero external dependencies (pure ANSI escape codes)

19. ✅ Debug mode - Verbose HTTP request/response logging
   - Global `--debug` or `-d` flag (netbird-manage.go)
   - Client.Debug field enables logging in makeRequest (client.go)
   - Shows: HTTP method, URL, headers (token redacted), request/response bodies (pretty-printed JSON)
   - All debug output to stderr (keeps stdout clean for scripting)
   - Color-coded status codes
   - Zero external dependencies

**Phase 7: Developer Experience**
20. ❌ Shell completion - Tab completion for bash/zsh/fish

### Implementation Notes

**Dependencies:**
- Minimal external dependencies - only `gopkg.in/yaml.v3` for YAML export/import functionality
- All other features implemented using pure Go standard library
- YAML library chosen for reliability and widespread adoption

**API Coverage Status:**
- ✅ **100% Coverage:** All 14 NetBird API resource types fully implemented
  - Peers, Groups, Networks, Policies, Setup Keys, Users, Tokens
  - Routes, DNS, Posture Checks, Events, Geo-Locations
  - Accounts, Ingress Ports (Cloud-only)

**Code Architecture:**
- Each resource type gets its own file (`{resource}.go`)
- Follow existing patterns in `policies.go`, `groups.go`, `networks.go`
- Use flag-based command parsing (consistent with current implementation)
- Keep HTTP client methods on `Client` struct
- Maintain table output with `tabwriter` for consistency

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
- [ ] Code builds successfully (`go build ./cmd/netbird-manage`)
- [ ] No linting errors (`go vet ./...`, `go fmt ./...`)
- [ ] Manually tested all affected commands
- [ ] Updated README.md if user-facing changes
- [ ] Updated CLAUDE.md if architectural changes
- [ ] Commit messages follow guidelines
- [ ] No debug code or commented-out code
- [ ] No hardcoded tokens or credentials
- [ ] No test result files (TEST_RESULTS.md, test-output.txt, etc.)
- [ ] No testing API keys or sensitive data in any committed files

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

### Go Packages Used

**Standard Library:**
- `encoding/json` - JSON encoding/decoding
- `flag` - Command-line flag parsing
- `fmt` - Formatted I/O
- `io` - I/O primitives
- `net/http` - HTTP client/server
- `os` - Operating system interface
- `path/filepath` - File path manipulation
- `strings` - String utilities
- `text/tabwriter` - Tabular output formatting
- `time` - Time formatting and parsing
- `bytes` - Byte buffer operations
- `bufio` - Buffered I/O (for user input)

**External Dependencies:**
- `gopkg.in/yaml.v3` - YAML parsing and generation (for export/import functionality)

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

**Last Updated:** 2025-11-23
**Document Version:** 1.1
**Codebase Structure:** cmd/internal package layout with ~22 Go files
