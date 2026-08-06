# Build the manager binary
FROM us-docker.pkg.dev/palette-images/build-base-images/golang:1.22-alpine as builder

# Make sure we use go modules
WORKDIR /

# Copy the Go Modules manifests
COPY . .

# Install dependencies
RUN go mod download

# Build cmd
RUN CGO_ENABLED=0 GO111MODULE=on go build -o /plugin main.go

# we use alpine for easier debugging
FROM alpine

# Set root path as working directory
WORKDIR /

RUN mkdir -p /plugin

COPY --from=builder /plugin /plugin/plugin