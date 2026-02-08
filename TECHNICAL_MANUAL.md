# LazyProxyFlare Technical Manual

A comprehensive guide to understanding the lazydns codebase for Go learners.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Bubbletea TUI Framework](#2-bubbletea-tui-framework)
3. [Configuration System](#3-configuration-system)
4. [Keyboard & Wizard Systems](#4-keyboard--wizard-systems)
5. [Cloudflare & Caddy Integration](#5-cloudflare--caddy-integration)
6. [Go Idioms & Patterns](#6-go-idioms--patterns)

---

## 1. Architecture Overview

### Package Structure

```
lazyproxyflare/
├── cmd/lazyproxyflare/      # Entry point
└── internal/
    ├── ui/                  # User interface (Bubbletea TUI)
    ├── config/              # Configuration management
    ├── cloudflare/          # Cloudflare API client
    ├── caddy/               # Caddy config parsing & management
    ├── diff/                # DNS/Caddy sync comparison engine
    └── audit/               # Audit logging
```

### Startup Flow

```go
// cmd/lazyproxyflare/main.go
func main() {
    profiles, _ := config.ListProfiles()

    switch len(profiles) {
    case 0:
        launchWizard()           // No profiles - setup wizard
    case 1:
        autoLoadProfile(profiles[0])  // Single profile - auto-load
    default:
        launchProfileSelector(profiles)  // Multiple - choose
    }
}
```

### Data Flow

```
Caddyfile (disk) → ParseCaddyfileWithSnippets() → []CaddyEntry
                                                       ↓
Cloudflare API → ListDNSRecords() → []DNSRecord → diff.Compare()
                                                       ↓
                                              []SyncedEntry → UI
```

### Key Design Principles

| Principle | Implementation |
|-----------|---------------|
| **Separation of Concerns** | UI, business logic, and I/O in separate packages |
| **Elm Architecture** | Model-Update-View pattern in UI layer |
| **Fail Fast** | Validate on load/save, prevent corrupted states |
| **Atomic Operations** | Backup → Modify → Validate → Reload with rollback |

---

## 2. Bubbletea TUI Framework

### The Elm Architecture (Model-Update-View)

```
User Input (KeyMsg, MouseMsg)
    ↓
Update() → processes message → returns (Model, Cmd)
    ↓
View() → renders state to string
    ↓
Cmd executes async → returns Message
    ↓
[Loop continues]
```

### Model Struct (`internal/ui/model.go`)

```go
type Model struct {
    // Data
    entries  []diff.SyncedEntry      // All domains with sync status
    snippets []caddy.Snippet         // Available Caddy snippets
    config   *config.Config

    // UI State
    currentView ViewMode             // Which view is active
    cursor      int                  // Selected item
    width, height int                // Terminal dimensions

    // Forms
    addForm          AddFormData
    wizardData       WizardData
    selectedEntries  map[string]bool // Multi-select
}
```

**Key Insight**: All state lives in ONE struct. No globals, no hidden state.

### ViewMode Enum (23 states)

```go
type ViewMode int

const (
    ViewList ViewMode = iota     // Main domain list
    ViewDetails                  // Selected domain details
    ViewAdd                      // Add entry form
    ViewEdit                     // Edit entry form
    ViewWizard                   // Setup wizard
    ViewSnippetWizard           // Snippet creation
    ViewBackupManager           // Restore from backups
    ViewAuditLog                // Operation history
    // ... 15 more views
)
```

### Value Receivers for Immutability

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // m is a COPY - modifications don't affect caller
    m.cursor = newValue
    return m, cmd  // Return the modified copy
}
```

**Why Value Receivers?**
- Each Update gets a fresh copy
- Updates are explicit via return values
- Forces immutability mindset
- Easier to reason about state changes

### Async Commands (tea.Cmd)

Commands run async operations without blocking the UI:

```go
func refreshDataCmd(cfg *config.Config) tea.Cmd {
    return func() tea.Msg {  // Returns a function
        // This runs in a goroutine
        caddyContent, _ := os.ReadFile(cfg.Caddy.CaddyfilePath)
        parsed := caddy.ParseCaddyfileWithSnippets(string(caddyContent))

        cfClient := cloudflare.NewClient(apiToken)
        records, _ := cfClient.ListDNSRecords(zoneID, "CNAME")

        entries := diff.Compare(records, parsed.Entries)

        return refreshCompleteMsg{  // Message goes back to Update()
            entries:  entries,
            snippets: parsed.Snippets,
        }
    }
}
```

**Pattern**:
1. Command function returns a closure
2. Bubbletea runs closure in goroutine
3. Closure does I/O (file reads, API calls)
4. Returns a message with results
5. Update() receives message and updates state

---

## 3. Configuration System

### Two-Layer Architecture

```
ProfileConfig (on disk - YAML)    →    Config (in memory)
         ↓                                    ↓
   ~/.config/lazyproxyflare/        Used by rest of app
   profiles/homelab.yaml
```

### ProfileConfig (Modern Format)

```go
type ProfileConfig struct {
    Profile    ProfileMetadata  `yaml:"profile"`
    Cloudflare CloudflareConfig `yaml:"cloudflare"`
    Domain     string           `yaml:"domain"`
    Proxy      ProxyConfig      `yaml:"proxy"`      // Nested for extensibility
    Defaults   DefaultsConfig   `yaml:"defaults"`
    UI         UIConfig         `yaml:"ui"`
}

type ProxyConfig struct {
    Type       ProxyType        `yaml:"type"`       // "caddy"
    Deployment DeploymentMethod `yaml:"deployment"` // "docker" or "system"
    Caddy      CaddyProxyConfig `yaml:"caddy,omitempty"`
}
```

### Config (Legacy Format)

```go
type Config struct {
    Cloudflare CloudflareConfig `yaml:"cloudflare"`
    Domain     string           `yaml:"domain"`
    Caddy      CaddyConfig      `yaml:"caddy"`      // Flat structure
    Defaults   DefaultsConfig   `yaml:"defaults"`
    UI         UIConfig         `yaml:"ui"`
}
```

### Conversion Bridge

```go
func ProfileToLegacyConfig(profile *ProfileConfig) *Config {
    return &Config{
        Cloudflare: profile.Cloudflare,
        Domain:     profile.Domain,
        Caddy: CaddyConfig{
            CaddyfilePath:   profile.Proxy.Caddy.CaddyfilePath,
            ContainerName:   profile.Proxy.Caddy.ContainerName,
            CaddyBinaryPath: profile.Proxy.Caddy.CaddyBinaryPath,
            // Compute validation command based on deployment method
            ValidationCommand: profile.Proxy.Caddy.GetValidationCommand(profile.Proxy.Deployment),
        },
        Defaults: profile.Defaults,
        UI:       profile.UI,
    }
}
```

**Why Two Formats?**
- `ProfileConfig`: Forward-compatible (can add nginx, traefik later)
- `Config`: Backward-compatible (existing UI code unchanged)
- One-way bridge at load time

### Environment Variable Expansion

```go
func expandEnvVars(value string) string {
    // Only expands ${VAR_NAME} format
    if !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
        return value  // Plain text - return as-is
    }

    varName := value[2:len(value)-1]  // Extract "VAR_NAME"
    return os.Getenv(varName)
}
```

**Usage in YAML**:
```yaml
cloudflare:
  api_token: "${CLOUDFLARE_API_TOKEN}"  # References env var
  zone_id: "abc123..."
```

### Three-Layer Validation

1. **Syntax**: Character validation (regex)
2. **Completeness**: Required fields present
3. **Semantics**: File paths exist, no conflicts

```go
func validateProfile(p *ProfileConfig) error {
    // Layer 1: Required fields
    if p.Cloudflare.APIToken == "" {
        return fmt.Errorf("cloudflare.api_token is required")
    }

    // Layer 2: Conditional requirements
    if p.Proxy.Deployment == DeploymentDocker && p.Proxy.Caddy.ContainerName == "" {
        return fmt.Errorf("container_name required for Docker deployment")
    }

    // Layer 3: Format validation
    if !isValidZoneID(p.Cloudflare.ZoneID) {
        return fmt.Errorf("zone_id must be 32 hex characters")
    }

    return nil
}
```

---

## 4. Keyboard & Wizard Systems

### Key Handler Hierarchy

```go
func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
    // CRITICAL: Text input FIRST (before keybindings)
    if m.currentView == ViewWizard && m.isWizardTextInputStep() {
        key := msg.String()
        // Let navigation keys through
        if key != "esc" && key != "enter" && key != "tab" {
            var cmd tea.Cmd
            m.wizardTextInput, cmd = m.wizardTextInput.Update(msg)
            return m, cmd
        }
    }

    // Then handle navigation keys
    switch msg.String() {
    case "q":
        return m, tea.Quit
    case "j", "down":
        m.cursor++
    case "k", "up":
        m.cursor--
    // ...
    }
}
```

**Pattern**: Text input is prioritized over navigation to prevent 'q' from quitting when typing.

### Wizard State Machine

```go
type WizardStep int

const (
    WizardStepWelcome WizardStep = iota   // 0
    WizardStepBasicInfo                   // 1
    WizardStepCloudflare                  // 2
    WizardStepDockerConfig                // 3
    WizardStepSummary                     // 4
)

type WizardField int

// Per-step field enums
const (
    FieldProfileName WizardField = iota
    FieldDomain
)

const (
    FieldDeploymentMethod WizardField = iota
    FieldComposeFilePath
    FieldContainerName
    FieldCaddyfilePath
    FieldCaddyBinaryPath
)
```

### Navigation with State-Aware Logic

```go
func (m Model) getNextCaddyField(current WizardField) WizardField {
    if m.wizardData.DeploymentMethod == config.DeploymentSystem {
        // System deployment: DeploymentMethod → CaddyfilePath → CaddyBinaryPath
        switch current {
        case FieldDeploymentMethod:
            return FieldCaddyfilePath
        case FieldCaddyfilePath:
            return FieldCaddyBinaryPath
        }
    } else {
        // Docker deployment: different sequence
        switch current {
        case FieldDeploymentMethod:
            if m.wizardData.DockerMethod == "compose" {
                return FieldComposeFilePath
            }
            return FieldContainerName
        }
    }
    return current
}
```

### Save-Before-Move Pattern

```go
func (m Model) handleWizardNextField() (Model, tea.Cmd) {
    // Save current field BEFORE moving
    m.saveCurrentFieldValue()

    // Then advance to next field
    m.wizardData.CurrentField = m.getNextField()
    m.configureTextInputForField()

    return m, nil
}

func (m *Model) saveCurrentFieldValue() {
    value := m.wizardTextInput.Value()

    switch m.wizardStep {
    case WizardStepBasicInfo:
        switch m.wizardData.CurrentField {
        case FieldProfileName:
            m.wizardData.ProfileName = value
        case FieldDomain:
            m.wizardData.Domain = value
        }
    }
}
```

---

## 5. Cloudflare & Caddy Integration

### Cloudflare HTTP Client

```go
type Client struct {
    apiToken   string
    httpClient *http.Client
}

func NewClient(apiToken string) *Client {
    return &Client{
        apiToken: apiToken,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *Client) ListDNSRecords(zoneID, recordType string) ([]DNSRecord, error) {
    url := fmt.Sprintf("%s/zones/%s/dns_records?type=%s&page=%d&per_page=100",
        apiBaseURL, zoneID, recordType, page)

    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+c.apiToken)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    // Parse response...
}
```

### Automatic Pagination

```go
var allRecords []DNSRecord
page := 1

for {
    records, resultInfo, err := c.fetchPage(zoneID, recordType, page)
    if err != nil {
        return nil, err
    }

    allRecords = append(allRecords, records...)

    if resultInfo == nil || page >= resultInfo.TotalPages {
        break
    }
    page++
}
return allRecords, nil
```

### Caddyfile Parsing: Quote-Aware Brace Counting

```go
func countBracesOutsideQuotes(line string) (open, close int) {
    inQuote := false
    quoteChar := rune(0)
    escaped := false

    for _, ch := range line {
        if escaped {
            escaped = false
            continue
        }
        if ch == '\\' {
            escaped = true
            continue
        }

        // Track quote state
        if !inQuote && (ch == '"' || ch == '\'') {
            inQuote = true
            quoteChar = ch
            continue
        }
        if inQuote && ch == quoteChar {
            inQuote = false
            continue
        }

        // Only count braces outside quotes
        if !inQuote {
            if ch == '{' { open++ }
            if ch == '}' { close++ }
        }
    }
    return
}
```

### Three-Map Diff Algorithm

```go
func Compare(dnsRecords []cloudflare.DNSRecord, caddyEntries []caddy.CaddyEntry) []SyncedEntry {
    // Map 1: DNS records by domain
    dnsMap := make(map[string]*cloudflare.DNSRecord)
    for i := range dnsRecords {
        domain := strings.ToLower(dnsRecords[i].Name)
        dnsMap[domain] = &dnsRecords[i]
    }

    // Map 2: Caddy entries by domain
    caddyMap := make(map[string]*caddy.CaddyEntry)
    for i := range caddyEntries {
        for _, domain := range caddyEntries[i].Domains {
            caddyMap[strings.ToLower(domain)] = &caddyEntries[i]
        }
    }

    // Set 3: All unique domains
    allDomains := make(map[string]bool)
    for domain := range dnsMap { allDomains[domain] = true }
    for domain := range caddyMap { allDomains[domain] = true }

    // Compare
    var results []SyncedEntry
    for domain := range allDomains {
        dns := dnsMap[domain]
        caddy := caddyMap[domain]

        status := StatusSynced
        if dns == nil { status = StatusOrphanedCaddy }
        if caddy == nil { status = StatusOrphanedDNS }

        results = append(results, SyncedEntry{
            Domain: domain,
            DNS:    dns,
            Caddy:  caddy,
            Status: status,
        })
    }
    return results
}
```

**Complexity**: O(d + c + s) where d=DNS records, c=Caddy entries, s=unique domains.

### Safe File Modification Pattern

```go
func AppendEntry(caddyfilePath string, block string) error {
    // Step 1: Preserve permissions
    fileInfo, _ := os.Stat(caddyfilePath)
    originalPerms := fileInfo.Mode().Perm()

    // Step 2: Read current content
    content, err := os.ReadFile(caddyfilePath)
    if err != nil {
        return fmt.Errorf("failed to read Caddyfile: %w", err)
    }

    // Step 3: Modify in memory
    newContent := string(content) + "\n" + block

    // Step 4: Atomic write with original permissions
    return os.WriteFile(caddyfilePath, []byte(newContent), originalPerms)
}
```

---

## 6. Go Idioms & Patterns

### Error Wrapping with Context

```go
if err != nil {
    return nil, fmt.Errorf("failed to read response: %w", err)
}
```

**Why `%w`?**
- Creates error chain (preserves original)
- Supports `errors.Is()` and `errors.Unwrap()`
- Better debugging with full context

### iota for Enums

```go
type SyncStatus int

const (
    StatusSynced        SyncStatus = iota  // 0
    StatusOrphanedDNS                      // 1
    StatusOrphanedCaddy                    // 2
)

func (s SyncStatus) String() string {
    switch s {
    case StatusSynced:
        return "Synced"
    case StatusOrphanedDNS:
        return "Orphaned (DNS)"
    case StatusOrphanedCaddy:
        return "Orphaned (Caddy)"
    default:
        return "Unknown"
    }
}
```

**Benefits**:
- Auto-incrementing values
- Compile-time constants
- `String()` implements `fmt.Stringer`

### strings.Builder for Concatenation

```go
// SLOW (O(n²))
var result string
for _, item := range items {
    result = result + item  // Creates new string each time
}

// FAST (O(n))
var b strings.Builder
for _, item := range items {
    b.WriteString(item)  // Appends to buffer
}
return b.String()
```

### Pointer Semantics in Maps

```go
// Store pointers for efficiency
dnsMap := make(map[string]*cloudflare.DNSRecord)
for i := range dnsRecords {
    record := &dnsRecords[i]  // Pointer to array element
    dnsMap[domain] = record   // Store pointer, not copy
}
```

**Why?**
- No memory copy of structs
- Maps store 8-byte pointers
- Original data stays in slice

### Map as Set

```go
// Go has no set type, use map[T]bool
allDomains := make(map[string]bool)
for domain := range dnsMap {
    allDomains[domain] = true  // Deduplicates automatically
}
```

### Closure Helpers

```go
// Helper function scoped to this function
hasSnippet := func(name string) bool {
    for _, s := range input.SelectedSnippets {
        if s == name {
            return true
        }
    }
    return false
}

if !hasSnippet("ip_restricted") {
    // ...
}
```

### Defer for Cleanup

```go
file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0644)
if err != nil {
    return err
}
defer file.Close()  // Guaranteed cleanup

// ... use file
```

### Struct Tags for Serialization

```go
type DNSRecord struct {
    ID      string `json:"id"`
    Type    string `json:"type"`
    Name    string `json:"name"`
    Content string `json:"content"`
    Proxied bool   `json:"proxied"`
}

type LogEntry struct {
    Timestamp time.Time              `json:"timestamp"`
    Details   map[string]interface{} `json:"details,omitempty"`  // Omit if nil
}
```

---

## Learning Path

**Recommended reading order:**

1. `cmd/lazyproxyflare/main.go` - Understand startup flow
2. `internal/config/types.go` - Config structure
3. `internal/diff/engine.go` - Simple business logic
4. `internal/cloudflare/client.go` - HTTP API pattern
5. `internal/caddy/parser.go` - String parsing
6. `internal/ui/model.go` - State management
7. `internal/ui/app.go` - Elm architecture

---

## Summary: Design Principles

| Principle | Why |
|-----------|-----|
| **Fail Fast** | Validate early, prevent bad states |
| **Explicit State** | All state in Model struct, no globals |
| **Error Chains** | `%w` wrapping preserves context |
| **Atomic Operations** | Backup → Modify → Validate → Reload |
| **Separation** | UI, business logic, I/O in separate packages |
| **Immutability** | Value receivers, return modified copies |
| **Simple > Clever** | Straightforward code over abstractions |

---

*Generated from codebase analysis - February 2026*
