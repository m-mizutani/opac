package mock

import "github.com/m-mizutani/opac"

func NewWithMock(f opac.MockFunc) *Foo {
	return &Foo{
		client: opac.NewMock(f),
	}
}
