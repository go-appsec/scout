# go-harden/scout

[![license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/go-harden/scout/blob/main/LICENSE)
[![Build Status](https://github.com/go-harden/scout/actions/workflows/tests-main.yml/badge.svg)](https://github.com/go-harden/scout/actions/workflows/tests-main.yml)

A lightweight Go library for passive reconnaissance of domains, discovering subdomains and URLs by querying public APIs. Scout provides a minimal, dependency-light approach to target enumeration for security testing.

## Features

- Ergonomic iterator-based API
- Concurrent source querying with configurable parallelism
- Automatic deduplication across all sources
- Context-aware with timeout support
- Minimal dependencies

## Supported Result Types

Scout can discover:

| Type | Description |
|------|-------------|
| Subdomains | Subdomains of the target domain (e.g., `api.example.com`) |
| URLs | Full URLs under the target domain (e.g., `https://example.com/path`) |

## Quick Start

```bash
go get github.com/go-harden/scout@latest
```

```go
package main

import (
    "context"
    "fmt"

    "github.com/go-harden/scout"
)

func main() {
    ctx := context.Background()

    // Query all subdomain sources for a domain
    for sub, err := range scout.Subdomains(ctx, "example.com") {
        if err != nil {
            // Ignore or handle errors, usually rate limits
            continue
        }
        fmt.Println(sub)
    }
}
```

## Usage Examples

### Query All Sources

```go
ctx := context.Background()

// Get both subdomains and URLs from all sources
for result, err := range scout.Query(ctx, "example.com") {
    if err != nil {
        // Ignore or handle errors, usually rate limits
        continue
    }
    fmt.Printf("[%s] %s: %s\n", result.Source, result.Type, result.Value)
}
```

### Query URLs Only

```go
// Get only URLs from URL-yielding sources
for url, err := range scout.URLs(ctx, "example.com") {
    if err == nil {
        fmt.Println(url)
    }
}
```

### Parallelism and Timeouts

```go
// Configure parallelism and timeout
for sub, err := range scout.Subdomains(ctx, "example.com",
    scout.WithParallelism(4),           // 4 sources at once
    scout.WithTimeout(30*time.Second),  // 30s per source
) {
    if err == nil {
        fmt.Println(sub)
    }
}
```

### Rate Limiting

```go
// Apply global and per-source rate limits
for sub, err := range scout.Subdomains(ctx, "example.com",
    scout.WithGlobalRateLimit(10),              // 10 req/sec globally
    scout.WithSourceRateLimit("commoncrawl", 0.25), // 15 req/min for commoncrawl
) {
    if err == nil {
        fmt.Println(sub)
    }
}
```

### API Keys for Enhanced Limits

```go
// Some sources support optional API keys for higher rate limits
for sub, err := range scout.Subdomains(ctx, "example.com",
    scout.WithAPIKey("virustotal", "your-api-key"),
    scout.WithAPIKey("shodan", "your-api-key"),
) {
    // Process results...
}
```

## API Reference

### Functions

| Function | Description |
|----------|-------------|
| `Query(ctx, domain, ...opts)` | Query sources and yield all results (subdomains and URLs) |
| `Subdomains(ctx, domain, ...opts)` | Query sources and yield only subdomains |
| `URLs(ctx, domain, ...opts)` | Query sources and yield only URLs |

### Options

| Option | Description |
|--------|-------------|
| `WithSources([]Source)` | Specify which sources to query |
| `WithParallelism(n)` | Set concurrent source count (default: NumCPU×2) |
| `WithTimeout(duration)` | Set per-source timeout (default: 30s) |
| `WithGlobalRateLimit(rps)` | Set global rate limit (requests/second) |
| `WithSourceRateLimit(name, rps)` | Set per-source rate limit |
| `WithHTTPClient(client)` | Use custom HTTP client |
| `WithAPIKey(source, key)` | Set API key for a source |

### Source Registry

| Function | Description |
|----------|-------------|
| `sources.All()` | Get all registered sources |
| `sources.ByName(name)` | Get source by name |
| `sources.ByNames(...names)` | Get multiple sources by name |
| `sources.ByType(type)` | Get sources yielding specific result type |
| `sources.Names()` | Get names of all registered sources |

### Types

```go
// Result represents a discovery from a source
type Result struct {
    Type   ResultType // Subdomain or URL
    Value  string     // The discovered value
    Source string     // Which source found it
}

// ResultType indicates the kind of result
type ResultType uint8

const (
    Subdomain ResultType = 1 << iota // Subdomain result
    URL                              // URL result
)
```

## Available Sources

### No API Key Required (11 sources)

| Source | Yields | Description |
|--------|--------|-------------|
| `anubis` | Subdomain | Anubis subdomain database |
| `crtsh` | Subdomain | Certificate transparency logs |
| `commoncrawl` | Subdomain, URL | Common Crawl web archive |
| `digitorus` | Subdomain | Certificate details database |
| `hudsonrock` | Subdomain, URL | Data breach information |
| `rapiddns` | Subdomain | DNS record aggregator |
| `sitedossier` | Subdomain | Domain analysis tool |
| `thc` | Subdomain | THC subdomain lookup API |
| `alienvault` | URL | AlienVault OTX URL list |
| `hackertarget` | Subdomain | Host search (limited without key) |
| `reconeer` | Subdomain | Subdomain enumeration (limited without key) |

