package search

import (
	"context"

	"go.opencensus.io/trace"

	"github.com/romanyx/places/internal/place"
)

// RepositoryWithTrace decorates requester with trace.
type RepositoryWithTrace struct {
	base Repository
}

// NewRepositoryWithTrace initialize decorator.
func NewRepositoryWithTrace(repository Repository) Repository {
	r := RepositoryWithTrace{
		base: repository,
	}

	return &r
}

// Cache decoraters cache method.
func (s *RepositoryWithTrace) Cache(ctx context.Context, p Params, places []place.Model) error {
	_, span := trace.StartSpan(ctx, "repository.cache")
	var err error

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeUnknown, Message: err.Error()})
			span.End()
			return
		}

		span.SetStatus(trace.Status{Code: trace.StatusCodeOK, Message: "cache successed"})
		span.End()
	}()

	err = s.base.Cache(ctx, p, places)
	return err
}

// Retrieve decoraters retrieve method.
func (s *RepositoryWithTrace) Retrieve(ctx context.Context, p Params) ([]place.Model, error) {
	_, span := trace.StartSpan(ctx, "repository.retrieve")
	var err error
	var places []place.Model

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeUnknown, Message: err.Error()})
			span.End()
			return
		}

		span.SetStatus(trace.Status{Code: trace.StatusCodeOK, Message: "retrieve successed"})
		span.End()
	}()

	places, err = s.base.Retrieve(ctx, p)
	return places, err
}
