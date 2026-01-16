package scout

import (
	"net/http"
	"runtime"
	"time"

	"golang.org/x/time/rate"
)

const Version = "0.1.0"

// Options configures the behavior of Query and related functions.
type Options struct {
	// HTTPClient is the client used for all requests. If nil, a default client with sensible timeouts is used.
	HTTPClient *http.Client

	// Parallelism controls how many sources run concurrently. Set to 1 for sequential execution.
	Parallelism int

	// GlobalRateLimit limits requests/second across all sources. Default is 0 (unlimited).
	GlobalRateLimit rate.Limit

	// SourceRateLimits sets per-source rate limits. Key is source name, value is requests/second.
	SourceRateLimits map[string]rate.Limit

	// Timeout is the per-source timeout.
	Timeout time.Duration

	// UserAgent is the User-Agent header sent with requests.
	UserAgent string

	// APIKeys maps source names to their API keys. Optional keys improve rate limits for some sources.
	APIKeys map[string]string
}

// Option is a functional option for configuring Query.
type Option func(*Options)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(o *Options) {
		o.HTTPClient = c
	}
}

// WithParallelism sets the number of concurrent sources.
func WithParallelism(n int) Option {
	return func(o *Options) {
		o.Parallelism = n
	}
}

// WithGlobalRateLimit sets a global rate limit (requests/second).
func WithGlobalRateLimit(rps float64) Option {
	return func(o *Options) {
		o.GlobalRateLimit = rate.Limit(rps)
	}
}

// WithSourceRateLimit sets a rate limit for a specific source.
func WithSourceRateLimit(source string, rps float64) Option {
	return func(o *Options) {
		if o.SourceRateLimits == nil {
			o.SourceRateLimits = make(map[string]rate.Limit)
		}
		o.SourceRateLimits[source] = rate.Limit(rps)
	}
}

// WithTimeout sets the per-source timeout.
func WithTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.Timeout = d
	}
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(o *Options) {
		o.UserAgent = ua
	}
}

// WithAPIKey sets an API key for a specific source.
func WithAPIKey(source, key string) Option {
	return func(o *Options) {
		if o.APIKeys == nil {
			o.APIKeys = make(map[string]string)
		}
		o.APIKeys[source] = key
	}
}

// defaultOptions returns Options with sensible defaults.
func defaultOptions() *Options {
	return &Options{
		Parallelism: runtime.NumCPU() * 2,
		Timeout:     30 * time.Second,
		UserAgent:   "Mozilla/5.0 (compatible; go-harden/scout-v" + Version + ")",
	}
}
