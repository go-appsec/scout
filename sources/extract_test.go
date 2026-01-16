package sources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubdomainExtractor(t *testing.T) {
	t.Parallel()

	t.Run("creates_extractor", func(t *testing.T) {
		e, err := NewSubdomainExtractor("example.com")

		require.NoError(t, err)
		assert.NotNil(t, e)
	})

	t.Run("special_domain", func(t *testing.T) {
		e, err := NewSubdomainExtractor("example.co.uk")
		require.NoError(t, err)

		got := e.Extract("api.example.co.uk found")

		assert.Equal(t, []string{"api.example.co.uk"}, got)
	})
}

func TestSubdomainExtractor(t *testing.T) {
	t.Parallel()

	e, err := NewSubdomainExtractor("example.com")
	require.NoError(t, err)

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single_subdomain",
			input: "Found api.example.com in text",
			want:  []string{"api.example.com"},
		},
		{
			name:  "multiple_subdomains",
			input: "Multiple: a.example.com and b.example.com",
			want:  []string{"a.example.com", "b.example.com"},
		},
		{
			name:  "no_match",
			input: "No match here",
			want:  nil,
		},
		{
			name:  "different_domain",
			input: "sub.other.com is not a match",
			want:  nil,
		},
		{
			name:  "uppercase_lowercased",
			input: "UPPER.EXAMPLE.COM should lowercase",
			want:  []string{"upper.example.com"},
		},
		{
			name:  "wildcard",
			input: "*.example.com wildcard",
			want:  []string{"*.example.com"},
		},
		{
			name:  "nested_subdomain",
			input: "deep.nested.sub.example.com",
			want:  []string{"deep.nested.sub.example.com"},
		},
		{
			name:  "with_hyphen",
			input: "my-api.example.com",
			want:  []string{"my-api.example.com"},
		},
		{
			name:  "with_underscore",
			input: "my_service.example.com",
			want:  []string{"my_service.example.com"},
		},
		{
			name:  "mixed_case",
			input: "Api.Example.COM mixed case",
			want:  []string{"api.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, e.Extract(tt.input))
		})
	}
}

func TestNewURLExtractor(t *testing.T) {
	t.Parallel()

	t.Run("creates_extractor", func(t *testing.T) {
		e, err := NewURLExtractor("example.com")

		require.NoError(t, err)
		assert.NotNil(t, e)
	})

	t.Run("special_domain", func(t *testing.T) {
		e, err := NewURLExtractor("example.co.uk")
		require.NoError(t, err)

		got := e.Extract("https://api.example.co.uk/path")

		assert.Equal(t, []string{"https://api.example.co.uk/path"}, got)
	})
}

func TestURLExtractor(t *testing.T) {
	t.Parallel()

	e, err := NewURLExtractor("example.com")
	require.NoError(t, err)

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "https_with_path",
			input: "Visit https://example.com/path",
			want:  []string{"https://example.com/path"},
		},
		{
			name:  "http_subdomain_path",
			input: "http://api.example.com/v1/users",
			want:  []string{"http://api.example.com/v1/users"},
		},
		{
			name:  "different_domain_no_match",
			input: "https://other.com/page",
			want:  nil,
		},
		{
			name:  "multiple_urls",
			input: "Multiple: https://a.example.com and https://b.example.com/x",
			want:  []string{"https://a.example.com", "https://b.example.com/x"},
		},
		{
			name:  "url_without_path",
			input: "https://example.com",
			want:  []string{"https://example.com"},
		},
		{
			name:  "url_with_query",
			input: "https://example.com/search?q=test&page=1",
			want:  []string{"https://example.com/search?q=test&page=1"},
		},
		{
			name:  "url_with_fragment",
			input: "https://example.com/page#section",
			want:  []string{"https://example.com/page#section"},
		},
		{
			name:  "url_in_quotes",
			input: `href="https://example.com/path"`,
			want:  []string{"https://example.com/path"},
		},
		{
			name:  "url_in_html",
			input: `<a href="https://example.com/link">click</a>`,
			want:  []string{"https://example.com/link"},
		},
		{
			name:  "deep_subdomain_url",
			input: "https://api.v2.example.com/endpoint",
			want:  []string{"https://api.v2.example.com/endpoint"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, e.Extract(tt.input))
		})
	}
}
