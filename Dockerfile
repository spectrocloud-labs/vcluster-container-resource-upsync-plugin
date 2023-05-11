# Build the manager binary
FROM --platform=linux/amd64 golang:1.19 as builder

# Make sure we use go modules
WORKDIR /

# Copy the Go Modules manifests
COPY . .

# Install dependencies
RUN go mod vendor

# Build cmd
RUN CGO_ENABLED=0 GO111MODULE=on go build -mod vendor -o /plugin main.go

# final image
FROM --platform=linux/amd64 gcr.io/distroless/static:nonroot
USER 65532:65532

# Set root path as working directory
WORKDIR /

COPY --from=builder /plugin .

ENTRYPOINT ["/plugin"]
