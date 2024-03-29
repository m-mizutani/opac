package opac_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/opac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalClient(t *testing.T) {
	t.Run("import recursive if specifing directory", func(t *testing.T) {
		client, err := opac.NewLocal(opac.WithDir("./testdata"))
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
		client, err := opac.NewLocal(opac.WithFile("./testdata/policy.rego"))
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
		_, err := opac.NewLocal(opac.WithFile("./testdata/not_found.rego"))
		require.Error(t, err)
	})

	t.Run("with package", func(t *testing.T) {
		client, err := opac.NewLocal(
			opac.WithFile("./testdata/policy.rego"),
			opac.WithPackage("color"),
		)
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

	t.Run("with print", func(t *testing.T) {
		var buf bytes.Buffer
		client, err := opac.NewLocal(
			opac.WithFile("./testdata/print.rego"),
			opac.WithPackage("print"),
			opac.WithRegoPrint(&buf),
		)
		require.NoError(t, err)
		in := map[string]string{
			"user": "blue",
		}
		out := map[string]interface{}{}
		require.NoError(t, client.Query(context.Background(), in, &out))
		assert.Equal(t, true, out["allow"])
		assert.Equal(t, "testdata/print.rego:4 blue", buf.String())
	})

	t.Run("with policy data", func(t *testing.T) {
		client, err := opac.NewLocal(
			opac.WithPolicyData("mypolicy", `package color
			allow {
				input.color == "orange"
			}`),
			opac.WithPackage("color"),
		)
		require.NoError(t, err)
		in := map[string]string{
			"color": "orange",
		}
		out := map[string]interface{}{}
		require.NoError(t, client.Query(context.Background(), in, &out))
		assert.Equal(t, true, out["allow"])
	})

	t.Run("failed if no policy data", func(t *testing.T) {
		client, err := opac.NewLocal()
		assert.ErrorIs(t, err, opac.ErrNoPolicyData)
		assert.Nil(t, client)
	})
}

func TestWithPackageSuffix(t *testing.T) {
	policy := `package color.test
	allow = true
	`
	client := gt.R1(opac.NewLocal(
		opac.WithPolicyData("mypolicy", policy),
		opac.WithPackage("color"),
	)).NoError(t)

	var r1 map[string]any
	gt.NoError(t, client.Query(context.Background(), map[string]string{}, &r1))
	gt.V(t, r1["allow"]).Nil()

	var r2 map[string]any
	gt.NoError(t, client.Query(context.Background(), map[string]string{}, &r2, opac.WithPackageSuffix(".test")))
	gt.V(t, r2["allow"]).Equal(true)
}
