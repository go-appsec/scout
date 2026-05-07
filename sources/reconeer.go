package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
)

func init() {
	Register(Reconeer)
}

// Reconeer queries the Reconeer API for subdomains.
// API key required.
var Reconeer = Source{
	Name:         "reconeer",
	Yields:       Subdomain,
	AuthRequired: true,
	Run:          runReconeer,
}

func runReconeer(ctx context.Context, client *http.Client, domain string, apiKey string) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		url := "https://www.reconeer.com/api/domain/" + domain

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			yield(Result{}, fmt.Errorf("reconeer: %w", err))
			return
		}
		req.Header.Set("Accept", "application/json")
		if apiKey != "" {
			req.Header.Set("X-API-KEY", apiKey)
		}

		resp, err := client.Do(req)
		if err != nil {
			yield(Result{}, fmt.Errorf("reconeer: %w", err))
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			yield(Result{}, fmt.Errorf("reconeer: unexpected status %d", resp.StatusCode))
			return
		}

		var response struct {
			Subdomains []struct {
				Subdomain string `json:"subdomain"`
			} `json:"subdomains"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			yield(Result{}, fmt.Errorf("reconeer: %w", err))
			return
		}

		for _, s := range response.Subdomains {
			if s.Subdomain == "" {
				continue
			}
			if !yield(Result{Type: Subdomain, Value: s.Subdomain, Source: "reconeer"}, nil) {
				return
			}
		}
	}
}
