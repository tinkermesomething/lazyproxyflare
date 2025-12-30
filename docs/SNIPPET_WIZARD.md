# Snippet Wizard Documentation

## Overview

The snippet wizard provides three modes for creating reusable Caddy configuration snippets:
- Templated Mode: Select from 14 pre-built templates with parameter configuration
- Custom Mode: Paste your own Caddy configuration
- Auto-Detection: Extract patterns from existing Caddyfile

## Available Templates

### Security (5 templates)

1. **CORS Headers** (Configurable)
   - Cross-Origin Resource Sharing configuration
   - Parameters:
     - Allowed Origins (default: *)
     - Allowed Methods (default: GET, POST, PUT, DELETE, OPTIONS)
     - Allow Credentials (default: false)
   - Generated snippet name: cors_headers

2. **Rate Limiting** (Configurable)
   - Zone-based request rate limiting
   - Parameters:
     - Requests per Second (default: 100)
     - Burst Size (default: 50)
   - Generated snippet name: rate_limiting

3. **Auth Headers**
   - Forwards authentication headers to upstream
   - Includes: X-Real-IP, X-Forwarded-For, X-Forwarded-Proto
   - Generated snippet name: auth_headers

4. **IP Restriction** (Configurable)
   - Limits access to specific IP ranges
   - Parameters:
     - LAN Subnet (default: 10.0.0.0/8)
     - Allowed External IP (optional, no default)
   - Generated snippet name: ip_restricted

5. **Security Headers** (Configurable)
   - Adds security response headers
   - Parameters:
     - Preset: basic, strict, paranoid (default: strict)
   - Generated snippet name: security_headers

### Performance (3 templates)

6. **Static Caching** (Configurable)
   - Cache-Control headers for static assets
   - Parameters:
     - Cache Max-Age in seconds (default: 86400 / 1 day)
     - Enable ETag (default: true)
   - File types: css, js, images, fonts
   - Generated snippet name: static_caching

7. **Advanced Compression** (Configurable)
   - Multi-encoder compression
   - Parameters:
     - Compression Level: 1-9 (default: 5)
     - Enable gzip (default: true)
     - Enable zstd (default: true)
     - Enable brotli (default: false)
   - Generated snippet name: compression_advanced

8. **Basic Performance**
   - Standard gzip and zstd compression
   - Generated snippet name: performance

### Backend Integration (3 templates)

9. **WebSocket Support**
   - Full WebSocket upgrade support
   - Infinite timeouts for persistent connections
   - Forwards Connection, Upgrade, X-Real-IP headers
   - Generated snippet name: websocket_advanced

10. **Extended Timeouts** (Configurable)
    - Custom transport timeout configuration
    - Parameters:
      - Read Timeout (default: 120s)
      - Write Timeout (default: 120s)
      - Dial Timeout (default: 30s)
    - Generated snippet name: extended_timeouts

11. **HTTPS Backend** (Configurable)
    - HTTPS upstream with TLS options
    - Parameters:
      - Skip TLS Verification (default: false)
      - Keepalive Connections (default: 100)
    - Generated snippet name: https_backend

### Content Control (3 templates)

12. **Large Uploads** (Configurable)
    - Request body size limits
    - Parameters:
      - Maximum Upload Size (default: 512MB)
      - Examples: 100MB, 1GB, 5GB
    - Generated snippet name: large_uploads

13. **Custom Headers**
    - Arbitrary header injection
    - Directions: upstream, response, or both
    - Generated snippet name: custom_headers_inject

14. **Frame Embedding**
    - CSP frame-ancestors configuration
    - Controls iframe embedding
    - Generated snippet name: frame_embedding

## Wizard Modes

### Templated Mode

Flow:
1. Welcome: Select "Templated - Advanced"
2. Auto-Detect: Review patterns found in existing Caddyfile
3. Template Selection: Choose from 14 templates
4. Parameter Configuration: Configure selected templates (if applicable)
5. Summary: Review all snippets to be created
6. Create: Write snippets to Caddyfile

Configurable templates will show parameter screens.
Non-configurable templates use smart defaults.

### Custom Mode

Flow:
1. Welcome: Select "Custom - Paste Your Own"
2. Custom Snippet: Enter snippet name and paste content
3. Summary: Review snippet
4. Create: Write snippet to Caddyfile

Note: Do NOT include the (name) { } wrapper in pasted content.
The wizard adds the wrapper automatically.

### Auto-Detection

The wizard automatically scans your existing Caddyfile for patterns:
- Large upload configurations (request_body max_size)
- Extended timeout configurations (transport timeouts)
- OAuth/auth header forwarding patterns
- Streaming support (flush_interval)
- Custom keepalive configurations

Detected patterns appear in the Auto-Detect step with checkboxes.
Select patterns to extract into reusable snippets.

## Parameter Configuration

### Text Input Fields
- Type new value and press Tab to move to next field
- Leave empty to use default value
- Examples shown as hints

### Checkbox Fields
- Press Space to toggle
- Tab to move to next field

### Navigation
- Tab: Next field
- Shift+Tab: Previous field (not implemented)
- Up/Down: Navigate fields
- Space: Toggle checkboxes
- Enter: Continue to next step
- ESC: Go back to previous step

## Default Values

All configurable templates have sensible defaults:

CORS Headers:
- Origins: * (all origins)
- Methods: GET, POST, PUT, DELETE, OPTIONS
- Credentials: false

Rate Limiting:
- Requests/sec: 100
- Burst: 50

Large Uploads:
- Max size: 512MB

Extended Timeouts:
- Read: 120s
- Write: 120s
- Dial: 30s

Static Caching:
- Max-age: 86400 seconds (1 day)
- ETag: true

IP Restriction:
- LAN Subnet: 10.0.0.0/8
- External IP: (optional, no default)

Security Headers:
- Preset: strict

Compression Advanced:
- Level: 5
- Gzip: true
- Zstd: true
- Brotli: false

HTTPS Backend:
- Skip TLS Verify: false
- Keepalive: 100

## Snippet Storage

Snippets are prepended to your Caddyfile with this format:

```
# === Snippets created by LazyProxyFlare ===
# Generated: 2025-12-29 12:34:56

(snippet_name) {
    snippet content here
}

# === End of snippets ===

[rest of Caddyfile]
```

## Using Snippets

After creation, import snippets in your domain blocks:

```
example.com {
    import cors_headers
    import rate_limiting
    reverse_proxy localhost:8080
}
```

## Backup and Safety

- Automatic backup created before modification
- Caddy validation runs before applying changes
- Rollback on validation failure
- Duplicate detection prevents overwriting

## Error Handling

The wizard will abort and show errors for:
- Duplicate snippet names
- Invalid Caddyfile syntax
- Failed Caddy validation
- File write errors

In all cases, your original Caddyfile is preserved via backup.
