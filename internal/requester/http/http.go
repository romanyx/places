package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/romanyx/places/internal/broker"
	"github.com/romanyx/places/internal/place"
	"github.com/romanyx/places/internal/search"
)

const (
	endpoint = "https://places.aviasales.ru/v2/places.json"
	typeCity = "city"
)

// New initializer for requester.
func New(client *http.Client) *Requester {
	r := Requester{
		client: client,
	}

	return &r
}

// Requester http implementation.
type Requester struct {
	client *http.Client
}

// Request make request to avaisalves.
func (r *Requester) Request(ctx context.Context, p search.Params) ([]place.Model, error) {
	url := endpoint + "?" + queryToString(p)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "build request")
	}

	req = req.WithContext(ctx)
	resp, err := r.client.Do(req)
	if err != nil {
		// Unwrap context errors from client.
		if deadlineError(err) {
			return nil, context.DeadlineExceeded
		}
		if cancelError(err) {
			return nil, context.Canceled
		}

		return nil, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	// Return error if got wrong status code.
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			return nil, broker.ErrBadRequest
		}
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var places []Place
	if err := json.NewDecoder(resp.Body).Decode(&places); err != nil {
		return nil, errors.Wrap(err, "decode body")
	}

	result := make([]place.Model, len(places))
	for i := range places {
		setPlaceFields(&result[i], &places[i])
	}

	return result, nil
}

func deadlineError(err error) bool {
	return strings.Contains(err.Error(), "context deadline exceeded")
}

func cancelError(err error) bool {
	return strings.Contains(err.Error(), "context canceled")
}

func queryToString(p search.Params) string {
	v := url.Values{}
	v.Set("term", p.Term)
	v.Set("locale", p.Locale)
	for _, t := range p.Types {
		v.Add("types[]", t)
	}

	return v.Encode()
}

// Place represents place from aviasales.
type Place struct {
	Type        string `json:"type"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	CountryName string `json:"country_name"`
	CityName    string `json:"city_name"`
}

func setPlaceFields(model *place.Model, place *Place) {
	model.Slug = place.Code
	model.Title = place.Name

	switch place.Type {
	case typeCity:
		model.SubTitle = place.CountryName
	default:
		model.SubTitle = place.CityName
	}
}
