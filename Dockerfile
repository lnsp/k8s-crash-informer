FROM golang:alpine AS builder
LABEL maintainer="Lennart Espe <lennart@espe.tech>"
ARG ARCH

RUN apk update && \
    apk add git build-base && \
    rm -rf /var/cache/apk/* && \
    mkdir -p "/build"

WORKDIR /build
COPY go.mod go.sum /build/
RUN go mod download

COPY . /build/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -a --installsuffix cgo --ldflags="-s" -o informer

FROM alpine
RUN apk add --update ca-certificates
COPY --from=builder /build/informer /bin/informer
ENTRYPOINT ["/bin/informer"]
