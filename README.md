# LazyProxyFlare

### WORK IN PROGRESS -  COMING SOON!!

> A fast, good looking terminal UI for managing Cloudflare DNS records and reverse proxy configurations in perfect sync.

---

## What is LazyProxyFlare?

LazyProxyFlare is a terminal UI (TUI) application inspired by [lazygit](https://github.com/jesseduffield/lazygit) that makes managing DNS records and reverse proxy configurations as intuitive as version control. Born from the frustration of manually keeping Cloudflare DNS and Caddy configurations synchronized, it provides a unified interface to view, create, update, and sync entries with confidence.

**Perfect for:**
- Homelab enthusiasts managing multiple services
- Self-hosters running Caddy reverse proxy
- Anyone who prefers terminal/keyboard-driven workflows

---

## Features

### Core Functionality
- **Dual DNS Type Support**: Manage both CNAME and A records
- **DNS-Only Mode**: Create DNS records without Caddy configuration
- **Real-time Sync Detection**: Visual indicators for synced, orphaned, and mismatched entries
- **CRUD Operations**: Create, Read, Update, Delete entries with validation
- **Smart Rollback**: Automatic rollback on failures (DNS + Caddyfile + container)
- **Automatic Backups**: Caddyfile backed up before every modification

### Profile & Configuration
- **Multi-Profile Support**: Manage multiple domains with separate profiles
- **Setup Wizard**: Interactive wizard for first-time configuration
- **Profile Switching**: Quick profile switching with `p` keybinding
- **Profile Selector**: Modal interface for choosing between multiple profiles
- **Flexible Deployment**: Support for Docker, systemd, and binary deployments
- **Custom Commands**: Configure custom validation and restart commands per profile

### Advanced Features
- **Multi-Level Filtering**: Filter by sync status, DNS type, or search query
- **Flexible Sorting**: Alphabetical or by sync status
- **Batch Operations**: Multi-select and bulk delete/sync
- **Backup Manager**: View, restore, delete, and auto-cleanup old backups
- **Audit Logging**: Complete operation history with timestamps
- **Multi-Panel Layout**: Lazygit-style interface with context-sensitive keybindings
- **Full Mouse Support**: Click, scroll, and navigate with mouse or keyboard
- **Snippet System**: Reusable configuration blocks for DRY Caddyfile management
  - **Snippet Wizard**: Interactive creation of IP restriction, security headers, and performance snippets
  - **Smart Forms**: Contextual snippet suggestions in add/edit forms
  - **Brownfield Support**: Automatic detection and editing of existing snippet imports
  - **Visual Indicators**: Color-coded category badges in three-panel layout

### Safety & Validation
- **Pre-flight Validation**: Caddy config validation before commit
- **Atomic Operations**: All-or-nothing updates with automatic rollback
- **Backup System**: 30-day retention with manual cleanup option
- **Input Validation**: IPv4, domain, and CIDR format checking
- **Confirmation Dialogs**: All destructive operations require confirmation

---

## Installation

### Prerequisites
- Go 1.21+ (for building from source)
- Cloudflare API token with DNS edit permissions
- Caddy server running in Docker
- Docker socket access (for container restart)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/tinkermesomething/lazyproxyflare.git
cd lazyproxyflare

# Build the binary
go build -o lazyproxyflare ./cmd/lazyproxyflare

# Move to PATH (very optional)
sudo mv lazyproxyflare /usr/local/bin/
```

### Quick Start

**Option 1: Setup Wizard (Recommended for first-time users)**

Simply run LazyProxyFlare and the interactive wizard will guide you through the setup:

```bash
lazyproxyflare
```

The wizard will ask you for:
- Profile name (e.g., homelab, production)
- Domain name
- Cloudflare API token and zone ID
- Reverse proxy type (Caddy, nginx, traefik - currently Caddy only)
- Deployment method (Docker, systemd, binary - currently Docker only)
- Caddyfile path and container name
- Optional custom validation command
- Default settings (CNAME target, port, SSL, proxied)

**Option 2: Manual Configuration**

1. **Create configuration directory**

```bash
mkdir -p ~/.config/lazyproxyflare/profiles
```

2. **Create a profile YAML file** (e.g., `~/.config/lazyproxyflare/profiles/homelab.yaml`):

```yaml
profile:
  name: "homelab"
  description: "Home lab environment"
  created: "2025-01-01T00:00:00Z"

cloudflare:
  api_token: "your_cloudflare_api_token"
  zone_id: "your_zone_id_32_hex_chars"

domain: "example.com"

proxy:
  type: "caddy"
  deployment: "docker"
  caddy:
    caddyfile_path: "/etc/caddy/Caddyfile"
    container_name: "caddy"
    # Optional: custom validation command
    validation_command: "docker exec {container} caddy validate --config {path}"

defaults:
  cname_target: "example.com"
  proxied: true
  port: 80
  ssl: false
```

3. **Run LazyProxyFlare**

```bash
lazyproxyflare
```

---

## Usage

### Main Interface

```
┌─ LazyProxyFlare ──────────────────────────────────────────────────────────┐
│                                                                           │
│ ┌─ Entries (15)  ──────────────┐ ┌─ Details ────────────────────────────┐ │
│ │                              │ │                                      │ │
│ │ ✓ plex.example.com           │ │ Domain: plex.example.com             │ │
│ │ ✓ grafana.example.com        │ │ DNS Type: CNAME                      │ │
│ │ ⚠ test.example.com           │ │ Target: mail.example.com             │ │
│ │                              │ │ Proxied: Yes                         │ │
│ │                              │ │ Port: 32400                          │ │
│ │                              │ │ SSL: Enabled                         │ │
│ └──────────────────────────────┘ └──────────────────────────────────────┘ │
│                                                                           │
│ a:add  e:edit  d:delete  s:sync  f:filter  t:type  o:sort  ?:help  q:quit │
└───────────────────────────────────────────────────────────────────────────┘
```

### Keyboard Shortcuts

#### Navigation
- `j/k` or `↑/↓` - Navigate up/down
- `g` - Jump to top
- `G` - Jump to bottom
- `Tab` - Cycle panels (List → Details → Snippets)
- `Shift+Tab` - Cycle panels backward

#### Actions
- `a` - Add new entry
- `e` - Edit selected entry
- `d` - Delete selected entry
- `s` - Sync orphaned entry (create missing DNS or Caddy)
- `r` - Refresh data from Cloudflare and Caddyfile
- `w` - Launch snippet creation wizard

#### Filtering & Search
- `f` - Cycle status filter (All → Synced → Orphaned DNS → Orphaned Caddy)
- `t` - Cycle DNS type filter (All → CNAME → A)
- `o` - Cycle sort mode (Alphabetical ↔ By Status)
- `/` - Search/filter by domain name
- `ESC` - Clear all filters and reset to default view

#### Batch Operations
- `Space` - Toggle selection checkbox
- `X` - Delete all selected entries
- `S` - Sync all selected entries
- `D` - Bulk delete menu (all orphaned DNS or Caddy)

#### Tools
- `b` - Backup manager (view, restore, cleanup)
- `l` - View audit log
- `p` or `Ctrl+P` - Profile selector (switch profiles)
- `?` - Help screen
- `q` - Quit

#### Mouse Support
- Click to select entries
- Click checkboxes to toggle selection
- Scroll wheel to navigate lists
- Click form fields to focus

---

## Configuration

### Profile System

LazyProxyFlare uses a profile-based configuration system that allows you to manage multiple domains and environments. Profiles are stored as YAML files in `~/.config/lazyproxyflare/profiles/`.

**Startup Behavior:**
- **No profiles** → Setup wizard launches automatically
- **One profile** → Auto-loads the profile and starts
- **Multiple profiles** → Profile selector modal appears

**Creating Profiles:**
1. **Setup Wizard** (Recommended): Run `lazyproxyflare` and follow the interactive wizard
2. **Manual**: Create a YAML file in `~/.config/lazyproxyflare/profiles/`
3. **Profile Selector**: Press `p` in main view, then `+` or `n` to launch wizard

### Profile File Structure

```yaml
profile:
  name: "homelab"
  description: "Home lab environment"
  created: "2025-01-01T00:00:00Z"

cloudflare:
  api_token: "your_api_token_here" #TODO: encrypt this thing. 
  zone_id: "your_zone_id_32_hex_chars"

domain: "example.com"

proxy:
  type: "caddy"              # caddy, nginx, traefik (only caddy implemented)
  deployment: "docker"        # docker, systemd, binary (only docker implemented)
  caddy:
    caddyfile_path: "/etc/caddy/Caddyfile"
    container_name: "caddy"
    # Optional: Custom validation command (supports {path} and {container} placeholders)
    validation_command: "docker exec {container} caddy validate --config {path}"
    # Optional: Custom restart command
    restart_command: "docker restart {container}"

defaults:
  cname_target: "mail.example.com"  # Default CNAME target for new entries
  proxied: true                     # Enable Cloudflare proxy by default
  port: 80                          # Default reverse proxy port
  ssl: false                        # Use HTTPS for upstream connections
  lan_subnet: "192.168.1.0/24"        # Optional: LAN subnet for IP restrictions
  allowed_external_ip: "1.2.3.4"   # Optional: Allowed external IP

ui:
  theme: "default"
```

### Profile Management

**Switching Profiles:**
- Press `p` or `Ctrl+P` to open profile selector
- Navigate with `j/k` or arrow keys
- Press `Enter` to switch to selected profile
- Data reloads automatically after switching

**Creating Additional Profiles:**
- From profile selector: Press `+` or `n` to launch wizard
- From main view: Press `p` to open selector, then `+`
- Wizard guides you through all configuration steps

**Last Used Profile:**
LazyProxyFlare remembers the last profile you used and highlights it in the selector.

---

## Snippet System

LazyProxyFlare includes a powerful snippet system for managing reusable Caddy configuration blocks. Snippets help you maintain DRY (Don't Repeat Yourself) configurations and apply consistent settings across multiple entries.

### What are Snippets?

Snippets are named configuration blocks that can be imported into multiple Caddy domain blocks using the `import` directive. Instead of repeating the same configuration (like IP restrictions or security headers) in every entry, you define it once as a snippet and import it where needed.

**Example:**
```
# Define snippet
(ip_restricted) {
    @external {
        not remote_ip 10.0.28.0/24
    }
    respond @external 404
}

# Use snippet
plex.example.com {
    import ip_restricted
    reverse_proxy http://localhost:32400
}
```

### Snippet Wizard

Press `w` to launch the interactive snippet creation wizard:

1. **Welcome**: Learn about snippets and what can be created
2. **IP Restriction**: Configure LAN subnet and optional external IP allowlist
3. **Security Headers**: Choose from basic, strict, or paranoid presets
4. **Performance**: Enable gzip and zstd compression
5. **Summary**: Review and create selected snippets

The wizard includes:
- Live preview of generated snippet code
- Duplicate detection (won't overwrite existing snippets)
- Automatic Caddy validation before writing
- Full rollback on validation failure

### Available Snippet Types

**IP Restriction (Access Control)**
- Restricts access to LAN subnet
- Optional external IP allowlist
- Returns 404 to unauthorized IPs

**Security Headers (Security)**
- **Basic**: X-Content-Type-Options, X-Frame-Options, removes Server header
- **Strict**: Adds CSP, Referrer-Policy, Permissions-Policy
- **Paranoid**: Adds HSTS, X-XSS-Protection with strictest settings

**Performance (Performance)**
- Enables gzip compression
- Enables zstd compression
- Improves load times and reduces bandwidth

### Using Snippets in Forms

When adding or editing entries, snippets appear in the form below the standard checkboxes:

- **Color-coded badges** show snippet category
- **Smart suggestions** recommend relevant snippets:
  - `ip_restricted` suggested when LANOnly is enabled
  - `security_headers` suggested when SSL is enabled
  - `performance` always suggested (beneficial for all)
- **Space bar** toggles snippet selection
- **Preview** shows final Caddyfile with `import` statements

### Three-Panel Layout

The main interface includes a dedicated snippets panel:

```
┌─ Entries ─────┬─ Details ──────┬─ Snippets ────────┐
│               │                │                   │
│ ✓ plex        │ Domain: plex   │ [Security]        │
│ ✓ grafana     │ Target: :32400 │ security_headers  │
│               │ Snippets:      │ (used by 5)       │
│               │ ip_restricted  │                   │
│               │ security...    │ [Performance]     │
│               │                │ performance       │
│               │                │ (used by 8)       │
└───────────────┴────────────────┴───────────────────┘
```

Navigate with `Tab` and `Shift+Tab`. Press `Enter` when focused on snippets panel to view snippet details.

### Brownfield Integration

LazyProxyFlare automatically detects existing snippet imports in your Caddyfile:

- **Details Panel**: Shows colored badges for applied snippets
- **Edit Form**: Pre-selects snippets currently used by the entry
- **Usage Tracking**: Shows which entries use each snippet
- **Seamless Updates**: Add or remove snippets through the form interface

**Workflow:**
1. Select entry with existing snippets
2. View colored badges in details panel
3. Press `e` to edit
4. Snippet checkboxes show current state
5. Toggle snippets on/off
6. Preview shows updated `import` statements
7. Save applies changes

---

## Workflows

### Adding a New Service

1. Press `a` to open the add entry form
2. Fill in the subdomain (e.g., "plex")
3. Choose DNS type (CNAME or A)
4. Enter target (domain for CNAME, IP for A record)
5. Configure Caddy options (port, SSL, restrictions)
6. Press `Enter` to preview
7. Confirm with `y` to create

**Result:** DNS record created in Cloudflare + Caddy block added + container restarted

### Syncing Orphaned Entries

**Scenario:** You added a Caddy configuration manually but forgot to create the DNS record.

1. Press `f` to filter → select "Orphaned Caddy"
2. Navigate to the entry showing `⚠ Orphaned (Caddy)`
3. Press `s` to sync
4. Choose "Create DNS record from Caddy config"
5. Confirm with `y`

**Result:** DNS record created automatically with settings from Caddyfile

### Batch Cleanup

**Scenario:** Clean up 10 old DNS records that no longer have Caddy configs.

1. Press `f` → select "Orphaned DNS"
2. Press `Space` on each entry to select
3. Press `X` to batch delete selected
4. Review the list and confirm with `y`

**Result:** All selected DNS records deleted in one operation

### Backup Management

1. Press `b` to open backup manager
2. Navigate with `j/k`
3. Press `p` to preview a backup
4. Press `R` to restore (with confirmation)
5. Press `c` to cleanup old backups (>30 days)

---

### Technology Stack

- **Language:** Go 1.21+
- **TUI Framework:** [Bubbletea](https://github.com/charmbracelet/bubbletea) (Elm architecture)
- **Styling:** [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Components:** [Bubbles](https://github.com/charmbracelet/bubbles)
- **Config:** YAML (gopkg.in/yaml.v3)
- **APIs:** Cloudflare DNS API, Docker CLI

---


## Roadmap

### v1.0.0 (Current)
- Complete CRUD operations for DNS and Caddy entries
- Advanced filtering, sorting, and batch operations
- Multi-profile system with interactive setup wizard
- Backup management with 30-day retention and audit logging
- Full mouse and keyboard support with lazygit-style interface
- **Snippet system** for reusable Caddyfile configurations
  - Three-panel layout with dedicated snippets panel
  - Interactive snippet creation wizard
  - Smart form suggestions and brownfield support
  - Auto-detection and visual indicators

### v1.1.0 (Planned)
- **TXT record support:** Beyond A and CNAME records
- **Configurable keybindings:** User-defined shortcuts
- **Theme customization:** Custom color schemes
- **Export/import:** Backup entire configurations
- **Snippet library:** Community-contributed snippet templates

### v2.0 (Future)
- Support for other DNS providers (Route53, DigitalOcean)
- Support for other reverse proxies (Traefik, nginx)

---

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup
- Code structure overview
- Testing guidelines
- Pull request process

---

## Troubleshooting

### Common Issues

**Config validation errors:**
- Ensure `zone_id` is exactly 32 hex characters
- Ensure `domain` is a valid FQDN (e.g., example.com)
- Check CIDR format for `lan_subnet` (e.g., 10.0.0.0/24)

**Caddy validation failures:**
- Check Caddyfile syntax with `caddy validate`
- Ensure LazyProxyFlare has read/write access to Caddyfile
- Review backup files in case of corruption

**Docker restart failures:**
- Verify container name matches `docker ps`
- Ensure user has Docker socket access
- Check Docker daemon is running

**API errors:**
- Verify API token permissions (Zone.DNS Edit)
- Check zone_id matches your domain in Cloudflare
- Ensure network connectivity to Cloudflare API

---

## Credits

- Inspired by [lazygit](https://github.com/jesseduffield/lazygit) by Jesse Duffield
- Built with [Charm.sh](https://charm.sh/) ecosystem
- Powered by [Cloudflare API](https://developers.cloudflare.com/api/)
- Caddy server: [caddyserver.com](https://caddyserver.com/)

---

## License

GNUGPLv3 License - see LICENSE file for details

---

## Support

- Documentation: See [KEYBINDINGS.md](docs/KEYBINDINGS.md)
- Issues: [GitHub Issues](https://github.com/tinkermesomething/lazyproxyflare/issues)
- Discussions: [GitHub Discussions](https://github.com/tinkermesomething/lazyproxyflare/discussions)

---

**Made with ❤️ for the homelab and self-hosting community**
