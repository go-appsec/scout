package sources

import (
	"regexp"
	"strings"
)

// SubdomainExtractor extracts subdomains matching a domain from text.
type SubdomainExtractor struct {
	domain  string
	pattern *regexp.Regexp
}

// NewSubdomainExtractor creates an extractor for the given domain.
func NewSubdomainExtractor(domain string) (*SubdomainExtractor, error) {
	// Match: subdomain.domain.tld (requires prefix before domain)
	// Case insensitive, allows wildcards (*), underscores, hyphens
	pattern, err := regexp.Compile(`(?i)[a-zA-Z0-9*_.-]+\.` + regexp.QuoteMeta(domain))
	if err != nil {
		return nil, err
	}
	return &SubdomainExtractor{domain: domain, pattern: pattern}, nil
}

// Extract finds all subdomains in the given text.
func (e *SubdomainExtractor) Extract(text string) []string {
	matches := e.pattern.FindAllString(text, -1)
	for i, m := range matches {
		matches[i] = strings.ToLower(m)
	}
	return matches
}

// URLExtractor extracts URLs matching a domain from text.
type URLExtractor struct {
	domain  string
	pattern *regexp.Regexp
}

// NewURLExtractor creates an extractor for URLs under the given domain.
func NewURLExtractor(domain string) (*URLExtractor, error) {
	// Match: http(s)://anything.domain.tld/path or http(s)://domain.tld/path
	quotedDomain := regexp.QuoteMeta(domain)
	pattern, err := regexp.Compile(`(?i)https?://(?:[a-zA-Z0-9_.-]+\.)?` + quotedDomain + `(?:/[^\s"'<>]*)?`)
	if err != nil {
		return nil, err
	}
	return &URLExtractor{domain: domain, pattern: pattern}, nil
}

// Extract finds all URLs in the given text.
func (e *URLExtractor) Extract(text string) []string {
	return e.pattern.FindAllString(text, -1)
}
