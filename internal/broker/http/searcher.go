package http

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/romanyx/places/internal/log"
	"github.com/romanyx/places/internal/place"
	"github.com/romanyx/places/internal/search"
)

// SearcherWithLog decorates searcher with logging.
type SearcherWithLog struct {
	base Searcher
}

// NewSearcherWithLog initialize decorator.
func NewSearcherWithLog(searcher Searcher) Searcher {
	s := SearcherWithLog{
		base: searcher,
	}

	return &s
}

// Search decoraters search method.
func (s *SearcherWithLog) Search(ctx context.Context, p search.Params) ([]place.Model, error) {
	var places []place.Model
	var err error
	start := time.Now()
	spanCtx := trace.FromContext(ctx).SpanContext()
	log.Debug("search processing", map[string]interface{}{
		"trace_id": spanCtx.TraceID,
		"term":     p.Term,
		"language": p.Locale,
		"types":    p.Types,
	})

	defer func() {
		if err != nil && errors.Cause(err) != search.ErrUnavailable {
			log.Error(errors.Wrap(err, "search error"), map[string]interface{}{
				"trace_id": spanCtx.TraceID,
				"term":     p.Term,
				"language": p.Locale,
				"types":    p.Types,
				"elapsed":  time.Since(start),
			})
			return
		}

		log.Debug("search processed", map[string]interface{}{
			"trace_id": spanCtx.TraceID,
			"term":     p.Term,
			"language": p.Locale,
			"types":    p.Types,
			"elapsed":  time.Since(start),
		})
	}()

	places, err = s.base.Search(ctx, p)
	return places, err
}

// SearcherWithTrace decorates searcher with trace.
type SearcherWithTrace struct {
	base Searcher
}

// NewSearcherWithTrace initialize decorator.
func NewSearcherWithTrace(searcher Searcher) Searcher {
	s := SearcherWithTrace{
		base: searcher,
	}

	return &s
}

// Search decoraters search method.
func (s *SearcherWithTrace) Search(ctx context.Context, p search.Params) ([]place.Model, error) {
	_, span := trace.StartSpan(ctx, "search.service")
	span.AddAttributes(
		trace.StringAttribute("term", p.Term),
		trace.StringAttribute("locale", p.Locale),
		trace.StringAttribute("types", fmt.Sprint(p.Types)),
	)
	var err error
	var places []place.Model

	defer func() {
		if err != nil && errors.Cause(err) != search.ErrUnavailable {
			span.SetStatus(trace.Status{Code: trace.StatusCodeUnknown, Message: err.Error()})
			span.End()
			return
		}

		span.SetStatus(trace.Status{Code: trace.StatusCodeOK, Message: "search successed"})
		span.End()
	}()

	places, err = s.base.Search(ctx, p)
	return places, err
}
