## places

[![CircleCI](https://circleci.com/gh/romanyx/places/tree/master.svg?style=svg)](https://circleci.com/gh/romanyx/places/tree/master)

#### test

```sh
make test
```

#### linter

```sh
make linter
```

#### dev env

* start docker-compose

```sh
make start
```

* make test request

```sh
curl -X GET "http://localhost:8080/places?term=Moscow&locale=en&types%5B%5D=airport&types%5B%5D=city"
```

#### profiling

```sh
go tool pprof http://localhost:1234/debug/pprof/allocs
```

make load

```sh
hey -n 1000 -c 8 -m GET "http://localhost:8080/places?term=Moscow&locale=en&types%5B%5D=airport&types%5B%5D=city"
```

#### traces

visit: http://localhost:16686

#### metrics

visit: http://localhost:9090

#### health checks

1. visit: http://localhost:8081/live
2. visit: http://localhost:8081/ready