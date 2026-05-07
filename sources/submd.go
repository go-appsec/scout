package sources

import (
	"bufio"
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strings"
)

func init() {
	Register(SubMD)
}

// SubMD queries the sub.md API for subdomains.
var SubMD = Source{
	Name:   "submd",
	Yields: Subdomain,
	Run:    runSubMD,
}

func runSubMD(ctx context.Context, client *http.Client, domain string, _ string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		apiURL := "https://api.sub.md/v1/search?apex=" + url.QueryEscape(domain)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			yield(Result{}, fmt.Errorf("submd: %w", err))
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			yield(Result{}, fmt.Errorf("submd: %w", err))
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			yield(Result{}, fmt.Errorf("submd: unexpected status %d", resp.StatusCode))
			return
		}

		extractor, err := NewSubdomainExtractor(domain)
		if err != nil {
			yield(Result{}, fmt.Errorf("submd: %w", err))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			for _, sub := range extractor.Extract(line) {
				if !yield(Result{Type: Subdomain, Value: sub, Source: "submd"}, nil) {
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			yield(Result{}, fmt.Errorf("submd: %w", err))
		}
	}
}
