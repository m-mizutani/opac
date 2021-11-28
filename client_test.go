package opaclient_test

import (
	"context"
	"fmt"

	opaclient "github.com/m-mizutani/opa-go-client"
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
