package scout

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestWithParallelism(t *testing.T) {
	t.Parallel()

	opts := defaultOptions()
	WithParallelism(4)(opts)

	assert.Equal(t, 4, opts.Parallelism)
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	opts := defaultOptions()
	WithTimeout(60 * time.Second)(opts)

	assert.Equal(t, 60*time.Second, opts.Timeout)
}

func TestWithGlobalRateLimit(t *testing.T) {
	t.Parallel()

	opts := defaultOptions()
	WithGlobalRateLimit(10)(opts)

	assert.InDelta(t, float64(rate.Limit(10)), float64(opts.GlobalRateLimit), 0.001)
}

func TestWithSourceRateLimit(t *testing.T) {
	t.Parallel()

	opts := defaultOptions()

	t.Run("first_source", func(t *testing.T) {
		WithSourceRateLimit("wayback", 5)(opts)

		require.NotNil(t, opts.SourceRateLimits)
		assert.InDelta(t, float64(rate.Limit(5)), float64(opts.SourceRateLimits["wayback"]), 0.001)
	})

	t.Run("second_source", func(t *testing.T) {
		WithSourceRateLimit("crtsh", 3)(opts)

		assert.InDelta(t, float64(rate.Limit(3)), float64(opts.SourceRateLimits["crtsh"]), 0.001)
	})

	t.Run("first_source_unchanged", func(t *testing.T) {
		assert.InDelta(t, float64(rate.Limit(5)), float64(opts.SourceRateLimits["wayback"]), 0.001)
	})
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	opts := defaultOptions()
	WithUserAgent("test-agent/1.0")(opts)

	assert.Equal(t, "test-agent/1.0", opts.UserAgent)
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	opts := defaultOptions()
	custom := &http.Client{Timeout: 5 * time.Second}
	WithHTTPClient(custom)(opts)

	assert.Same(t, custom, opts.HTTPClient)
}
