# LazyProxyFlare Keybindings Reference

Complete keyboard shortcut reference organized by context. All keybindings are context-sensitive and displayed in the status bar based on the current view.

---

## Table of Contents

- [Global Keys](#global-keys)
- [Main List View](#main-list-view)
- [Forms (Add/Edit)](#forms-addedit)
- [Snippet Detail View](#snippet-detail-view)
- [Snippet Wizard](#snippet-wizard)
- [Backup Manager](#backup-manager)
- [Audit Log Viewer](#audit-log-viewer)
- [Confirmation Dialogs](#confirmation-dialogs)
- [Help Screen](#help-screen)
- [Mouse Controls](#mouse-controls)
- [Setup Wizard](#setup-wizard)
- [Profile Selector](#profile-selector)
- [Profile Editor](#profile-editor)

---

## Global Keys

These keys work in most contexts unless overridden by a specific view.

| Key | Action | Notes |
|-----|--------|-------|
| `ESC` | Cancel / Go back | Returns to previous view, clears filters/search |
| `Ctrl+W` | Close modal | Closes modal windows (profile editor, etc.) |
| `?` | Toggle help screen | Shows all keybindings and available actions |
| `q` | Quit application | Exit the application from main view (shows confirmation) |
| `Ctrl+C` | Force quit | Emergency exit (no confirmation) |

---

## Main List View (Dashboard)

The primary interface showing entries with sync status.

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `↓` / `j` | Move down | Navigate down through entries (j only in list view) |
| `↑` / `k` | Move up | Navigate up through entries (k only in list view) |
| `g` | Jump to top | Go to first entry |
| `G` | Jump to bottom | Go to last entry |
| `Home` | Jump to top | Alternative to `g` |
| `End` | Jump to bottom | Alternative to `G` |
| `Tab` | Switch views | Toggle between Cloudflare DNS and Caddy views |
| `Shift+Tab` | Previous panel | Reverse cycle between panels |

**Note:** In modal windows (wizards, forms, dialogs), use arrow keys for navigation. The `j` and `k` keys are reserved for text input in modals.

### Actions

| Key | Action | Description |
|-----|--------|-------------|
| `a` | Add new entry | Opens form to create DNS + Caddy entry |
| `Enter` | Edit entry | Edit the selected DNS entry |
| `d` | Delete entry | Delete the selected entry |
| `s` | Sync entry | Create missing DNS or Caddy for orphaned entries |
| `w` | Snippet wizard | Open snippet wizard to create reusable Caddy config blocks |
| `b` | Backup manager | View, restore, preview, and delete Caddyfile backups |
| `p` | Profile selector | Switch between profiles or create new ones |
| `r` | Refresh | Reload data from Cloudflare and Caddyfile |
| `Enter` | View details | Open detail view for selected entry (context-dependent) |

**Panel Focus:**
- The interface has multiple panels showing entries from both Cloudflare and Caddy
- Use `Tab` / `Shift+Tab` to cycle between panels
- Navigation keys (`j`, `k`, `g`, `G`) operate on the currently focused panel
- Current panel is indicated by a highlighted border

### Filtering & Sorting

| Key | Action | Filter Options |
|-----|--------|----------------|
| `f` | Cycle status filter | **All** → Synced → Orphaned DNS → Orphaned Caddy → All |
| `t` | Cycle DNS type filter | **All** → CNAME → A → All |
| `o` | Cycle sort mode | **Alphabetical** ↔ By Status |
| `/` | Search mode | Enter domain search (type to filter) |
| `ESC` | Clear all filters | Resets filters, sort, and search to defaults |

**Filter Combinations:**
- Filters stack: Can use status filter + DNS type filter + search simultaneously
- Example: `f` (Orphaned DNS) + `t` (CNAME) = Show only orphaned CNAME records
- `ESC` clears everything and returns to default view (All, All, Alphabetical)

### Multi-Select & Batch Operations

| Key | Action | Description |
|-----|--------|-------------|
| `Space` | Toggle selection | Add/remove current entry from selection |
| `X` | Batch delete selected | Delete all selected entries (with confirmation) |
| `S` | Batch sync selected | Sync all selected orphaned entries |
| `D` | Bulk delete menu | Choose: Delete all orphaned DNS **or** all orphaned Caddy |

**Selection Notes:**
- Selected entries show `[✓]` checkbox
- Selection count shown in info line: "X selected"
- Selections cleared after batch operations complete
- Can navigate while entries are selected

### Help

| Key | Action | Description |
|-----|--------|-------------|
| `?` | Toggle help screen | Shows all keybindings and available actions |
| `q` | Quit application | Exit the application |

---

## Forms (Add/Edit)

Interactive forms for creating or editing entries.

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `↓` | Next field | Move to next form field |
| `↑` | Previous field | Move to previous form field |
| `Tab` | Next field | Alternative to `↓` |
| `Shift+Tab` | Previous field | Alternative to `↑` |

### Input

| Key | Action | Context |
|-----|--------|---------|
| `Type characters` | Enter text | Text input fields (subdomain, target, port) |
| `Backspace` | Delete character | Text input fields |
| `Space` | Toggle checkbox | Checkbox fields (Proxied, SSL, LAN Only, etc.) |
| `Space` | Cycle option | DNS Type selector (CNAME ↔ A) |

### Actions

| Key | Action | Description |
|-----|--------|-------------|
| `Enter` | Next step | Preview form **or** confirm preview (context-dependent) |
| `Ctrl+M` | Insert newline | In Custom Caddy Config field only (Enter proceeds to preview) |
| `ESC` | Cancel | Close form without saving, return to main view |

**Form Fields:**

1. **Subdomain(s)** - Multi-line text input (supports multiple subdomains)
   - Single subdomain: `plex`
   - Multiple subdomains: Enter one per line (e.g., `mail` ↵ `webmail` ↵ `imap`)
   - Press `Enter` to add new line in subdomain field
   - Creates N DNS records + 1 Caddy block with all domains
2. **DNS Type** - Toggle (CNAME / A) - Use `Space` to toggle
3. **Target/IP** - Text input (domain for CNAME, IPv4 for A)
4. **DNS Only** - Checkbox - Skip Caddy configuration
5. **Proxied** - Checkbox - Cloudflare proxy (orange cloud)
6. **Reverse Proxy Target** - Text input (internal IP or hostname for Caddy)
7. **Service Port** - Number input (only if not DNS-only)
8. **SSL** - Checkbox - HTTPS upstream connection

**Multi-Subdomain Support:**
- Enter multiple subdomains (one per line) to create multi-domain entries
- Preview shows all FQDNs that will be created
- Single Caddyfile block with comma-separated domains: `mail.com, webmail.com, imap.com { ... }`
- Separate DNS record created for each subdomain
- Example: mail, webmail, imap → 3 DNS records + 1 Caddy block

**Smart Navigation:**
- DNS-only mode hides Caddy fields (Port, SSL, LAN, OAuth, WebSocket)
- Form validation prevents preview until all required fields are filled
- IPv4 validation for A records (0-255 per octet)

---

## Backup Manager

View and manage Caddyfile backups.

| Key | Action | Description |
|-----|--------|-------------|
| `↓` | Move down | Navigate to next backup |
| `↑` | Move up | Navigate to previous backup |
| `p` | Preview backup | View full contents of selected backup file |
| `R` | Restore backup | Restore selected backup (with confirmation + validation) |
| `x` | Delete backup | Delete selected backup file (with confirmation) |
| `c` | Cleanup old backups | Delete backups older than retention period (default: 30 days) |
| `ESC` | Close manager | Return to main view |

**Backup Information Displayed:**
- Filename with timestamp (e.g., `Caddyfile.backup.20231228_143022`)
- File size (e.g., "12.5 KB")
- Age (e.g., "2 hours ago", "3 days ago")

**Cleanup Preview:**
- Shows which backups will be deleted
- Displays total space to be freed
- Requires confirmation before deletion

---

## Audit Log Viewer

View history of all operations performed in LazyProxyFlare.

| Key | Action | Description |
|-----|--------|-------------|
| `↓` | Scroll down | View older log entries |
| `↑` | Scroll up | View newer log entries |
| `ESC` | Close log | Return to main view |

**Log Entry Format:**
```
2025-12-28 14:30:45  ✓ CREATE  plex.example.com
  Type: CNAME, Target: mail.example.com, Proxied
```

**Operations Logged:**
- **CREATE** - New entry added (DNS + Caddy)
- **UPDATE** - Existing entry modified
- **DELETE** - Entry removed (DNS, Caddy, or both)
- **SYNC** - Orphaned entry synced (created missing counterpart)
- **RESTORE** - Backup restored

**Result Indicators:**
- `✓` - Success
- `✗` - Failure (with error message)

**Details Shown:**
- DNS type (CNAME / A)
- Target or IP address
- Proxied status
- Sync direction (for sync operations)
- Batch count (for batch operations)
- Error messages (for failures)

---

## Confirmation Dialogs

All destructive operations require confirmation.

| Key | Action | Description |
|-----|--------|-------------|
| `y` / `Y` | Confirm | Proceed with the operation |
| `n` / `N` | Cancel | Abort operation, return to previous view |
| `ESC` | Cancel | Alternative to `n` |

**Confirmation Types:**
- **Delete** - Single entry or batch
- **Sync** - Create missing DNS or Caddy
- **Restore** - Restore Caddyfile from backup
- **Cleanup** - Delete old backup files
- **Bulk Delete** - Delete all orphaned entries

---

## Help Screen

Interactive help display with all keybindings.

| Key | Action | Description |
|-----|--------|-------------|
| `↓` | Scroll down | View more keybindings |
| `↑` | Scroll up | Scroll back up |
| `ESC` / `?` | Close help | Return to main view |

**Help Sections:**
1. Navigation
2. Actions
3. Filtering
4. Batch Operations
5. Tools

---

## Snippet Detail View

View and edit reusable Caddy configuration snippets.

### View Mode (Read-Only)

| Key | Action | Description |
|-----|--------|-------------|
| `e` | Edit snippet | Enter edit mode (editable textarea) |
| `ESC` | Close detail | Return to main view |

**Display Information:**
- Snippet name and category badge
- Description and auto-detection confidence (if applicable)
- Usage statistics (X entries using this snippet)
- List of entries importing this snippet
- Location in Caddyfile (line numbers)
- Full snippet content (syntax highlighted)

### Edit Mode

Press `e` in view mode to enter edit mode.

| Key | Action | Description |
|-----|--------|-------------|
| `y` / `Enter` | Save changes | Validate and write to Caddyfile |
| `d` | Delete snippet | Remove snippet (checks if in use first) |
| `ESC` | Cancel edit | Discard changes, return to view mode |
| `Type characters` | Edit content | Multi-line textarea with scrolling |
| `Backspace` | Delete character | Standard text editing |
| `Arrow keys` | Navigate cursor | Move within textarea |
| `Tab` | Insert tab | (Not for navigation - inserts literal tab) |

**Edit Mode Features:**
- 70x15 textarea with syntax highlighting
- Auto-backup before save
- Caddyfile validation before commit
- Automatic rollback on validation failure
- Caddy auto-reload on success
- Audit logging

**Deletion Safety:**
- Cannot delete snippet currently in use
- Shows error: "cannot delete snippet 'X': currently used by N entries"
- Must remove from entries first
- No confirmation dialog (direct action)

**Example Workflow:**
```
1. Navigate to snippet in snippets panel
2. Press Enter → Opens detail view
3. Press e → Enters edit mode
4. Edit content in textarea
5. Press y → Saves changes
   - Creates backup
   - Validates Caddyfile
   - Reloads Caddy
   - Returns to view mode
```

---

## Snippet Wizard

Interactive wizard for creating reusable Caddyfile configuration blocks.

### Wizard Modes

The wizard offers three creation modes:

| Mode | Description | Use Case |
|------|-------------|----------|
| **Templated - Advanced** | Choose from pre-configured templates | Quick setup for common patterns |
| **Custom - Paste Your Own** | Paste existing Caddy config | Convert existing config to snippet |
| **Guided - Step by Step** | Walk through each snippet type | Learn about available options |

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `↓` | Move down | Navigate options/fields |
| `↑` | Move up | Navigate options/fields |
| `Tab` | Next field | Forward navigation in multi-field screens |
| `Shift+Tab` | Previous field | Backward navigation |
| `Space` | Toggle selection | Select/deselect templates or options |
| `Enter` | Next step | Proceed to next wizard screen |
| `b` | Go back | Return to previous wizard step |
| `ESC` | Cancel wizard | Exit wizard, return to main view |

**Navigation Pattern:** Snippet wizard uses `Tab` for field navigation (like Setup wizard) and `Enter` to proceed between steps.

### Custom Snippet Mode

Special keybindings for "Custom - Paste Your Own" mode:

| Key | Action | Context |
|-----|--------|---------|
| `Type characters` | Enter text | Name field or content textarea |
| `Ctrl+V` | Paste content | Paste multi-line Caddy config |
| `Tab` | Next field | Name → Content |
| `Shift+Tab` | Previous field | Content → Name |
| `Enter` | Create snippet | Generate and save (when both fields filled) |
| `ESC` | Cancel | Return to main view |

**Custom Mode Features:**
- Name field: Single-line text input (50 char limit)
- Content field: Multi-line textarea (70x10, unlimited chars)
- Live preview showing snippet format: `(name) { content }`
- Tab inserts literal tab in content (not navigation)
- Arrow keys navigate within focused field

### Wizard Steps (Guided Mode)

1. **Welcome** - Introduction and mode selection
2. **Category Selection** - Choose which snippet categories to configure
3. **Template Selection** - (Templated mode) Choose from 14+ templates
4. **Parameter Configuration** - (Per template) Configure dynamic parameters
5. **IP Restriction** - Configure LAN subnet and allowed external IPs
6. **Security Headers** - Choose preset (basic/strict/paranoid)
7. **Performance** - Enable compression options
8. **Summary** - Review all snippets before creation

### Parameter Validation

Templates with configurable parameters include real-time validation:

| Template | Parameters | Validation |
|----------|-----------|------------|
| CORS Headers | Allowed origins, methods, headers | URL format, HTTP method names |
| Rate Limiting | Requests per window, window duration | Positive integers, duration format |
| Extended Timeouts | Read, write, dial timeouts | Duration format (e.g., "5m", "30s") |
| Large Uploads | Max body size | Size format (e.g., "100MB", "1GB") |

**Validation Errors:**
- Shown inline under invalid fields
- Red text with error message
- Cannot proceed to next step until fixed
- Format hints displayed in placeholder text

---

## Mouse Controls

Full mouse support for navigation and interaction.

### List View

| Action | Result |
|--------|--------|
| **Click on entry** | Select entry (moves cursor) |
| **Click on checkbox** | Toggle selection for batch operations |
| **Scroll wheel up/down** | Navigate through list |

### Forms

| Action | Result |
|--------|--------|
| **Click on field** | Focus that form field |
| **Click on checkbox** | Toggle checkbox value |
| **Scroll wheel** | Navigate form fields |

### Backup Manager

| Action | Result |
|--------|--------|
| **Click on backup** | Select that backup |
| **Scroll wheel** | Navigate backup list |

### Modals

| Action | Result |
|--------|--------|
| **Click outside modal** | (No effect - must use ESC or action keys) |
| **Scroll wheel** | Scroll modal content if scrollable |

---

## Context-Sensitive Status Bar

The status bar at the bottom of the screen changes based on context to show only relevant keys.

### Main View (Normal Mode)
```
↑/↓:nav  a:add  Enter:edit  d:delete  s:sync  w:snippets  b:backups  p:profiles  ?:help  q:quit
```

### Main View (Selection Mode)
```
Space:select  X:delete selected  S:sync selected  ESC:clear  ?:help  q:quit
```

### Main View (Search Mode)
```
Type to search...  ESC:cancel
```

### Forms
```
Tab/↓/↑:navigate  Space:toggle  Enter:preview  ESC:cancel
```

### Form Preview
```
y:confirm  n:cancel  ESC:back
```

### Backup Manager
```
p:preview  R:restore  x:delete  c:cleanup  ESC:close
```

### Audit Log
```
↓/↑:scroll  ESC:close
```

---

## Tips & Tricks

### Efficient Workflows

**1. Quick Filter Workflow:**
```
f → f → f    # Cycle to desired filter
Space Space Space    # Select multiple entries
X    # Batch delete
```

**2. Search + Filter:**
```
/    # Enter search mode
plex    # Type search term
f    # Add status filter while search is active
```

**3. Bulk Cleanup:**
```
D    # Open bulk delete menu
↓    # Select "Delete all orphaned DNS"
Enter    # Confirm
y    # Execute
```

**4. Emergency Recovery:**
```
b    # Open backup manager
↑    # Select most recent backup
R    # Restore
y    # Confirm
```

---

## Setup Wizard

Interactive 12-step wizard for first-time configuration and creating new profiles.

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `↓` | Next option | Navigate radio buttons (Proxy Type, Deployment, Encryption) |
| `↑` | Previous option | Navigate radio buttons |
| `Tab` | Next option | Alternative to `↓` for radio button navigation |
| `Shift+Tab` | Previous option | Alternative to `↑` for radio button navigation |
| `Enter` | Confirm / Next | Accept input and proceed to next step |
| `b` | Back | Return to previous wizard step |
| `ESC` | Cancel wizard | Exit wizard (returns to profile selector if profiles exist) |

**Navigation Pattern:** Setup wizard supports both arrow keys (`↑`/`↓`) and Tab keys for navigating selection options. All methods work consistently across all wizard steps.

### Input

| Key | Action | Context |
|-----|--------|---------|
| `Type characters` | Enter text | Text input fields (profile name, domain, API token, etc.) |
| `Backspace` | Delete character | Text input fields |

### Wizard Steps

1. **Welcome** - Introduction screen
2. **Profile Name** - Choose a name for this profile
3. **Domain** - Enter your domain name
4. **API Token** - Cloudflare API token
5. **Zone ID** - Cloudflare zone ID
6. **Proxy Type** - Select reverse proxy (Caddy/nginx/traefik)
7. **Deployment** - Select deployment method (Docker/systemd/binary)
8. **Caddyfile Path** - Path to Caddyfile
9. **Container Name** - Docker container name (if Docker deployment)
10. **Validation Command** - Optional custom validation command
11. **Defaults** - Default settings for new entries
12. **Summary** - Review and confirm

### Summary Screen

| Key | Action | Description |
|-----|--------|-------------|
| `y` | Save profile | Create profile and start using it |
| `n` | Cancel | Return to previous step |
| `b` | Go back | Edit previous settings |

**Input Validation:**
- Profile name: Alphanumeric, hyphen, underscore
- Domain: Valid domain characters
- API Token/Zone ID: Alphanumeric only
- Caddyfile Path: Valid path characters
- Container Name: Valid container name characters
- Validation Command: Any printable characters

---

## Profile Selector

Modal interface for switching between profiles, editing profiles, or creating new ones.

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `↓` | Next profile | Navigate down the profile list |
| `↑` | Previous profile | Navigate up the profile list |
| `Enter` | Select profile | Load selected profile and switch to it |
| `ESC` | Cancel | Close selector, return to main view |

### Actions

| Key | Action | Description |
|-----|--------|-------------|
| `e` | Edit profile | Edit the selected profile's settings |
| `+` | New profile | Launch setup wizard to create new profile |
| `n` | New profile | Same as `+` |

**Profile Selector Features:**
- Currently active profile highlighted with "(active)" label
- Last used profile remembered and pre-selected
- Shows "Add new profile (run wizard)" option at bottom
- Automatic data reload when switching profiles
- Can edit profile settings without leaving selector

**Example Display:**
```
Select Profile

  1. homelab (active)
> 2. production
  3. staging

  + Add new profile (run wizard)

j/k: navigate  Enter: select  e: edit  +/n: new  ESC: cancel
```

---

## Profile Editor

Modal interface for editing profile settings without creating a new profile.

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `Tab` | Next field | Move to next editable field |
| `Shift+Tab` | Previous field | Move to previous field |
| `↓` | Next field | Alternative to Tab for vertical navigation |
| `↑` | Previous field | Alternative to Shift+Tab for vertical navigation |

### Input

| Key | Action | Context |
|-----|--------|---------|
| `Type characters` | Enter text | Text input fields (profile name, API token, domain, paths, etc.) |
| `Backspace` | Delete character | Text input fields |
| `Space` | Toggle boolean | Boolean/checkbox fields (SSL, Proxied, etc.) |

### Fields

Profile editor allows editing:

1. **Profile Name** - Name of the profile (alphanumeric, hyphen, underscore)
2. **Domain** - Base domain for this profile
3. **API Token** - Cloudflare API token
4. **Zone ID** - Cloudflare zone ID
5. **Caddyfile Path** - Path to the Caddyfile
6. **Container Name** - Docker container name (if using Docker)
7. **Default CNAME Target** - Default target for new CNAME entries
8. **Default Port** - Default port for reverse proxy entries
9. **Default SSL** - Toggle SSL for upstream connections (on/off)
10. **Default Proxied** - Toggle Cloudflare proxy status (on/off)

### Actions

| Key | Action | Description |
|-----|--------|-------------|
| `Enter` | Save changes | Save edited profile and return to selector |
| `ESC` | Cancel | Discard changes and return to selector |

**Profile Editor Features:**
- Edit profile settings without recreating the entire profile
- All validation from setup wizard applies
- Changes take effect immediately when saved
- Currently active profile shows "(active)" label
- Cannot delete profile from editor (use profile management)

**Example Display:**
```
Edit Profile: homelab

Profile Name ..................... homelab
Domain ............................ example.com
API Token ......................... ••••••••••••
Zone ID ........................... ••••••••••••

Caddyfile Path .................... /etc/caddy/Caddyfile
Container Name .................... caddy

Default CNAME Target .............. www.example.com
Default Port ...................... 80
Default SSL ....................... [✓] On
Default Proxied ................... [✓] On

Tab/j/k: navigate  Space: toggle  Enter: save  ESC: cancel
```

---

### Keyboard Navigation Philosophy

LazyProxyFlare uses intuitive navigation patterns:
- Arrow keys (`↑`/`↓`) for all navigation (always works)
- `j/k` for vertical movement in main list view only (vim-style convenience)
- `g/G` for jump to top/bottom
- `ESC` to cancel or go back
- `Ctrl+W` to close modal windows
- `?` for help/keybindings
- `q` to quit (with confirmation)
- `Space` for selection/toggle
- `Enter` to edit selected entry
- Single character keys for common actions (a, d, s, w, b, p)

---

## Customization (Future)

In v1.1+, keybindings will be customizable via config:

```yaml
keybindings:
  add: "a"
  delete: "d"
  quit: "ctrl+q"
  # ... etc
```

---

## Wizard Navigation Patterns

Both wizards (Setup and Snippet) follow consistent navigation patterns:

### Common Keybindings

| Wizard | Option Navigation | Step Progression | Go Back | Cancel |
|--------|------------------|------------------|---------|--------|
| **Setup Wizard** | `Tab` / `Shift+Tab` / `↑/↓` | `Enter` | `b` | `ESC` |
| **Snippet Wizard** | `Tab` / `Shift+Tab` / `↑/↓` | `Enter` | `b` | `ESC` |

### Navigation Consistency

Both wizards support **two equivalent methods** for navigating options:
1. **Arrow keys** (`↑` / `↓`) - Standard terminal navigation
2. **Tab keys** (`Tab` / `Shift+Tab`) - Traditional form navigation

Both methods work identically in both wizards for maximum flexibility.

### Key Differences

| Feature | Setup Wizard | Snippet Wizard |
|---------|-------------|----------------|
| **Radio buttons** | Both navigation methods | Both navigation methods |
| **Checkboxes** | N/A | `Space` to toggle |
| **Multi-select** | N/A | `Space` for templates |
| **Multi-field steps** | Single field per step | Name + Content fields (custom mode) |

### Navigation Tips

1. **Consistent across wizards** - Both navigation methods work in both wizards
2. **Use whichever feels natural** - Arrow keys or Tab - your choice
3. **Use `b` to go back** - Review and edit previous steps without losing progress
4. **`ESC` cancels safely** - Returns to main view without saving
5. **Input validation** - Both wizards validate before allowing progression to next step

---

## Getting Help

- Press `?` any time to see context-appropriate keybindings
- Status bar always shows available actions for current context
- This document: Complete reference for all shortcuts

---

**Last Updated:** 2026-01-25
**Version:** 1.1 (with profile editing)
