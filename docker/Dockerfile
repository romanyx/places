FROM golang:latest as build_env

WORKDIR $GOPATH/src/github.com/romanyx/places
COPY . .

ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go mod download
RUN go build -a -o /go/bin/places ./cmd/places

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM alpine:latest
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build_env /go/bin/places /bin/places

ENTRYPOINT ["places"]
