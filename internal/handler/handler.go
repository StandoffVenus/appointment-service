package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

type Request struct {
	Context         context.Context
	Headers         http.Header
	PathParameters  map[string]string
	QueryParameters url.Values
	Body            io.ReadCloser
}

type Response struct {
	Code    int
	Headers http.Header
	Body    io.Reader
}

type Error struct {
	Message string `json:",omitempty"`
}

type Handler func(Request) (Response, error)

func OK[T any](t T) Response {
	return MakeResponse(t, http.StatusOK)
}

func NoContent() Response {
	return MakeResponse("", http.StatusNoContent)
}

func BadRequest(msg string) Response {
	return MakeResponse(Error{Message: msg}, http.StatusBadRequest)
}

func Conflict(msg string) Response {
	return MakeResponse(Error{Message: msg}, http.StatusConflict)
}

func MakeResponse[T any](t T, status int) Response {
	var buffer bytes.Buffer
	if err := jsoniter.NewEncoder(&buffer).Encode(t); err != nil {
		log.
			Warn().
			Err(err).
			Msg("Could not marshal HTTP response to JSON.")

		buffer.Reset()
	}

	headers := make(http.Header)
	headers.Add("Content-Type", "application/json")

	return Response{
		Code:    status,
		Body:    &buffer,
		Headers: headers,
	}
}

func ParseBody[T any](r Request) (T, error) {
	defer r.Body.Close()

	var t T
	if err := jsoniter.NewDecoder(r.Body).Decode(&t); err != nil {
		return t, err
	}

	return t, nil
}
