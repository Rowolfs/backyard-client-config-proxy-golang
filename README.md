# Backyard Client Config Proxy

A lightweight Go reverse proxy for intercepting and modifying client subscription responses before they are returned to Happ.

The proxy sits between Caddy and the Remnawave subscription service.

```text
Happ
  ↓
Caddy
  ↓
Backyard Client Config Proxy
  ↓
Remnawave Subscription Service
  ↓
Backyard Client Config Proxy
  ↓
Caddy
  ↓
Happ
```

