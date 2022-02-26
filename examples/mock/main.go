package mock

import (
	"context"

	"github.com/m-mizutani/opac"
)

type Foo struct {
	client opac.Client
}

type Input struct{ User string }
type Result struct{ Allow bool }

func New(url string) *Foo {
	client, err := opac.NewRemote(url)
	if err != nil {
		panic(err)
	}

	return &Foo{
		client: client,
	}
}

func (x *Foo) IsAllow(user string) bool {
	input := &Input{User: user}
	var result Result
	if err := x.client.Query(context.Background(), input, &result); err != nil {
		panic(err)
	}

	return result.Allow
}
