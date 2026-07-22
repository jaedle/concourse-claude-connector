FROM golang:1.26 AS build
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY cmd/ cmd/
COPY internal/ internal/
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -mod=vendor -ldflags="-s -w -X main.version=${VERSION}" \
    -o /connector ./cmd/connector

# default target
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /connector /connector
HEALTHCHECK --interval=30s --timeout=5s CMD ["/connector", "health"]
ENTRYPOINT ["/connector"]
