# Server Configuration

In addition to the usual settings, Glance now supports a background refresh
system that periodically updates widget data even when no client is
connected. This keeps dashboards fast and fresh.

Example:
```yaml
server:
  host: 0.0.0.0
  port: 8080
  assets-path: /home/user/glance-assets
  background-refresh-enabled: true
  background-refresh-interval: 15m
```

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `background-refresh-enabled` | boolean | `false` | Enable the background widget refresh goroutine. |
| `background-refresh-interval` | duration | `15m` | How often to check and update outdated widgets. |

> **Tip:** shorter intervals keep data fresher but increase CPU/IO usage.
