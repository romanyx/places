package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"

	"github.com/romanyx/places/internal/broker"
	"github.com/romanyx/places/internal/log"
	"github.com/romanyx/places/internal/place"
	"github.com/romanyx/places/internal/search"
)

const (
	timeout = 30 * time.Second
)

// Searcher represents search interface.
type Searcher interface {
	Search(context.Context, search.Params) ([]place.Model, error)
}

// Handler allows to handle requests.
type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request) error
}

type searchHandler struct {
	Searcher
}

func newSearchHandler(searcher Searcher) http.Handler {
	searchHandler := searchHandler{
		Searcher: searcher,
	}

	h := httpHandler{&searchHandler}
	return &h
}

func (h *searchHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return methodNotAllowedResponse(w)
	}

	if err := r.ParseForm(); err != nil {
		return errors.Wrap(err, "parse form")
	}
	var params search.Params
	setParams(&params, r.Form)

	places, err := h.Search(r.Context(), params)
	if err != nil {
		switch errors.Cause(err) {
		case search.ErrUnavailable:
			return unavailableResponse(w)
		case broker.ErrBadRequest:
			return badRequestResponse(w)
		default:
			return internalServerErrorResponse(w)
		}
	}

	if err := json.NewEncoder(w).Encode(&places); err != nil {
		return errors.Wrap(err, "encode json")
	}

	return nil
}

func setParams(params *search.Params, f url.Values) {
	if len(f["term"]) > 0 {
		params.Term = f["term"][0]
	}
	if len(f["locale"]) > 0 {
		params.Locale = f["locale"][0]
	}
	params.Types = f["types[]"]
}

// NewServer initialize http.Server.
func NewServer(addr string, searcher Searcher) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/places", ochttp.WithRouteTag(newSearchHandler(searcher), "/places"))

	s := http.Server{
		Addr: addr,
		Handler: &ochttp.Handler{
			Handler:     mux,
			Propagation: &b3.HTTPFormat{},
		},
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	return &s
}

// httpHandler allows to implement ServeHTTP for Handler.
type httpHandler struct {
	Handler
}

// ServeHTTP implements http.Handler.
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.Handle(w, r); err != nil {
		spanCtx := trace.FromContext(r.Context()).SpanContext()
		log.Error(errors.Wrap(err, "serve http"), map[string]interface{}{
			"trace_id": spanCtx.TraceID,
		})
	}
}

func internalServerErrorResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusInternalServerError)
	return nil
}

func unavailableResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusServiceUnavailable)
	return nil
}

func badRequestResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusBadRequest)
	return nil
}

func methodNotAllowedResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return nil
}
