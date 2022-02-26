package opac_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/opac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalClient(t *testing.T) {
	t.Run("import recursive if specifing directory", func(t *testing.T) {
		client, err := opac.NewLocal("./testdata")
		require.NoError(t, err)
		in := map[string]string{
			"color":  "blue",
			"number": "five",
		}
		out := map[string]map[string]interface{}{}
		require.NoError(t, client.Query(context.Background(), in, &out))
		assert.Equal(t, true, out["color"]["allow"])
		assert.Equal(t, true, out["number"]["allow"])
	})

	t.Run("import a file if specifing file path", func(t *testing.T) {
		client, err := opac.NewLocal("./testdata/policy.rego")
		require.NoError(t, err)
		in := map[string]string{
			"color":  "blue",
			"number": "five",
		}
		out := map[string]map[string]interface{}{}
		require.NoError(t, client.Query(context.Background(), in, &out))
		assert.Equal(t, true, out["color"]["allow"])
		assert.Equal(t, nil, out["number"]["allow"])
	})

	t.Run("fail by specifying invalid path", func(t *testing.T) {
		_, err := opac.NewLocal("./testdata/not_found.rego")
		require.Error(t, err)
	})

	t.Run("with package", func(t *testing.T) {
		client, err := opac.NewLocal("./testdata/policy.rego", opac.WithPackage("color"))
		require.NoError(t, err)
		in := map[string]string{
			"color":  "blue",
			"number": "five",
		}
		out := map[string]interface{}{}
		require.NoError(t, client.Query(context.Background(), in, &out))
		assert.Equal(t, true, out["allow"])
		assert.Nil(t, out["color"])
		assert.Nil(t, out["number"])

	})
}
