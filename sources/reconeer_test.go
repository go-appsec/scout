package sources

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconeer(t *testing.T) {
	t.Parallel()

	t.Run("registered", func(t *testing.T) {
		src := ByName("reconeer")
		require.NotNil(t, src)
		assert.Equal(t, Subdomain, src.Yields)
	})

	t.Run("integration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test")
		}
		apiKey := os.Getenv("RECONEER_API_KEY")
		if apiKey == "" {
			t.Skip("RECONEER_API_KEY not set")
		}

		ctx := t.Context()
		client := &http.Client{Timeout: 30 * time.Second}
		subdomains, _, errors := collectResults(Reconeer.Run(ctx, client, "github.com", apiKey))

		if len(errors) > 0 {
			t.Logf("errors: %v", errors)
		}

		t.Logf("found %d subdomains", len(subdomains))
		assert.NotEmpty(t, subdomains)
		assertResults(t, subdomains, "reconeer", Subdomain)
	})
}
