package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
	"github.com/standoffvenus/future/internal/empty"
)

type httprouterRouter struct {
	router *httprouter.Router

	oncer        sync.Once
	shutdownChan chan struct{}
	err          error
}

var _ Router = new(httprouterRouter)

func (r *httprouterRouter) Serve(ctx context.Context, addr string) error {
	r.oncer.Do(func() {
		srv := http.Server{
			Addr:    addr,
			Handler: r.router,
		}

		shutdown := make(chan struct{})
		go func() {
			defer close(shutdown)

			log.
				Info().
				Fields(map[string]any{
					"addr": srv.Addr,
				}).
				Msg("Starting server.")

			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.
					Error().
					Err(err).
					Msg("Server could not start.")

				r.err = err
			}
		}()

		select {
		case <-shutdown:
		case <-ctx.Done():
			log.Info().Msg("Shutting down server.")

			srv.Close()
			<-shutdown
		}
	})

	return r.err
}

func (r *httprouterRouter) addHandler(endpoint Endpoint) {
	r.router.Handle(
		endpoint.Method,
		endpoint.Path,
		makeHTTPRouterHandler(endpoint.Handler))

}

func makeHTTPRouterHandler(h Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		resp, err := h(Request{
			Context:         r.Context(),
			PathParameters:  paramsToMap(p),
			QueryParameters: r.URL.Query(),
			Headers:         r.Header,
			Body:            r.Body,
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.
				Error().
				Err(err).
				Msg("Service encountered error.")
		} else {
			w.WriteHeader(resp.Code)
			writeHeaders(w.Header(), resp.Headers)
			io.Copy(w, resp.Body)
		}
	}
}

func writeHeaders(headers http.Header, newHeaders map[string][]string) {
	for name, values := range newHeaders {
		for _, v := range values {
			headers.Add(name, v)
		}
	}
}

func paramsToMap(params httprouter.Params) map[string]string {
	m := make(map[string]string, len(params))
	for _, p := range params {
		if !empty.String(p.Value) {
			m[p.Key] = p.Value
		}
	}

	return m
}
