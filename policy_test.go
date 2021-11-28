package opaclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPolicy(t *testing.T) {
	ctx := context.Background()
	client := setupClient(t)

	t.Run("GET policies", func(t *testing.T) {
		resp, err := client.ListPolicy(ctx)
		require.NoError(t, err)
		require.Len(t, resp, 1)
		assert.Equal(t, "testdata/policy/example.rego", resp[0].ID)
	})
}
