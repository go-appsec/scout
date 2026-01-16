package scout

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"

	"github.com/go-harden/scout/sources"
)

// Query runs the specified sources against a domain and yields results.
// Results are deduplicated across all sources.
func Query(ctx context.Context, domain string, srcs []sources.Source, opts ...Option) iter.Seq2[sources.Result, error] {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	return func(yield func(sources.Result, error) bool) {
		// Cancel context when iterator returns to prevent goroutine leaks
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		client := cfg.HTTPClient
		if client == nil {
			client = &http.Client{
				Timeout:   cfg.Timeout,
				Transport: http.DefaultTransport,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return errors.New("redirects not allowed")
				},
			}
		}

		if cfg.UserAgent != "" {
			client = wrapClientWithUserAgent(client, cfg.UserAgent)
		}

		if cfg.GlobalRateLimit > 0 {
			client = wrapClientWithRateLimiter(client, rate.NewLimiter(cfg.GlobalRateLimit, 1))
		}

		dedupe := &deduplicator{}

		// Results channel
		type resultItem struct {
			result sources.Result
			err    error
		}
		results := make(chan resultItem)

		// Semaphore for parallelism control
		sem := make(chan struct{}, cfg.Parallelism)

		// Start source goroutines
		var wg sync.WaitGroup
		for _, src := range srcs {
			wg.Add(1)
			go func(s sources.Source) {
				defer wg.Done()

				// Acquire semaphore slot
				sem <- struct{}{}
				defer func() { <-sem }()

				// Per-source timeout context
				srcCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
				defer cancel()

				// Apply per-source rate limiting
				srcClient := client
				if limit, ok := cfg.SourceRateLimits[s.Name]; ok {
					srcClient = wrapClientWithRateLimiter(srcClient, rate.NewLimiter(limit, 1))
				}

				// Get API key for this source (if configured)
				var apiKey string
				if cfg.APIKeys != nil {
					apiKey = cfg.APIKeys[s.Name]
				}

				for result, err := range s.Run(srcCtx, domain, srcClient, apiKey) {
					select {
					case <-ctx.Done():
						return
					case results <- resultItem{result: result, err: err}:
					}
				}
			}(src)
		}

		// Close results when all sources complete
		go func() {
			wg.Wait()
			close(results)
		}()

		// Yield results with deduplication
		for r := range results {
			if r.err != nil {
				if !yield(sources.Result{}, r.err) {
					return
				}
				continue
			}

			if dedupe.seen(r.result.Value) {
				continue // skip duplicates
			}

			if !yield(r.result, nil) {
				return
			}
		}
	}
}

// Subdomains is a convenience wrapper that filters for Subdomain results only.
func Subdomains(ctx context.Context, domain string, srcs []sources.Source, opts ...Option) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for result, err := range Query(ctx, domain, srcs, opts...) {
			if err != nil {
				if !yield("", err) {
					return
				}
				continue
			}
			if result.Type == sources.Subdomain {
				if !yield(result.Value, nil) {
					return
				}
			}
		}
	}
}

// URLs is a convenience wrapper that filters for URL results only.
func URLs(ctx context.Context, domain string, srcs []sources.Source, opts ...Option) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for result, err := range Query(ctx, domain, srcs, opts...) {
			if err != nil {
				if !yield("", err) {
					return
				}
				continue
			}
			if result.Type == sources.URL {
				if !yield(result.Value, nil) {
					return
				}
			}
		}
	}
}

// deduplicator tracks seen values and filters duplicates.
type deduplicator struct {
	values sync.Map
}

// seen returns true if the value was already seen, and marks it as seen.
func (d *deduplicator) seen(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	_, loaded := d.values.LoadOrStore(normalized, struct{}{})
	return loaded
}

// userAgentTransport wraps an http.RoundTripper to set User-Agent header.
type userAgentTransport struct {
	base      http.RoundTripper
	userAgent string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", t.userAgent)
	return t.base.RoundTrip(req)
}

// rateLimitTransport wraps an http.RoundTripper to apply rate limiting.
type rateLimitTransport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

func (t *rateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.limiter.Wait(req.Context()); err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req)
}

// wrapClientWithRateLimiter returns a new client that applies rate limiting to all requests.
func wrapClientWithRateLimiter(client *http.Client, limiter *rate.Limiter) *http.Client {
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &http.Client{
		Transport:     &rateLimitTransport{base: transport, limiter: limiter},
		CheckRedirect: client.CheckRedirect,
		Jar:           client.Jar,
		Timeout:       client.Timeout,
	}
}

// wrapClientWithUserAgent returns a new client that sets the User-Agent header on all requests.
func wrapClientWithUserAgent(client *http.Client, userAgent string) *http.Client {
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &http.Client{
		Transport:     &userAgentTransport{base: transport, userAgent: userAgent},
		CheckRedirect: client.CheckRedirect,
		Jar:           client.Jar,
		Timeout:       client.Timeout,
	}
}
