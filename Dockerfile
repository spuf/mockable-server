FROM golang:1.13-alpine AS builder

RUN apk add --no-cache curl git build-base

ARG golangci_version=1.23.1
RUN curl -sfL 'https://install.goreleaser.com/github.com/golangci/golangci-lint.sh' | \
        sh -s -- -b "$(go env GOPATH)/bin" "v${golangci_version}"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

# test
RUN golangci-lint run ./...
RUN go test ./...

# build
RUN go build -o /go/bin/app

###
FROM alpine:3.11

COPY --from=builder /go/bin/app /mockable-server

STOPSIGNAL SIGINT
USER nobody:nogroup
ENTRYPOINT ["/mockable-server"]
