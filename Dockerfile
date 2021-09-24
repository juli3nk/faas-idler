FROM golang:1.17-alpine3.14 AS builder

RUN apk --update add \
		ca-certificates \
		gcc \
		git \
		musl-dev

COPY go.mod go.sum /go/src/github.com/juli3nk/faas-idler/
WORKDIR /go/src/github.com/juli3nk/faas-idler

ENV GO111MODULE on
RUN go mod download

COPY . .

RUN go build -ldflags "-linkmode external -extldflags -static -s -w" -o /tmp/faas-idler


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /tmp/faas-idler /faas-idler

ENTRYPOINT ["/faas-idler"]
