package main

import (
	"context"
	"crypto/tls"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-redis/redis"
	"github.com/ory/dockertest"

	"github.com/romanyx/places/internal/docker"
	logPkg "github.com/romanyx/places/internal/log"
)

var (
	redisClient *redis.Client
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func TestMain(m *testing.M) {
	flag.Parse()

	logPkg.SetOutput(ioutil.Discard)

	if testing.Short() {
		os.Exit(m.Run())
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %v", err)
	}

	redisDocker, err := docker.NewRedis(pool)
	if err != nil {
		log.Fatalf("prepare redis with docker: %v", err)
	}
	redisClient = redisDocker.Client

	code := m.Run()

	redisClient.Close()
	if err := pool.Purge(redisDocker.Resource); err != nil {
		log.Fatalf("could not purge redis docker: %v", err)
	}

	os.Exit(code)
}

func prepareServer(h http.Handler) *http.Server {
	s := httptest.NewTLSServer(h)

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	server := setupServer("", &client, redisClient)
	return server
}
