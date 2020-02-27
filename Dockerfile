FROM golang:1.14-alpine3.11 AS builder

RUN apk add --no-cache curl git build-base

ARG golangci_version=1.23.6
RUN curl -sfL 'https://install.goreleaser.com/github.com/golangci/golangci-lint.sh' | \
        sh -s -- -b "$(go env GOPATH)/bin" "v${golangci_version}"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

COPY . ./

# test
RUN golangci-lint run ./...
RUN go test -cover ./...

# build
ARG version=""
RUN go build \
    -ldflags="-X main.Version=${version}" \
    -o /go/bin/app

###
FROM alpine:3.11

COPY --from=builder /go/bin/app /mockable-server

STOPSIGNAL SIGINT
USER nobody:nogroup
ENTRYPOINT ["/mockable-server"]
