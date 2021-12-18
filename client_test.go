package opac_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	opac "github.com/m-mizutani/opac"
	"github.com/stretchr/testify/require"
)

func ExampleClient() {
	client, err := opac.New("http://localhost:8181")
	if err != nil {
		panic(err)
	}

	req := opac.DataRequest{
		Path: "example/policy",
		Input: map[string]string{
			"user": "m-mizutani",
		},
	}
	resp := struct {
		Allowed bool `json:"allowed"`
	}{}

	if err := client.GetData(context.Background(), &req, &resp); err != nil {
		panic(err)
	}

	fmt.Println("allowed? =>", resp.Allowed)
}

func setupClient(t *testing.T) *opac.Client {
	url, ok := os.LookupEnv("OPA_BASE_URL")
	if !ok {
		t.Skip("OPA_BASE_URL is not set")
	}

	client, err := opac.New(url)
	require.NoError(t, err)
	return client
}
