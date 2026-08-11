# Build the plugin binary.
#
# Single Dockerfile for both the FIPS and the non-FIPS image, following the
# convention used by palette/vmo: CRYPTO_LIB selects the crypto backend.
#
#   non-FIPS: docker build .
#   FIPS:     docker build . --build-arg CRYPTO_LIB=fips
#
ARG BUILDER_GOLANG_VERSION=1.26.5
ARG ALPINE_BASE_IMAGE=us-central1-docker.pkg.dev/palette-images-dev/hardened-images/alpine:3.23-dev

FROM us-central1-docker.pkg.dev/palette-images-dev/hardened-images/builder/golang:${BUILDER_GOLANG_VERSION}-alpine AS builder

ARG CRYPTO_LIB

# Make sure we use go modules
WORKDIR /workspace

# Copy the Go Modules manifests
COPY . .

# Install dependencies
RUN go mod download

# Build cmd. CRYPTO_LIB set (e.g. --build-arg CRYPTO_LIB=fips) -> FIPS 140-3
# build with the Go Cryptographic Module v1.0.0 (GOFIPS140, CMVP #4985);
# unset -> regular build against the standard Go crypto library.
RUN if [ "${CRYPTO_LIB}" ]; then \
        CGO_ENABLED=0 GOFIPS140=v1.0.0 go build -a -mod=mod -o /plugin . ; \
    else \
        CGO_ENABLED=0 GO111MODULE=on go build -mod=mod -o /plugin . ; \
    fi

# hardened runtime base
FROM ${ALPINE_BASE_IMAGE}

ARG CRYPTO_LIB
# Enforce FIPS 140-3 mode at runtime for FIPS builds
ENV GODEBUG=${CRYPTO_LIB:+fips140=on}

# Set root path as working directory
WORKDIR /

RUN mkdir -p /plugin

COPY --from=builder /plugin /plugin/plugin
