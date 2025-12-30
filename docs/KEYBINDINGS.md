# LazyProxyFlare Keybindings Reference

Complete keyboard shortcut reference organized by context. All keybindings are context-sensitive and displayed in the status bar based on the current view.

---

## Table of Contents

- [Global Keys](#global-keys)
- [Main List View](#main-list-view)
- [Forms (Add/Edit)](#forms-addedit)
- [Backup Manager](#backup-manager)
- [Audit Log Viewer](#audit-log-viewer)
- [Confirmation Dialogs](#confirmation-dialogs)
- [Help Screen](#help-screen)
- [Mouse Controls](#mouse-controls)
- [Setup Wizard](#setup-wizard)
- [Profile Selector](#profile-selector)

---

## Global Keys

These keys work in most contexts unless overridden by a specific view.

| Key | Action | Notes |
|-----|--------|-------|
| `ESC` | Cancel / Go back | Returns to main view, clears filters/search |
| `q` | Quit application | Confirms before exit if in main view |
| `?` | Toggle help screen | Shows all keybindings |
| `Ctrl+C` | Force quit | Emergency exit (no confirmation) |

---

## Main List View

The primary interface showing entries with sync status.

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Move down | Navigate to next entry |
| `k` / `↑` | Move up | Navigate to previous entry |
| `g` | Jump to top | Go to first entry |
| `G` | Jump to bottom | Go to last entry |
| `Home` | Jump to top | Alternative to `g` |
| `End` | Jump to bottom | Alternative to `G` |

### Actions

| Key | Action | Description |
|-----|--------|-------------|
| `a` | Add new entry | Opens form to create DNS + Caddy entry |
| `e` | Edit selected entry | Modify existing entry (DNS and/or Caddy) |
| `d` | Delete entry | Delete from DNS, Caddy, or both (with confirmation) |
| `s` | Sync entry | Create missing DNS or Caddy for orphaned entries |
| `r` | Refresh | Reload data from Cloudflare and Caddyfile |
| `Enter` | View details | (Currently shown in right panel automatically) |

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

### Tools & Utilities

| Key | Action | Description |
|-----|--------|-------------|
| `b` | Backup manager | View, restore, preview, delete Caddyfile backups |
| `l` | Audit log | View history of all operations (create, update, delete, sync) |
| `p` / `Ctrl+P` | Profile selector | Switch between profiles or create new ones |

---

## Forms (Add/Edit)

Interactive forms for creating or editing entries.

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Next field | Move to next form field |
| `k` / `↑` | Previous field | Move to previous form field |
| `Tab` | Next field | Alternative to `j`/`↓` |
| `Shift+Tab` | Previous field | Alternative to `k`/`↑` |

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
| `ESC` | Cancel | Close form without saving, return to main view |

**Form Fields:**

1. **Subdomain** - Text input (e.g., "plex")
2. **DNS Type** - Toggle (CNAME / A) - Use `Space` to toggle
3. **Target/IP** - Text input (domain for CNAME, IPv4 for A)
4. **DNS Only** - Checkbox - Skip Caddy configuration
5. **Port** - Number input (only if not DNS-only)
6. **Proxied** - Checkbox - Cloudflare proxy (orange cloud)
7. **SSL** - Checkbox - HTTPS upstream connection
8. **LAN Only** - Checkbox - Restrict to LAN subnet
9. **OAuth** - Checkbox - Include OAuth headers
10. **WebSocket** - Checkbox - WebSocket support headers

**Smart Navigation:**
- DNS-only mode hides Caddy fields (Port, SSL, LAN, OAuth, WebSocket)
- Form validation prevents preview until all required fields are filled
- IPv4 validation for A records (0-255 per octet)

---

## Backup Manager

View and manage Caddyfile backups.

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Move down | Navigate to next backup |
| `k` / `↑` | Move up | Navigate to previous backup |
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
| `j` / `↓` | Scroll down | View older log entries |
| `k` / `↑` | Scroll up | View newer log entries |
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
| `j` / `↓` | Scroll down | View more keybindings |
| `k` / `↑` | Scroll up | Scroll back up |
| `ESC` / `?` | Close help | Return to main view |

**Help Sections:**
1. Navigation
2. Actions
3. Filtering
4. Batch Operations
5. Tools

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
a:add  e:edit  d:delete  s:sync  f:filter  t:type  o:sort  b:backups  l:log  ?:help  q:quit
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
Tab/j/k:navigate  Space:toggle  Enter:preview  ESC:cancel
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
j/k:scroll  ESC:close
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
| `j` / `↓` | Next option | Navigate radio buttons (Proxy Type, Deployment) |
| `k` / `↑` | Previous option | Navigate radio buttons |
| `Enter` | Confirm / Next | Accept input and proceed to next step |
| `b` | Back | Return to previous wizard step |
| `ESC` | Cancel wizard | Exit wizard (returns to profile selector if profiles exist) |

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

Modal interface for switching between profiles or creating new ones.

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Next profile | Navigate down the profile list |
| `k` / `↑` | Previous profile | Navigate up the profile list |
| `Enter` | Select profile | Load selected profile and switch to it |
| `ESC` | Cancel | Close selector, return to main view |
| `p` / `Ctrl+P` | Close selector | Same as ESC |

### Actions

| Key | Action | Description |
|-----|--------|-------------|
| `+` | New profile | Launch setup wizard to create new profile |
| `n` | New profile | Same as `+` |

**Profile Selector Features:**
- Currently active profile highlighted with "(active)" label
- Last used profile remembered and pre-selected
- Shows "Add new profile (run wizard)" option at bottom
- Automatic data reload when switching profiles

**Example Display:**
```
Select Profile

  1. homelab (active)
> 2. production
  3. staging

  + Add new profile (run wizard)

j/k: navigate  Enter: select  +/n: new profile  ESC: cancel
```

---

### Keyboard Navigation Philosophy

LazyProxyFlare follows Vi/Vim-style navigation patterns:
- `j/k` for vertical movement (down/up)
- `g/G` for jump to top/bottom
- `ESC` to cancel or go back
- `/` for search
- `Space` for selection/toggle

If you're familiar with lazygit, you'll feel right at home!

---

## Customization (Future)

In v1.1+, keybindings will be customizable via config:

```yaml
keybindings:
  add: "a"
  delete: "d"
  quit: "q"
  # ... etc
```

---

## Getting Help

- Press `?` any time to see context-appropriate keybindings
- Status bar always shows available actions for current context
- This document: Complete reference for all shortcuts

---

**Last Updated:** 2025-12-28
**Version:** 1.0
