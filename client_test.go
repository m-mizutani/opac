package opaclient_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	opaclient "github.com/m-mizutani/opa-go-client"
	"github.com/stretchr/testify/require"
)

func ExampleClient() {
	client, err := opaclient.New("http://localhost:8181")
	if err != nil {
		panic(err)
	}

	req := opaclient.DataRequest{
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

func setupClient(t *testing.T) *opaclient.Client {
	url, ok := os.LookupEnv("OPA_BASE_URL")
	if !ok {
		t.Skip("OPA_BASE_URL is not set")
	}

	client, err := opaclient.New(url)
	require.NoError(t, err)
	return client
}
