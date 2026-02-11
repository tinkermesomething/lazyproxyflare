# LazyProxyFlare

> A terminal UI for managing Cloudflare DNS records and Caddy reverse proxy configurations in sync — inspired by [lazygit](https://github.com/jesseduffield/lazygit).

> [!WARNING]
> This tool **directly modifies** your Cloudflare DNS records and Caddyfile. Mistakes can take your services offline. **Always back up before making changes** — use the built-in backup manager (`b`) and profile export (`x`). The author is not responsible for any data loss, DNS misconfiguration, or downtime caused by using this tool.

---

## Why?

If you self-host behind Caddy and use Cloudflare for DNS, you know the pain: add a service, create the DNS record, edit the Caddyfile, validate, restart the container — and hope you didn't fat-finger something. LazyProxyFlare keeps both in sync from one keyboard-driven interface.

---

## Features

- **DNS + Caddy in sync** — create, edit, delete entries that update both Cloudflare DNS and your Caddyfile atomically with automatic rollback on failure
- **CNAME and A records** — including DNS-only mode (no Caddy block)
- **Orphan detection** — visual indicators for entries that exist in DNS but not Caddy (or vice versa), with one-key sync
- **Multi-profile** — manage multiple domains/environments with separate profiles, export/import as `.tar.gz`
- **Setup wizard** — interactive first-run configuration, no manual YAML required
- **Batch operations** — multi-select entries for bulk delete or sync
- **Snippet system** — reusable Caddy config blocks (IP restrictions, security headers, compression) with an interactive wizard (`w`) and smart form suggestions
- **Backup manager** — automatic Caddyfile backups before every change, with restore, cleanup, and configurable rotation limits
- **Audit log** — full operation history with filtering by type, result, and domain search
- **Editor integration** — open your Caddyfile in `$EDITOR` directly from the UI (`E`)
- **Mouse + keyboard** — full mouse support alongside vim-style and arrow key navigation
- **Safety first** — pre-flight Caddy validation, confirmation dialogs on destructive ops, input format checking

---

## Installation

### Prerequisites

- Go 1.22+
- Cloudflare API token with DNS edit permissions
- Caddy server (Docker, docker-compose, or systemd)

### Build from Source

```bash
git clone https://github.com/tinkermesomething/lazyproxyflare.git
cd lazyproxyflare
make && sudo make install
```

### Quick Start

Just run it — the setup wizard handles the rest:

```bash
lazyproxyflare
```

The wizard walks you through profile name, domain, Cloudflare credentials, and Caddy deployment settings. For multiple profiles, press `p` then `+` to create more.

To skip the wizard, create a YAML file manually in `~/.config/lazyproxyflare/profiles/` — see [`config.example.yaml`](config.example.yaml) for the full reference.

### CLI Flags

```bash
lazyproxyflare              # Launch TUI
lazyproxyflare --profile X  # Load a specific profile by name
lazyproxyflare --version    # Show version
lazyproxyflare --help       # Show usage
```

---

## Keybindings

Press `?` in the app for the full 5-page help screen. The essentials:

| Key | Action |
|-----|--------|
| `a` | Add new entry |
| `Enter` | Edit selected entry |
| `d` | Delete selected entry |
| `s` | Sync orphaned entry |
| `Space` | Toggle selection |
| `X` / `S` / `D` | Batch delete / sync / bulk menu |
| `Tab` | Switch DNS ↔ Caddy tab |
| `f` / `t` / `o` | Filter by status / DNS type / sort |
| `/` | Search by domain |
| `p` | Profile selector |
| `b` | Backup manager |
| `w` | Snippet wizard |
| `l` | Audit log |
| `E` | Open Caddyfile in editor |
| `r` | Refresh data |
| `q` | Quit |

Full reference: [KEYBINDINGS.md](docs/KEYBINDINGS.md)

---

## Configuration

Profiles live in `~/.config/lazyproxyflare/profiles/` as YAML files. See [`config.example.yaml`](config.example.yaml) for all available options.

**Startup behavior:**
- No profiles → wizard launches automatically
- One profile → auto-loads
- Multiple profiles → selector appears

**Profile management** (from the profile selector, `p`):
- `e` edit, `d` delete, `x` export, `i` import, `n`/`+` create new

---

## Troubleshooting

**Config validation errors:**
- `zone_id` must be exactly 32 hex characters
- `domain` must be a valid FQDN (e.g., `example.com`)
- `lan_subnet` must be valid CIDR (e.g., `10.0.0.0/24`)

**Caddy validation failures:**
- Check syntax with `caddy validate`
- Ensure LazyProxyFlare has read/write access to the Caddyfile
- Check backups if the file got corrupted

**Docker restart failures:**
- Verify container name matches `docker ps` output
- Ensure your user has Docker socket access

**API errors:**
- Verify API token has `Zone.DNS Edit` permission
- Check `zone_id` matches your domain in Cloudflare dashboard

---

## Roadmap

### v1.3 (Planned)
- OS keyring integration for API token storage
- Configurable keybindings and theme customization
- TXT record support
- Community snippet library

### v2.0 (Future)
- Additional DNS providers (Route53, DigitalOcean)
- Additional reverse proxies (nginx, Traefik)

---

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for development setup, code structure, and testing guidelines.

---

## Credits

- Inspired by [lazygit](https://github.com/jesseduffield/lazygit) by Jesse Duffield
- Built with the [Charm.sh](https://charm.sh/) ecosystem ([Bubbletea](https://github.com/charmbracelet/bubbletea), [Lipgloss](https://github.com/charmbracelet/lipgloss), [Bubbles](https://github.com/charmbracelet/bubbles))
- Powered by [Cloudflare API](https://developers.cloudflare.com/api/) and [Caddy](https://caddyserver.com/)
- Developed with [Claude Code](https://claude.ai/claude-code) by Anthropic

## License

GPLv3 — see [LICENSE](LICENSE) for details.

## Support

- [GitHub Issues](https://github.com/tinkermesomething/lazyproxyflare/issues)
- [GitHub Discussions](https://github.com/tinkermesomething/lazyproxyflare/discussions)
- [Full keybinding reference](docs/KEYBINDINGS.md) · [Technical manual](TECHNICAL_MANUAL.md)
