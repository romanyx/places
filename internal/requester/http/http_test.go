package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/romanyx/places/internal/place"
	"github.com/romanyx/places/internal/search"
)

func Test_queryToString(t *testing.T) {
	expect := "locale=ru&term=%D0%9C%D0%BE%D1%81%D0%BA%D0%B2%D0%B0&types%5B%5D=city&types%5B%5D=airport"
	params := search.Params{
		Term:   "Москва",
		Locale: "ru",
		Types: []string{
			"city",
			"airport",
		},
	}

	got := queryToString(params)
	if expect != got {
		t.Errorf("expected: %s got: %s", expect, got)
	}
}

func TestRequesterRequest(t *testing.T) {
	tt := []struct {
		name      string
		expectErr bool
		handler   func(w http.ResponseWriter, r *http.Request)
		expect    []place.Model
	}{
		{
			name: "ok response",
			expect: []place.Model{
				{
					Slug:     "MOW",
					SubTitle: "Russia",
					Title:    "Moscow",
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, okResponse)
			},
		},
		{
			name: "bad response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectErr: true,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, teardown := newClient(tc.handler)
			defer teardown()
			r := New(client)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			got, err := r.Request(ctx, search.Params{})

			if tc.expectErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(tc.expect, got) {
				t.Errorf("expected: %v got: %v", tc.expect, got)
			}
		})
	}
}

func newClient(handler http.HandlerFunc) (*http.Client, func()) {
	s := httptest.NewTLSServer(handler)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return client, s.Close
}

const (
	okResponse = `
[{"main_airport_name": null, "type": "city", "country_code": "RU", "name": "Moscow", "coordinates": {"lat": 55.755786, "lon": 37.617633}, "country_name": "Russia", "weight": 1006321, "code": "MOW", "index_strings": ["defaultcity", "defaultcity", "maskava", "maskva", "mosca", "moscou", "moscova", "moscovo", "moscow", "mosc\u00fa", "moskau", "moskou", "moskova", "moskow", "moskva", "moskwa", "moszkva", "\u03bc\u03cc\u03c3\u03c7\u03b1", "\u043c\u043e\u0441\u043a\u0432\u0430", "\u043d\u0435\u0440\u0435\u0437\u0438\u043d\u043e\u0432\u0430\u044f", "\u043d\u0435\u0440\u0435\u0437\u0438\u043d\u043e\u0432\u0430\u044f", "\u043d\u0435\u0440\u0435\u0437\u0438\u043d\u043e\u0432\u0441\u043a", "\u043d\u0435\u0440\u0435\u0437\u0438\u043d\u043e\u0432\u0441\u043a", "\u043f\u043e\u043d\u0430\u0435\u0445\u0430\u0432\u0441\u043a", "\u043f\u043e\u043d\u0430\u0435\u0445\u0430\u0432\u0441\u043a", "\u0574\u0578\u057d\u056f\u057e\u0561", "\u05de\u05d5\u05e1\u05e7\u05d1\u05d4", "\u0645\u0633\u06a9\u0648", "\u0645\u0648\u0633\u0643\u0648", "\u092e\u093e\u0938\u094d\u0915\u094b", "\u0e21\u0e2d\u0e2a\u0e42\u0e01", "\u10db\u10dd\u10e1\u10d9\u10dd\u10d5\u10d8", "\u30e2\u30b9\u30af\u30ef", "\u83ab\u65af\u79d1", "\ubaa8\uc2a4\ud06c\ubc14"], "state_code": null, "cases": null, "country_cases": null}]
`
)
