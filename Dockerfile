ARG go_version=1.14
ARG alpine_version=3.11

###
FROM golang:${go_version}-alpine${alpine_version} AS builder

# golangci deps
RUN apk add --no-cache git build-base

ARG golangci_version=1.23.7
RUN wget -O- -nv 'https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh' | sh -s "v${golangci_version}"

WORKDIR /go/src/mockable-server

COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

COPY . ./

# test
RUN golangci-lint run ./...
RUN go test -cover ./...

# build
ARG version=""
RUN go build -ldflags="-X main.Version=${version}" -o /go/bin/mockable-server

###
FROM alpine:${alpine_version}

COPY --from=builder /go/bin/mockable-server /mockable-server

STOPSIGNAL SIGINT
USER nobody:nogroup
ENTRYPOINT ["/mockable-server"]
