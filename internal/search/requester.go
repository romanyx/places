package search

import (
	"context"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/romanyx/places/internal/place"
)

// RequesterWithTrace decorates requester with trace.
type RequesterWithTrace struct {
	base Requester
}

// NewRequesterWithTrace initialize decorator.
func NewRequesterWithTrace(requester Requester) Requester {
	s := RequesterWithTrace{
		base: requester,
	}

	return &s
}

// Request decoraters search method.
func (s *RequesterWithTrace) Request(ctx context.Context, p Params) ([]place.Model, error) {
	_, span := trace.StartSpan(ctx, "requester.request")
	var err error
	var places []place.Model

	defer func() {
		if err != nil {
			switch errors.Cause(err) {
			case context.DeadlineExceeded, context.Canceled:
			default:
				span.SetStatus(trace.Status{Code: trace.StatusCodeUnknown, Message: err.Error()})
			}
			span.End()
			return
		}

		span.SetStatus(trace.Status{Code: trace.StatusCodeOK, Message: "request successed"})
		span.End()
	}()

	places, err = s.base.Request(ctx, p)
	return places, err
}
