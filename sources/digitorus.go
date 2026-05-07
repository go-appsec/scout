package sources

import (
	"bufio"
	"context"
	"fmt"
	"iter"
	"net/http"
	"strings"
)

func init() {
	Register(Digitorus)
}

// Digitorus queries the CertificateDetails website for subdomains.
// API key required.
var Digitorus = Source{
	Name:         "digitorus",
	Yields:       Subdomain,
	AuthRequired: true,
	Run:          runDigitorus,
}

func runDigitorus(ctx context.Context, client *http.Client, domain string, _ string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		url := "https://certificatedetails.com/" + domain

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			yield(Result{}, fmt.Errorf("digitorus: %w", err))
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			yield(Result{}, fmt.Errorf("digitorus: %w", err))
			return
		}
		defer func() { _ = resp.Body.Close() }()

		// 404 pages still contain subdomains, treat as success
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			yield(Result{}, fmt.Errorf("digitorus: unexpected status %d", resp.StatusCode))
			return
		}

		extractor, err := NewSubdomainExtractor(domain)
		if err != nil {
			yield(Result{}, fmt.Errorf("digitorus: %w", err))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			for _, sub := range extractor.Extract(scanner.Text()) {
				// Trim leading dots from extracted subdomains
				sub = strings.TrimLeft(sub, ".")
				if !yield(Result{Type: Subdomain, Value: sub, Source: "digitorus"}, nil) {
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			yield(Result{}, fmt.Errorf("digitorus: %w", err))
		}
	}
}
