package main

import (
	"context"
	"fmt"

	"github.com/m-mizutani/opac"
)

func main() {
	client, err := opac.NewRemote("https://opa-server-h6tk4k5hyq-an.a.run.app/v1/data/example")
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
