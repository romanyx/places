package docker

import (
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"

	"github.com/ory/dockertest"
	"github.com/pkg/errors"
)

const (
	dockerStartWait = 30 * time.Second
)

// RedisDocker holds connection
// to the db and resource info
// for shutdown.
type RedisDocker struct {
	Client   *redis.Client
	Resource *dockertest.Resource
}

// NewRedis starts lates Redis docker image
// and tries to connect with it.
func NewRedis(pool *dockertest.Pool) (*RedisDocker, error) {
	res, err := pool.Run("redis", "latest", []string{})
	if err != nil {
		return nil, errors.Wrap(err, "start redis")
	}

	purge := func() {
		if err := pool.Purge(res); err != nil {
			log.Printf("failed to purge redis: %v", err)
		}
	}

	errChan := make(chan error)
	done := make(chan struct{})

	var client *redis.Client

	go func() {
		if err := pool.Retry(func() error {
			c := redis.NewClient(&redis.Options{
				Addr: fmt.Sprintf("localhost:%s", res.GetPort("6379/tcp")),
			})
			if _, err := c.Ping().Result(); err != nil {
				return err
			}

			client = c
			return nil
		}); err != nil {
			errChan <- err
		}

		close(done)
	}()

	select {
	case err := <-errChan:
		purge()
		return nil, errors.Wrap(err, "check connection")
	case <-time.After(dockerStartWait):
		purge()
		return nil, errors.New("timeout on checking Redis connection")
	case <-done:
		close(errChan)
	}

	rd := RedisDocker{
		Client:   client,
		Resource: res,
	}

	return &rd, nil
}
