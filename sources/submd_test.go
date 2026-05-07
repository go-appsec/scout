package sources

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubMD(t *testing.T) {
	t.Parallel()

	t.Run("registered", func(t *testing.T) {
		src := ByName("submd")
		require.NotNil(t, src)
		assert.Equal(t, Subdomain, src.Yields)
	})

	t.Run("integration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test")
		}

		ctx := t.Context()
		client := &http.Client{Timeout: 60 * time.Second}
		subdomains, _, errors := collectResults(SubMD.Run(ctx, client, "github.com", ""))

		if len(errors) > 0 {
			t.Logf("errors: %v", errors)
		}

		t.Logf("found %d subdomains", len(subdomains))
		assert.NotEmpty(t, subdomains)
		assertResults(t, subdomains, "submd", Subdomain)
	})
}
