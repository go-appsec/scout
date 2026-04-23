## Project Overview

Scout is a passive reconnaissance Go library that queries external sources to discover subdomains and URLs for a given domain. It uses Go 1.23+ iterators for lazy result streaming and runs sources concurrently with configurable parallelism and rate limiting.

## Build & Test Commands

```bash
make test          # Run tests (short mode)
make test-all      # Run tests with API calls, race detection and coverage
make lint          # Run golangci-lint and go vet
```

## Architecture

### Core Components

**scout.go** - Main entry point with three public functions:

- `Query(ctx, domain, opts...)` - Primary function that returns `iter.Seq2[sources.Result, error]`. Runs all configured sources concurrently with:
  - Semaphore-controlled parallelism (default: `runtime.NumCPU() * 2`)
  - Case-insensitive deduplication via `sync.Map` across all sources
  - Per-source timeout contexts (default: 30s)
  - Transport wrapping for User-Agent and rate limiting
  - Optional API key support for enhanced limits
  - Proper cleanup via `sync.WaitGroup` and context cancellation

- `Subdomains(ctx, domain, opts...)` - Convenience wrapper that auto-filters for subdomain sources and returns `iter.Seq2[string, error]`

- `URLs(ctx, domain, opts...)` - Convenience wrapper that auto-filters for URL sources and returns `iter.Seq2[string, error]`

Internal transport wrappers:
- `userAgentTransport` - Sets `Mozilla/5.0 (compatible; go-appsec/scout-v{Version})` header
- `rateLimitTransport` - Applies rate limiting using `golang.org/x/time/rate`

**options.go** - Functional options pattern with:

- `WithSources([]Source)` - Specify which sources to query (defaults to all registered)
- `WithHTTPClient(*http.Client)` - Custom HTTP client (creates default if nil)
- `WithParallelism(int)` - Concurrent source count
- `WithGlobalRateLimit(rate.Limit)` - Rate limit applied to all sources
- `WithSourceRateLimit(name, rate.Limit)` - Per-source rate limits
- `WithTimeout(duration)` - Per-source timeout
- `WithAPIKey(name, key)` - Optional API keys for sources

**util.go** - Helper utilities:

- `Collect[T any](iter.Seq2[T, error])` - Collects all results into a slice, joins errors

**sources/** - Source abstraction layer (52 files):

- `source.go` - Core abstractions:
  - `Source` struct with `Name`, `Yields` (ResultType flags), and `Run` function
  - `Result` struct with `Type`, `Value`, and `Source` fields
  - `ResultType` bit flags: `Subdomain = 1<<0`, `URL = 1<<1`
  - Global registry with thread-safe access (`sync.RWMutex`)
  - Registry functions: `Register`, `All`, `ByName`, `ByNames`, `ByType`, `Names`

- `extract.go` - Extraction utilities:
  - `SubdomainExtractor` - Regex-based subdomain extraction (case-insensitive, returns lowercase)
  - `URLExtractor` - Regex-based URL extraction with path support

- **11 implemented sources**: anubis, alienvault, alienvaultpassivedns, commoncrawl, crtsh, digitorus, hackertarget, hudsonrock, rapiddns, reconeer, sitedossier, thc
  - Mix of JSON APIs, HTML scraping, and text-based responses
  - Complex sources handle pagination (rapiddns, digitorus, commoncrawl, hudsonrock)
  - API key support where applicable (hackertarget)

- **39 stub sources** (TODO implementations with API details documented)

### Key Patterns

- **Iterator-based streaming**: Results yield as they arrive via `iter.Seq2[T, error]`, allowing early termination
- **Concurrent execution**: Each source runs in a goroutine, results multiplexed through a single channel
- **Deduplication**: Case-insensitive, space-trimmed, atomic via `sync.Map.LoadOrStore`
- **Rate limiting**: Stacked transport wrappers for global and per-source limits
- **Error handling**: Errors yielded individually through iterator without breaking iteration
- **Source registration**: `init()` functions call `Register()` to populate global registry
- **Early return**: Sources check `if !yield(...) { return }` for cancellation support

### Source Implementation Patterns

When implementing a new source:

1. **Registration**: Add `init()` function that calls `sources.Register()` with the source definition
2. **Source struct**: Define with `Name`, `Yields` (ResultType), and `Run` function
3. **Run function**: Private `run*` function that implements the iterator logic
4. **HTTP requests**: Use the provided `client` parameter, not a custom client
5. **Error handling**: Yield individual errors via `yield(Result{}, err)`, don't stop iteration
6. **Cancellation**: Check context via `ctx.Err()` in loops, respect `!yield(...)` early return
7. **Extraction**: Use `SubdomainExtractor` or `URLExtractor` for parsing responses when applicable

Common source patterns:
- **Simple JSON API**: Single request, parse JSON, extract results (anubis, reconeer)
- **Paginated API**: Loop fetching pages until no more results (rapiddns, digitorus, hudsonrock)
- **Multi-step**: Fetch index/list, then query each item (commoncrawl)
- **HTML scraping**: Regex extraction from HTML responses (sitedossier)
- **Text-based**: Line-by-line parsing (thc, hackertarget)

API key handling:
- Check `apiKey` parameter, use enhanced endpoint if provided
- Fallback to public endpoint if no key
- Document key requirements in source comments

### Code Style

- Use `var` style for zero-value initialization: `var foo bool` not `foo := false`
- Comments should be concise simple and short phrases rather than full sentences when possible
- Comments should only be added when they describe non-obvious context (skip comments when the code or line is very obvious)
- Godocs should only describe the inputs and outputs, not how the function works
- Follow existing naming conventions and neighboring code style

### Testing

Structure and conventions:
- One `_test.go` file per implementation file that requires testing
- One `func Test<FunctionName>` per target function, using table-driven tests for consistent validation or `t.Run` test cases when assertions vary
- Test case names should be at most 3 to 5 words and in lower case with underscores
- Use `t.Parallel()` at test function start when no shared state, but not in the test cases
- Isolated temp directories via `t.TempDir()` when needed
- Context timeouts via `t.Context()` for tests with I/O

Assertions and validation:
- Assertions rely on `testify` (`require` for setup, `assert` for assertions)
- Don't include messages unless the message provides context outside of the test point or the two variables being evaluated
- Do NOT use time.Sleep for tests, instead use require.Eventually or deterministic triggers

Test helpers:
- `mockSource()` creates test sources with configurable results for unit testing
- Sources can return multiple results and errors to simulate real behavior

Verification:
- Always verify with `make test-all` and `make lint` before considering changes complete
