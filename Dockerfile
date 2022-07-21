FROM golang:1.17-alpine as builder

ARG VERSION

RUN apk add gcc libc-dev

COPY / /schema-source
WORKDIR /schema-source

RUN go build -ldflags="-X 'main.Version=${VERSION}'" -o schema ./cmd/schema/*.go

FROM alpine

COPY --from=builder /schema-source/schema /usr/bin/
CMD ["schema"]