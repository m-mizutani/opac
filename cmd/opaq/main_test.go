package main_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/m-mizutani/opac/cmd/opaq"
)

func TestIsEmpty(t *testing.T) {
	testCases := []struct {
		title string
		input string
		exp   bool
	}{
		{
			title: "empty array",
			input: "[]",
			exp:   true,
		},
		{
			title: "array has number",
			input: "[1]",
			exp:   false,
		},
		{
			title: "empty map",
			input: "{}",
			exp:   true,
		},
		{
			title: "map has a value",
			input: `{"a":1}`,
			exp:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			var d interface{}
			require.NoError(t, json.Unmarshal([]byte(tc.input), &d))
			assert.Equal(t, tc.exp, main.IsEmpty(d))
		})
	}
}
