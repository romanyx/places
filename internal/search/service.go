package search

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/romanyx/places/internal/broker"
	"github.com/romanyx/places/internal/log"
	"github.com/romanyx/places/internal/place"
	"github.com/romanyx/places/internal/storage"
)

//go:generate mockgen -package=search -destination=service.mock_test.go -source=service.go Repository

var (
	// ErrUnavailable returns when request failed and cache not found.
	ErrUnavailable = errors.New("places unavailable")
)

// Repository is a data access layer.
type Repository interface {
	Cache(context.Context, Params, []place.Model) error
	Retrieve(context.Context, Params) ([]place.Model, error)
}

// Requester requests aviasales places endpoint.
type Requester interface {
	Request(context.Context, Params) ([]place.Model, error)
}

// NewService initialize search service.
func NewService(rq Requester, repo Repository, timeout time.Duration) *Service {
	s := Service{
		Requester:  rq,
		Repository: repo,
		timeout:    timeout,
	}

	return &s
}

// Params represents params by which search will be done.
type Params struct {
	Term   string   `url:"term"`
	Locale string   `url:"locale"`
	Types  []string `url:"types"`
}

// Service contains domain logic for finding process.
type Service struct {
	Requester
	Repository
	timeout time.Duration
}

// Search searches place in aviasales. It will try to
// make request using requester - if request fails will
// try to retieve cache and return it, otherwise will
// save cache of the request and return result.
// If request will fail or timeout and there would not be
// any cache in storage will return ErrUnavailable.
//
// TODO(romanyx): Think about idea that first step should be
// cache retrieve and only then, if cache found - there should
// be timeout for request, otherwise when cache is not found in
// a storage request should be without timeout.
func (s *Service) Search(ctx context.Context, p Params) ([]place.Model, error) {
	spanCtx := trace.FromContext(ctx).SpanContext()
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	places, err := s.Request(ctx, p)
	if err != nil {
		// Log unexpected error.
		switch errors.Cause(err) {
		case context.DeadlineExceeded:
			// Retrieve places from cache if request deadline exceeded.
		case context.Canceled:
			// Return when cancelled no need to process futher.
			return places, nil
		case broker.ErrBadRequest:
			// When aviasales server returns bad request show it.
			return nil, err
		default:
			log.Warn(errors.Wrap(err, "unexpected error on request"), map[string]interface{}{
				"trace_id": spanCtx.TraceID,
			})
		}

		// Continue to retrive cache.
		places, err = s.Retrieve(ctx, p)
		if err != nil {
			if errors.Cause(err) != storage.ErrCacheNotFound {
				// Log error only if it is unexpected cache not found is
				// expected one.
				log.Error(errors.Wrap(err, "unexpected error on retrieve"), map[string]interface{}{
					"trace_id": spanCtx.TraceID,
				})
			}
			return nil, ErrUnavailable
		}
		return places, nil
	}

	// Save cache of request if it was successfull.
	go func() {
		if err := s.Cache(ctx, p, places); err != nil {
			log.Error(errors.Wrap(err, "cache failed"), map[string]interface{}{
				"trace_id": spanCtx.TraceID,
			})
		}
		cached()
	}()

	return places, nil
}

var cached = func() {}
