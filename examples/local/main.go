package main

import (
	"context"
	"fmt"

	"github.com/m-mizutani/opac"
)

func main() {
	client, err := opac.NewLocal(
		opac.WithFile("./examples/local/policy.rego"),
		opac.WithPackage("example"),
	)
	if err != nil {
		panic(err.Error())
	}

	input := struct {
		Color string `json:"color"`
	}{
		Color: "blue",
	}
	output := struct {
		Allow bool `json:"allow"`
	}{}

	if err := client.Query(context.Background(), input, &output); err != nil {
		panic(err.Error())
	}

	fmt.Println("result:", output)
}
