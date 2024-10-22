############# builder
FROM golang:1.23.0 AS builder

ENV BINARY_PATH=/go/bin
WORKDIR /go/src/github.com/23technologies/gardener-extension-provider-hcloud

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG EFFECTIVE_VERSION

RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

############# base
FROM gcr.io/distroless/static-debian11:nonroot as base

WORKDIR /

############# gardener-extension-provider-hcloud
FROM base AS gardener-extension-provider-hcloud
LABEL org.opencontainers.image.source="https://github.com/23technologies/gardener-extension-provider-hcloud"

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-provider-hcloud /gardener-extension-provider-hcloud
ENTRYPOINT ["/gardener-extension-provider-hcloud"]

############# gardener-extension-admission-hcloud
FROM base AS gardener-extension-admission-hcloud

COPY --from=builder /go/bin/gardener-extension-admission-hcloud /gardener-extension-admission-hcloud
ENTRYPOINT ["/gardener-extension-admission-hcloud"]

