# Contributing to LazyProxyFlare

Thank you for considering contributing to LazyProxyFlare! This document provides guidelines and information for developers.

---

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Code Style & Conventions](#code-style--conventions)
- [Testing Guidelines](#testing-guidelines)
- [Submitting Changes](#submitting-changes)
- [Areas for Contribution](#areas-for-contribution)

---

## Development Setup

### Prerequisites

- **Go 1.21+** ([install](https://go.dev/doc/install))
- **Docker** (for Caddy container restart testing)
- **Cloudflare Account** with API token (for integration testing)
- **Git** for version control

### Getting the Code

```bash
# Clone the repository
git clone https://github.com/yourusername/lazyproxyflare.git
cd lazyproxyflare

# Install dependencies
go mod download

# Build the project
go build -o lazyproxyflare ./cmd/lazyproxyflare

# Run the application
./lazyproxyflare
```

### Development Configuration

Create a test configuration file for development:

```bash
mkdir -p config
cp config.example.yaml config/test.yaml
# Edit config/test.yaml with your test credentials
```

**⚠️ Important:** Never commit `config/test.yaml` with real credentials!

---

## Project Structure

### Directory Layout

```
lazyproxyflare/
├── cmd/lazyproxyflare/          # Application entry point
│   └── main.go                  # Main function, initialization
│
├── internal/                    # Private application code
│   ├── audit/                   # Audit logging system
│   │   └── logger.go            # Log operations to file
│   │
│   ├── caddy/                   # Caddyfile operations
│   │   ├── manager.go           # Backup, restore, cleanup
│   │   ├── parser.go            # Parse Caddyfile syntax
│   │   └── types.go             # Data structures
│   │
│   ├── cloudflare/              # Cloudflare API client
│   │   ├── client.go            # HTTP client, API calls
│   │   └── types.go             # DNSRecord, API response types
│   │
│   ├── config/                  # Configuration management
│   │   ├── config.go            # YAML loading, validation
│   │   └── types.go             # Config structs
│   │
│   ├── diff/                    # DNS ↔ Caddy comparison engine
│   │   ├── engine.go            # Compare() logic
│   │   └── types.go             # SyncedEntry, SyncStatus
│   │
│   └── ui/                      # Terminal UI (Bubbletea)
│       ├── app.go               # Main Update() and View() logic
│       ├── model.go             # Application state
│       ├── forms.go             # Add/edit form rendering
│       ├── confirmations.go     # Confirmation dialogs
│       ├── backups.go           # Backup manager views
│       ├── views.go             # Panel and list rendering
│       ├── helpers.go           # Filtering & sorting
│       ├── colors.go            # Color palette constants
│       ├── panels.go            # Layout system
│       └── auditlog.go          # Audit log viewer
│
├── config.example.yaml          # Configuration template
├── README.md                    # Project documentation
├── KEYBINDINGS.md               # Keyboard shortcuts reference
├── CONTRIBUTING.md              # This file
└── go.mod                       # Go module dependencies
```

### Architecture Overview

LazyProxyFlare follows a **layered architecture**:

```
┌─────────────────────────────────────────┐
│      Presentation Layer (UI)            │  ← Bubbletea TUI
│      internal/ui/                       │
└────────────────┬────────────────────────┘
                 │
┌────────────────┴────────────────────────┐
│      Application Layer                  │  ← Business logic
│      internal/diff/                     │
└────────────────┬────────────────────────┘
                 │
┌────────────────┴────────────────────────┐
│      Domain Layer                       │  ← Core operations
│      internal/caddy/                    │
│      internal/cloudflare/               │
│      internal/audit/                    │
└────────────────┬────────────────────────┘
                 │
┌────────────────┴────────────────────────┐
│      Infrastructure Layer               │  ← External I/O
│      File I/O, HTTP client, Docker CLI  │
└─────────────────────────────────────────┘
```

**Design Principles:**
1. **Separation of Concerns:** Each layer has distinct responsibilities
2. **Dependency Direction:** Higher layers depend on lower layers, not vice versa
3. **Testability:** Business logic (diff, caddy, cloudflare) is UI-independent
4. **Immutability:** Bubbletea models are immutable (Update returns new state)

---

## Code Style & Conventions

### Go Code Standards

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting (automatically applied)
- Use `golint` for linting
- Keep functions small and focused (<50 lines when possible)
- Write descriptive comments for public functions

### Naming Conventions

- **Files:** Lowercase, underscores for multi-word (e.g., `audit_log.go`)
- **Types:** PascalCase (e.g., `SyncedEntry`, `DNSRecord`)
- **Functions:** camelCase for private, PascalCase for public
- **Constants:** ALL_CAPS for package-level, or PascalCase for enum-style

### Project-Specific Conventions

#### UI State Management (Bubbletea)

```go
// Good: Return new model, don't mutate
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    newModel := m  // Copy
    newModel.cursor++
    return newModel, nil
}

// Bad: Mutating receiver
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.cursor++  // Don't do this!
    return m, nil
}
```

#### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to parse Caddyfile: %w", err)
}

// Bad: Return raw error
if err != nil {
    return err
}
```

#### File Organization

- Group related functions together
- Put `Update()` and `View()` methods in `app.go`
- Keep rendering functions in separate files (`forms.go`, `views.go`, etc.)
- Use `// Section:` comments to mark logical sections in large files

---

## Testing Guidelines

### Current State

LazyProxyFlare has **14 test files** covering the `internal/caddy/` and `internal/ui/` packages. Additional test contributions are welcome, especially for `internal/cloudflare/`, `internal/config/`, and `internal/diff/`.

Run tests with: `go test ./...`

### Testing Strategy

#### 1. Unit Tests (Priority: HIGH)

Test pure business logic with no external dependencies:

```go
// Example: internal/diff/engine_test.go
func TestCompareSyncedEntries(t *testing.T) {
    dns := []cloudflare.DNSRecord{
        {Name: "test.example.com", Type: "CNAME", Content: "target.com"},
    }
    caddy := []caddy.CaddyEntry{
        {Domain: "test.example.com", Port: 80},
    }

    result := diff.Compare(dns, caddy)

    if len(result) != 1 {
        t.Errorf("Expected 1 entry, got %d", len(result))
    }
    if result[0].Status != diff.StatusSynced {
        t.Errorf("Expected synced, got %v", result[0].Status)
    }
}
```

**Target Coverage:**
- `internal/diff/` - 80%+
- `internal/caddy/parser.go` - 70%+
- `internal/config/` - 80%+

#### 2. Integration Tests (Priority: MEDIUM)

Test with mocked external dependencies:

```go
// Example: internal/cloudflare/client_test.go
func TestListDNSRecords(t *testing.T) {
    // Use httptest to mock Cloudflare API
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"result":[...]}`))
    }))
    defer server.Close()

    client := cloudflare.NewClient("test-token")
    client.BaseURL = server.URL  // Override base URL

    records, err := client.ListDNSRecords("test-zone", "CNAME")
    // Assert results...
}
```

#### 3. End-to-End Tests (Priority: LOW)

Manual testing checklist (see `MANUAL_TESTING.md` - to be created)

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/diff/

# Run with verbose output
go test -v ./...
```

---

## Submitting Changes

### Before You Start

1. **Check existing issues** - Someone might already be working on it
2. **Open an issue** - Discuss your idea before writing code
3. **Fork the repository** - Don't work directly on main

### Development Workflow

```bash
# 1. Create a feature branch
git checkout -b feature/your-feature-name

# 2. Make your changes
# ... edit code ...

# 3. Test your changes
go build -o lazyproxyflare ./cmd/lazyproxyflare
./lazyproxyflare  # Manual testing

# 4. Commit with descriptive message
git add .
git commit -m "feat: Add support for TXT records"

# 5. Push to your fork
git push origin feature/your-feature-name

# 6. Open a Pull Request on GitHub
```

### Commit Message Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code restructuring (no functional change)
- `test:` - Adding tests
- `chore:` - Maintenance tasks

**Examples:**
```
feat: Add support for TXT records
fix: Correct IPv4 validation in A record form
docs: Update README with multi-domain examples
refactor: Extract backup manager into separate file
test: Add unit tests for diff engine
```

### Pull Request Guidelines

1. **Title:** Use conventional commit format
2. **Description:** Explain what and why (not just how)
3. **Testing:** Describe how you tested the changes
4. **Screenshots:** Include for UI changes
5. **Breaking Changes:** Clearly mark any breaking changes

**PR Template:**

```markdown
## What does this PR do?

Brief description of the change.

## Why is this needed?

Explain the problem this solves or feature this adds.

## How was this tested?

- [ ] Manual testing with X scenarios
- [ ] Unit tests added/updated
- [ ] Tested on macOS/Linux/Windows

## Screenshots (if UI changes)

[Paste screenshots here]

## Checklist

- [ ] Code follows project conventions
- [ ] Documentation updated (if needed)
- [ ] No sensitive data (API keys, passwords) in code
- [ ] Builds successfully (`go build`)
```

### Review Process

1. Maintainer reviews PR (usually within 1-3 days)
2. Address any requested changes
3. Once approved, maintainer will merge
4. Your contribution will be in the next release!

---

## Areas for Contribution

### High Priority

1. **Testing Infrastructure**
   - Write unit tests for diff engine
   - Write parser tests with sample Caddyfiles
   - Mock Cloudflare API responses
   - Integration test suite

2. **Documentation** 
   - More usage examples
   - Video tutorials / GIFs
   - Common workflows guide
   - Troubleshooting section

3. **Bug Fixes**
   - Report bugs via GitHub Issues
   - Fix existing bugs

### Medium Priority

4. **Feature Enhancements**
   - Multi-domain profiles (v1.1)
   - TXT record support
   - Configurable keybindings
   - Theme customization

5. **Performance**
   - Optimize large Caddyfile parsing
   - Cache DNS records
   - Reduce API calls

6. **Distribution**
   - Create Homebrew formula
   - Create AUR package (Arch Linux)
   - GoReleaser configuration
   - Installation scripts

### Nice to Have

7. **UI/UX Improvements**
   - Additional color themes
   - Customizable status bar
   - Configuration wizard (first-run)

8. **New Features** 
   - Other DNS providers (Route53, etc.)
   - Other reverse proxies (Traefik, nginx)
   - Remote Caddyfile support (SSH)

---

## Getting Help

- **Questions?** Open a [Discussion](https://github.com/yourusername/lazyproxyflare/discussions)
- **Found a bug?** Open an [Issue](https://github.com/yourusername/lazyproxyflare/issues)
- **Feature idea?** Open an [Issue](https://github.com/yourusername/lazyproxyflare/issues) with `[Feature Request]`

---

## Code of Conduct

Be respectful, inclusive, and constructive. We're all here to build something useful together.

- **Be kind** - We're all learning
- **Be patient** - Maintainers are volunteers
- **Be constructive** - Focus on solutions, not problems

---

## Recognition

All contributors will be acknowledged in:
- `README.md` contributors section
- Release notes
- GitHub contributors graph

Thank you for making LazyProxyFlare better!

---

**Last Updated:** 2025-12-28
**Maintained by:** [Your Name]
