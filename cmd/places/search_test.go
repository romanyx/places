package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/romanyx/places/internal/place"
	"github.com/romanyx/places/internal/search"
	"github.com/romanyx/places/internal/storage/redis"
)

func TestSearch(t *testing.T) {
	t.Run("getPlaces200", getPlaces200)
	t.Run("getPlaces200Cache", getPlaces200Cache)
	t.Run("getPlaces400", getPlaces400)
	t.Run("getPlaces405", getPlaces405)
	t.Run("getPlaces503", getPlaces503)
}

func getPlaces200(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, okResponse)
	})
	server := prepareServer(h)
	r := httptest.NewRequest(http.MethodGet, "/places", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, r)

	t.Log("Given the need to fetch an list of places")
	{
		t.Log("\tWhen fetching list of places")
		{
			if w.Code != http.StatusOK {
				t.Errorf("\t%s\tShould receive a status code of 200 for the response: %v", failed, w.Code)
				return
			}
			t.Logf("\t%s\tShould receive a status code of 200 for the response", success)

			expect := `[{"slug":"MOW","subtitle":"Russia","title":"Moscow"}]`
			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Errorf("\t%s\tShould be able to read body: %v", failed, err)
				return
			}
			t.Logf("\t%s\tShould be able to read body", success)

			got := string(body)

			if !strings.Contains(got, expect) {
				t.Errorf("\t%s\tShould get expected result:\nexpect:\n%s\ngot:\n%s", failed, expect, got)
				return
			}
			t.Logf("\t%s\tShould get expected result", success)
		}
	}
}

func getPlaces200Cache(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := prepareServer(h)
	r := httptest.NewRequest(http.MethodGet, "/places", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, r)

	repo := redis.NewRepository(redisClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo.Cache(ctx, search.Params{}, []place.Model{
		{
			Slug:     "MOW",
			SubTitle: "Russia",
			Title:    "Moscow",
		},
	})
	defer redisClient.FlushDB()

	t.Log("Given the need to fetch an list of places when request failed and have cache")
	{
		t.Log("\tWhen fetching list of places")
		{
			if w.Code != http.StatusOK {
				t.Errorf("\t%s\tShould receive a status code of 200 for the response: %v", failed, w.Code)
				return
			}
			t.Logf("\t%s\tShould receive a status code of 200 for the response", success)

			expect := `[{"slug":"MOW","subtitle":"Russia","title":"Moscow"}]`
			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Errorf("\t%s\tShould be able to read body: %v", failed, err)
				return
			}
			t.Logf("\t%s\tShould be able to read body", success)

			got := string(body)

			if !strings.Contains(got, expect) {
				t.Errorf("\t%s\tShould get expected result:\nexpect:\n%s\ngot:\n%s", failed, expect, got)
				return
			}
			t.Logf("\t%s\tShould get expected result", success)
		}
	}
}

func getPlaces400(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	server := prepareServer(h)
	r := httptest.NewRequest(http.MethodGet, "/places", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, r)

	t.Log("Given the need to to check response when request to aviasales returns bad request status")
	{
		t.Log("\tWhen fetching list of places")
		{
			if w.Code != http.StatusBadRequest {
				t.Errorf("\t%s\tShould receive a status code of 400 for the response: %v", failed, w.Code)
				return
			}
			t.Logf("\t%s\tShould receive a status code of 400 for the response", success)
		}
	}
}

func getPlaces405(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	server := prepareServer(h)
	r := httptest.NewRequest(http.MethodPost, "/places", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, r)

	t.Log("Given the need to to check response when request has wrong method")
	{
		t.Log("\tWhen fetching list of places")
		{
			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("\t%s\tShould receive a status code of 405 for the response: %v", failed, w.Code)
				return
			}
			t.Logf("\t%s\tShould receive a status code of 405 for the response", success)
		}
	}
}

func getPlaces503(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := prepareServer(h)
	r := httptest.NewRequest(http.MethodGet, "/places", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, r)

	t.Log("Given the need to to check response when request to aviasales failed and no cache in storage")
	{
		t.Log("\tWhen fetching list of places")
		{
			if w.Code != http.StatusServiceUnavailable {
				t.Errorf("\t%s\tShould receive a status code of 503 for the response: %v", failed, w.Code)
				return
			}
			t.Logf("\t%s\tShould receive a status code of 503 for the response", success)
		}
	}
}

const (
	okResponse = `
[{"main_airport_name": null, "type": "city", "country_code": "RU", "name": "Moscow", "coordinates": {"lat": 55.755786, "lon": 37.617633}, "country_name": "Russia", "weight": 1006321, "code": "MOW", "index_strings": ["defaultcity", "defaultcity", "maskava", "maskva", "mosca", "moscou", "moscova", "moscovo", "moscow", "mosc\u00fa", "moskau", "moskou", "moskova", "moskow", "moskva", "moskwa", "moszkva", "\u03bc\u03cc\u03c3\u03c7\u03b1", "\u043c\u043e\u0441\u043a\u0432\u0430", "\u043d\u0435\u0440\u0435\u0437\u0438\u043d\u043e\u0432\u0430\u044f", "\u043d\u0435\u0440\u0435\u0437\u0438\u043d\u043e\u0432\u0430\u044f", "\u043d\u0435\u0440\u0435\u0437\u0438\u043d\u043e\u0432\u0441\u043a", "\u043d\u0435\u0440\u0435\u0437\u0438\u043d\u043e\u0432\u0441\u043a", "\u043f\u043e\u043d\u0430\u0435\u0445\u0430\u0432\u0441\u043a", "\u043f\u043e\u043d\u0430\u0435\u0445\u0430\u0432\u0441\u043a", "\u0574\u0578\u057d\u056f\u057e\u0561", "\u05de\u05d5\u05e1\u05e7\u05d1\u05d4", "\u0645\u0633\u06a9\u0648", "\u0645\u0648\u0633\u0643\u0648", "\u092e\u093e\u0938\u094d\u0915\u094b", "\u0e21\u0e2d\u0e2a\u0e42\u0e01", "\u10db\u10dd\u10e1\u10d9\u10dd\u10d5\u10d8", "\u30e2\u30b9\u30af\u30ef", "\u83ab\u65af\u79d1", "\ubaa8\uc2a4\ud06c\ubc14"], "state_code": null, "cases": null, "country_cases": null}]
`
)
