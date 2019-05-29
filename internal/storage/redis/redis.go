package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"

	"github.com/romanyx/places/internal/place"
	"github.com/romanyx/places/internal/search"
	"github.com/romanyx/places/internal/storage"
)

// NewRepository initializer for repository.
func NewRepository(client *redis.Client) *Repository {
	r := Repository{
		client: client,
	}

	return &r
}

// Repository represnets redis storage.
type Repository struct {
	client *redis.Client
}

// Cache caches query in storage.
func (r *Repository) Cache(ctx context.Context, p search.Params, places []place.Model) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(places); err != nil {
		return errors.Wrap(err, "encode gob")
	}

	key := paramsToHex(p)
	if err := r.client.Set(key, buf.Bytes(), 0).Err(); err != nil {
		return errors.Wrap(err, "set key")
	}
	return nil
}

// Retrieve retieves cache from storage
func (r *Repository) Retrieve(ctx context.Context, p search.Params) ([]place.Model, error) {
	key := paramsToHex(p)
	data, err := r.client.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, storage.ErrCacheNotFound
		}

		return nil, errors.Wrap(err, "get key")
	}

	reader := strings.NewReader(data)
	var places []place.Model
	if err := gob.NewDecoder(reader).Decode(&places); err != nil {
		return nil, errors.Wrap(err, "decode gob")
	}

	return places, nil
}

func paramsToHex(p search.Params) string {
	return hex.EncodeToString([]byte(fmt.Sprint(p)))
}
