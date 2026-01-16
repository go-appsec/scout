package scout

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-harden/scout/sources"
)

// mockSource creates a test source that yields the given results.
func mockSource(name string, yields sources.ResultType, results []sources.Result, errs []error) sources.Source {
	return sources.Source{
		Name:   name,
		Yields: yields,
		Run: func(_ context.Context, _ string, _ *http.Client, _ string) iter.Seq2[sources.Result, error] {
			return func(yield func(sources.Result, error) bool) {
				for _, err := range errs {
					if !yield(sources.Result{}, err) {
						return
					}
				}
				for _, r := range results {
					if !yield(r, nil) {
						return
					}
				}
			}
		},
	}
}

func TestQuery(t *testing.T) {
	t.Parallel()

	t.Run("returns_results", func(t *testing.T) {
		ctx := t.Context()

		src := mockSource("test", sources.Subdomain, []sources.Result{
			{Type: sources.Subdomain, Value: "api.example.com", Source: "test"},
			{Type: sources.Subdomain, Value: "www.example.com", Source: "test"},
		}, nil)

		var results []sources.Result
		for result, err := range Query(ctx, "example.com", []sources.Source{src}, WithParallelism(1)) {
			require.NoError(t, err)
			results = append(results, result)
		}

		assert.Len(t, results, 2)
	})

	t.Run("deduplicates_results", func(t *testing.T) {
		ctx := t.Context()

		src1 := mockSource("src1", sources.Subdomain, []sources.Result{
			{Type: sources.Subdomain, Value: "api.example.com", Source: "src1"},
		}, nil)
		src2 := mockSource("src2", sources.Subdomain, []sources.Result{
			{Type: sources.Subdomain, Value: "api.example.com", Source: "src2"},
		}, nil)

		var results []sources.Result
		for result, err := range Query(ctx, "example.com", []sources.Source{src1, src2}, WithParallelism(1)) {
			require.NoError(t, err)
			results = append(results, result)
		}

		assert.Len(t, results, 1)
	})

	t.Run("deduplicates_case_insensitive", func(t *testing.T) {
		ctx := t.Context()

		src := mockSource("test", sources.Subdomain, []sources.Result{
			{Type: sources.Subdomain, Value: "API.example.com", Source: "test"},
			{Type: sources.Subdomain, Value: "api.example.com", Source: "test"},
			{Type: sources.Subdomain, Value: "Api.Example.COM", Source: "test"},
		}, nil)

		var results []sources.Result
		for result, err := range Query(ctx, "example.com", []sources.Source{src}, WithParallelism(1)) {
			require.NoError(t, err)
			results = append(results, result)
		}

		assert.Len(t, results, 1)
	})

	t.Run("yields_errors", func(t *testing.T) {
		ctx := t.Context()

		testErr := errors.New("test error")
		src := mockSource("test", sources.Subdomain, []sources.Result{
			{Type: sources.Subdomain, Value: "api.example.com", Source: "test"},
		}, []error{testErr})

		var results []sources.Result
		var errs []error
		for result, err := range Query(ctx, "example.com", []sources.Source{src}, WithParallelism(1)) {
			if err != nil {
				errs = append(errs, err)
				continue
			}
			results = append(results, result)
		}

		assert.Len(t, errs, 1)
		assert.Len(t, results, 1)
	})

	t.Run("respects_context_cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		src := mockSource("test", sources.Subdomain, []sources.Result{
			{Type: sources.Subdomain, Value: "a.example.com", Source: "test"},
			{Type: sources.Subdomain, Value: "b.example.com", Source: "test"},
			{Type: sources.Subdomain, Value: "c.example.com", Source: "test"},
		}, nil)

		count := 0
		for range Query(ctx, "example.com", []sources.Source{src}, WithParallelism(1)) {
			count++
			if count == 1 {
				cancel()
				break
			}
		}

		assert.Equal(t, 1, count)
	})

	t.Run("respects_timeout", func(t *testing.T) {
		ctx := t.Context()

		slowSource := sources.Source{
			Name:   "slow",
			Yields: sources.Subdomain,
			Run: func(ctx context.Context, _ string, _ *http.Client, _ string) iter.Seq2[sources.Result, error] {
				return func(yield func(sources.Result, error) bool) {
					select {
					case <-ctx.Done():
						return
					case <-time.After(100 * time.Millisecond):
						yield(sources.Result{
							Type:   sources.Subdomain,
							Value:  "slow.example.com",
							Source: "slow",
						}, nil)
					}
				}
			},
		}

		var results []sources.Result
		for result, err := range Query(ctx, "example.com", []sources.Source{slowSource},
			WithParallelism(1),
			WithTimeout(10*time.Millisecond),
		) {
			if err != nil {
				continue
			}
			results = append(results, result)
		}

		assert.Empty(t, results)
	})

	t.Run("handles_empty_sources", func(t *testing.T) {
		ctx := t.Context()

		var count int
		for range Query(ctx, "example.com", nil, WithParallelism(1)) {
			count++
		}

		assert.Zero(t, count)
	})
}

func TestSubdomains(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	src := mockSource("test", sources.Subdomain|sources.URL, []sources.Result{
		{Type: sources.Subdomain, Value: "api.example.com", Source: "test"},
		{Type: sources.URL, Value: "https://example.com/path", Source: "test"},
		{Type: sources.Subdomain, Value: "www.example.com", Source: "test"},
	}, nil)

	subs := make([]string, 0, 2)
	for sub, err := range Subdomains(ctx, "example.com", []sources.Source{src}, WithParallelism(1)) {
		require.NoError(t, err)
		subs = append(subs, sub)
	}

	assert.Len(t, subs, 2)
	assert.True(t, slices.Contains(subs, "api.example.com"))
	assert.True(t, slices.Contains(subs, "www.example.com"))
}

func TestURLs(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	src := mockSource("test", sources.Subdomain|sources.URL, []sources.Result{
		{Type: sources.Subdomain, Value: "api.example.com", Source: "test"},
		{Type: sources.URL, Value: "https://example.com/path", Source: "test"},
		{Type: sources.URL, Value: "https://example.com/other", Source: "test"},
	}, nil)

	urls := make([]string, 0, 2)
	for url, err := range URLs(ctx, "example.com", []sources.Source{src}, WithParallelism(1)) {
		require.NoError(t, err)
		urls = append(urls, url)
	}

	assert.Len(t, urls, 2)
	assert.True(t, slices.Contains(urls, "https://example.com/path"))
	assert.True(t, slices.Contains(urls, "https://example.com/other"))
}

func TestDeduplicator(t *testing.T) {
	t.Parallel()

	d := &deduplicator{}

	t.Run("first_occurrence_false", func(t *testing.T) {
		assert.False(t, d.seen("test"))
	})

	t.Run("duplicate_returns_true", func(t *testing.T) {
		assert.True(t, d.seen("test"))
	})

	t.Run("case_insensitive", func(t *testing.T) {
		assert.True(t, d.seen("TEST"))
	})

	t.Run("trims_whitespace", func(t *testing.T) {
		assert.True(t, d.seen("  test  "))
	})

	t.Run("new_value_false", func(t *testing.T) {
		assert.False(t, d.seen("other"))
	})
}

func TestUserAgentTransport(t *testing.T) {
	t.Parallel()

	transport := &userAgentTransport{
		base:      http.DefaultTransport,
		userAgent: "test-agent/1.0",
	}

	var _ http.RoundTripper = transport
}
