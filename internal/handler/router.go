package handler

import (
	"context"

	"github.com/julienschmidt/httprouter"
)

type Endpoint struct {
	Path    string
	Method  string
	Handler Handler
}

type Router interface {
	Serve(ctx context.Context, addr string) error
}

func NewRouter(endpoints []Endpoint) Router {
	r := httprouterRouter{router: httprouter.New()}
	for _, e := range endpoints {
		r.addHandler(e)
	}

	return &r
}
