# syntax=docker/dockerfile:1
ARG GO_VERSION=1.21.3
ARG GOLANGCI_LINT_VERSION=v1.55

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine as build
WORKDIR /src

RUN GRPC_HEALTH_PROBE_VERSION=v0.4.23 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download -x

ARG APP_VERSION="v0.0.0+unknown"
RUN --mount=type=bind,target=. \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build --ldflags "-s" --ldflags "-X main.version=$APP_VERSION" -o /bin/order ./cmd/

FROM scratch as binary
COPY --from=build /bin/order /bin/

FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION} as lint
WORKDIR /test
RUN --mount=type=bind,target=. \
    golangci-lint run

from scratch
COPY --from=build /bin/order /bin/
COPY --from=build /bin/grpc_health_probe /bin/
ENTRYPOINT [ "/bin/order" ]
