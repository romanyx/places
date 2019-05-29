package search

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/romanyx/places/internal/place"
	"github.com/romanyx/places/internal/storage"
)

func TestServiceSearch(t *testing.T) {
	tt := []struct {
		name          string
		requesterFunc func(ctx context.Context, p Params) ([]place.Model, error)
		repoFunc      func(m *MockRepository)
		cacheResponse bool
		expectErr     bool
	}{
		{
			name: "happy path",
			requesterFunc: func(ctx context.Context, p Params) ([]place.Model, error) {
				return make([]place.Model, 0), nil
			},
			repoFunc: func(m *MockRepository) {
				m.EXPECT().
					Cache(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			cacheResponse: true,
		},
		{
			name: "request timeout",
			requesterFunc: func(ctx context.Context, p Params) ([]place.Model, error) {
				return nil, context.DeadlineExceeded
			},
			repoFunc: func(m *MockRepository) {
				m.EXPECT().
					Retrieve(gomock.Any(), gomock.Any()).
					Return(make([]place.Model, 0), nil)
			},
		},
		{
			name: "retrieve no cache",
			requesterFunc: func(ctx context.Context, p Params) ([]place.Model, error) {
				return nil, context.DeadlineExceeded
			},
			repoFunc: func(m *MockRepository) {
				m.EXPECT().
					Retrieve(gomock.Any(), gomock.Any()).
					Return(nil, storage.ErrCacheNotFound)
			},
			expectErr: true,
		},
		{
			name: "context cancel",
			requesterFunc: func(ctx context.Context, p Params) ([]place.Model, error) {
				return nil, context.Canceled
			},
			repoFunc: func(m *MockRepository) {},
		},
	}

	doneChan := make(chan struct{})
	cached = func() {
		doneChan <- struct{}{}
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := NewMockRepository(ctrl)
			tc.repoFunc(repo)

			s := NewService(requesterFunc(tc.requesterFunc), repo, time.Second)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_, err := s.Search(ctx, Params{})

			if tc.cacheResponse {
				select {
				case <-doneChan:
				case <-time.After(3 * time.Second):
					t.Error("expected to cache response")
				}
			}

			if tc.expectErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

type requesterFunc func(context.Context, Params) ([]place.Model, error)

func (f requesterFunc) Request(ctx context.Context, q Params) ([]place.Model, error) {
	return f(ctx, q)
}
