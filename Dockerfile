FROM golang:alpine AS builder
MAINTAINER "Lennart Espe <lennart@espe.tech>"

RUN apk update && \
    apk add git build-base && \
    rm -rf /var/cache/apk/* && \
    mkdir -p "$GOPATH/src/github.com/lnsp/k8s-crash-informer"

ADD . "$GOPATH/src/github.com/lnsp/k8s-crash-informer"
RUN cd "$GOPATH/src/github.com/lnsp/k8s-crash-informer" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get -v . && go build -a --installsuffix cgo --ldflags="-s" -o /informer

FROM alpine:3.4
RUN apk add --update ca-certificates
COPY --from=builder /informer /bin/informer
ENTRYPOINT ["/bin/informer"]