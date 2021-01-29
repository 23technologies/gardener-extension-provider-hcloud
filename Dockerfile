############# builder
FROM eu.gcr.io/gardener-project/3rd/golang:1.15.5 AS builder

WORKDIR /go/src/github.com/23technologies/gardener-extension-provider-hcloud
COPY . .
RUN make install

############# base
FROM eu.gcr.io/gardener-project/3rd/alpine:3.12.1 AS base

############# gardener-extension-provider-hcloud
FROM base AS gardener-extension-provider-hcloud

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-provider-hcloud /gardener-extension-provider-hcloud
ENTRYPOINT ["/gardener-extension-provider-hcloud"]

############# gardener-extension-validator-hcloud
FROM base AS gardener-extension-validator-hcloud

COPY --from=builder /go/bin/gardener-extension-validator-hcloud /gardener-extension-validator-hcloud
ENTRYPOINT ["/gardener-extension-validator-hcloud"]
