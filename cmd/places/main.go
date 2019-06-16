package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/go-redis/redis"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	httpBroker "github.com/romanyx/places/internal/broker/http"
	"github.com/romanyx/places/internal/log"
	httpRequester "github.com/romanyx/places/internal/requester/http"
	"github.com/romanyx/places/internal/search"
	redisRepository "github.com/romanyx/places/internal/storage/redis"
)

const (
	timeout            = 3 * time.Second
	redisCheckInterval = 15 * time.Second
)

func main() {
	var (
		addr        = flag.String("addr", ":8080", "address of http server")
		redisURL    = flag.String("redis", "127.0.0.1:6379", "redis database URL")
		debugAddr   = flag.String("debug", ":1234", "debug server addr")
		healthAddr  = flag.String("health", ":8081", "health check addr")
		metricsAddr = flag.String("metrics", ":8082", "metrics server addr")
		jaegerURL   = flag.String("jaeger", "http://127.0.0.1:14268", "jaeger server url")
		logLevel    = flag.String("log-level", "debug", "log level")
	)
	flag.Parse()
	log.SetLevel(*logLevel)

	// Health checker handler.
	health := healthcheck.NewHandler()
	// Make a channel for errors.
	errChan := make(chan error)

	// Prepare prometheus metrics.
	log.Info("register exporter", nil)
	pex, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatal(errors.Wrap(err, "register exporter"), nil)
	}
	view.RegisterExporter(pex)
	if err := view.Register(
		ochttp.DefaultServerViews...,
	); err != nil {
		log.Fatal(errors.Wrap(err, "failed to register views"), nil)
	}

	// Build and start metrics server.
	mux := http.NewServeMux()
	mux.Handle("/metrics", pex)
	metricsServer := http.Server{
		Addr:    *metricsAddr,
		Handler: mux,
	}
	log.Info("starting metrics server", map[string]interface{}{
		"addr": *metricsAddr,
	})
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil {
			errChan <- errors.Wrap(err, "metrics server")
		}
	}()

	// Register trace exporter.
	log.Info("register jaeger exporter", map[string]interface{}{
		"addr": *jaegerURL,
	})
	jexp, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: fmt.Sprintf("%s/api/traces", *jaegerURL),
		ServiceName:       "places",
	})
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to create jaeger exporter"), nil)
	}
	defer jexp.Flush()
	trace.RegisterExporter(jexp)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.ProbabilitySampler(0.1),
	})

	// Redis connection.
	log.Info("connectng to redis", map[string]interface{}{
		"addr": *redisURL,
	})
	redis := redis.NewClient(&redis.Options{
		Addr: *redisURL,
	})

	// Add redis health check.
	redisPing := healthcheck.Check(func() error {
		if _, err := redis.Ping().Result(); err == nil {
			return errors.Wrap(err, "ping")
		}

		return nil
	})

	health.AddReadinessCheck("redis ready", redisPing)
	health.AddLivenessCheck("redis live", healthcheck.Async(redisPing, redisCheckInterval))

	// Build and start health server.
	healthServer := http.Server{
		Addr:    *healthAddr,
		Handler: health,
	}

	log.Info("starting health server", map[string]interface{}{
		"addr": *healthAddr,
	})
	go func() {
		if err := healthServer.ListenAndServe(); err != nil {
			errChan <- errors.Wrap(err, "health server")
		}
	}()

	client := http.Client{}
	// Start API server.
	server := setupServer(*addr, &client, redis)

	go func() {
		log.Info("startng server", map[string]interface{}{
			"addr": server.Addr,
		})
		if err := server.ListenAndServe(); err != nil {
			errChan <- errors.Wrap(err, "failed to serve grpc")
		}
	}()

	// Start debug server.
	debugServer := setupDebugServer(*debugAddr)
	go func() {
		log.Info("startng debug server", map[string]interface{}{
			"addr": debugServer.Addr,
		})
		if err := debugServer.ListenAndServe(); err != nil {
			errChan <- errors.Wrap(err, "debug server")
		}
	}()
	defer debugServer.Close()

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Fatal(errors.Wrap(err, "critical error"), nil)
	case <-osSignals:
		log.Info("stop by signal", nil)
		if err := server.Close(); err != nil {
			log.Fatal(errors.Wrap(err, "failed to stop server"), nil)
		}
	}
}

func setupServer(addr string, client *http.Client, redis *redis.Client) *http.Server {
	var requester search.Requester
	requester = httpRequester.New(client)
	requester = search.NewRequesterWithTrace(requester)

	var repository search.Repository
	repository = redisRepository.NewRepository(redis)
	repository = search.NewRepositoryWithTrace(repository)

	var searcher httpBroker.Searcher
	searcher = search.NewService(requester, repository, timeout)
	searcher = httpBroker.NewSearcherWithTrace(searcher)
	searcher = httpBroker.NewSearcherWithLog(searcher)

	server := httpBroker.NewServer(addr, searcher)
	return server
}

func setupDebugServer(addr string) *http.Server {
	s := http.Server{
		Addr:    addr,
		Handler: http.DefaultServeMux,
	}
	return &s
}
