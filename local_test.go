package opac_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/opac"
)

func TestLocal(t *testing.T) {
	type testCase struct {
		src      opac.Source
		query    string
		input    map[string]any
		output   map[string]any
		newErr   bool
		queryErr bool
	}

	doTest := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			client, err := opac.New(tc.src)
			if tc.newErr {
				gt.Error(t, err)
				return
			}
			gt.NoError(t, err)
			ctx := context.Background()

			var output map[string]any
			err = client.Query(ctx, tc.query, tc.input, &output)
			if tc.queryErr {
				gt.Error(t, err)
				return
			}
			gt.NoError(t, err)
			gt.Equal(t, tc.output, output)
		}
	}

	t.Run("success", doTest(testCase{
		src:   opac.SrcFiles("testdata/local/f1.rego", "testdata/local/f2.rego"),
		query: "data.color",
		input: map[string]any{
			"color": "blue",
		},
		output: map[string]any{
			"number": float64(5),
		},
	}))

	t.Run("no policy file", doTest(testCase{
		src:    opac.SrcFiles("testdata/empty"),
		query:  "data.color",
		newErr: true,
	}))

	t.Run("no policy content", doTest(testCase{
		src:   opac.SrcFiles("testdata/no_content/policy.rego"),
		query: "data.color",
		input: map[string]any{
			"color": "blue",
		},
		output: map[string]any{},
	}))

	t.Run("policy data", doTest(testCase{
		src:   opac.SrcData(map[string]string{"policy.rego": "package color\nnumber := 5 { input.color == \"blue\" }"}),
		query: "data.color",
		input: map[string]any{
			"color": "blue",
		},
		output: map[string]any{
			"number": float64(5),
		},
	}))
}
