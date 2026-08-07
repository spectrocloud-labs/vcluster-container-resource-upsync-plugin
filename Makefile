# If you update this file, please follow:
# https://suva.sh/posts/well-documented-makefiles/

.DEFAULT_GOAL:=help

VERSION_SUFFIX ?= -dev
PROD_VERSION ?= 2.6.0${VERSION_SUFFIX}
PROD_BUILD_ID ?= $(shell date +%Y%m%d.%H%M)

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH ?= $(shell go env GOPATH)

IMG_REPO ?= us-docker.pkg.dev/palette-images/palette/vcluster-container-resource-upsync-plugin
IMG_TAG ?= v0.0.9
# FIPS and non-FIPS images are built from the same Dockerfile and differ only
# by the CRYPTO_LIB build arg, so they get distinct tags.
FIPS_TAG_SUFFIX ?= -fips
IMG_NAME ?= ${IMG_REPO}:${IMG_TAG}
FIPS_IMG_NAME ?= ${IMG_REPO}:${IMG_TAG}${FIPS_TAG_SUFFIX}

GOLANGCI_VERSION ?= 1.46.2

BIN_DIR ?= ./bin
COVER_DIR=_build/cov
COVER_PKGS=$(shell go list ./... | grep -vE 'tests|api|fake|cmd|hack|config|test|config' | tr "\n" ",")

## Basics

all: build ## Generate all

build: tidy
	CGO_ENABLED=0 GO111MODULE=on go build -o ./bin/container-resource-upsync-plugin main.go

bin-dir:
	test -d $(BIN_DIR) || mkdir $(BIN_DIR)

clean: tidy clean-bin ## Clean up code

clean-bin:
	rm -rf bin/

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

tidy: ## Removes unused dependencies in project
	go mod tidy

version: ## Prints version of current make
	@echo $(PROD_VERSION)

## Static analysis

fmt: ## Run go fmt against code
	go fmt ./...

lint: tidy fmt golangci-lint ## Run golangci-lint against code
	$(GOLANGCI_LINT) run

vet: ## Run go vet against code
	go vet ./...

## Tests

test-unit: ## Run unit tests
	@mkdir -p $(COVER_DIR)
	rm -f $(COVER_DIR)/*
	go test -v -covermode=count -coverprofile=$(COVER_DIR)/unit.out ./...

test: test-unit gocovmerge gocover ## Run unit tests and generate a test report
	$(GOCOVMERGE) $(COVER_DIR)/*.out > $(COVER_DIR)/coverage.out
	go tool cover -func=$(COVER_DIR)/coverage.out -o $(COVER_DIR)/cover.func
	go tool cover -html=$(COVER_DIR)/coverage.out -o $(COVER_DIR)/cover.html
	go tool cover -func $(COVER_DIR)/coverage.out | grep total

## Images

docker: docker-build docker-push ## Tags docker image and also pushes it to container registry

docker-fips: docker-build-fips docker-push-fips ## Tags FIPS docker image and also pushes it to container registry

docker-all: docker docker-fips ## Builds and pushes both the non-FIPS and the FIPS image

docker-build: ## Builds docker image
	docker build . --platform=linux/amd64 -t ${IMG_NAME} -f ./Dockerfile

docker-build-fips: ## Builds FIPS docker image (same Dockerfile, CRYPTO_LIB=fips)
	docker build . --platform=linux/amd64 --build-arg CRYPTO_LIB=fips -t ${FIPS_IMG_NAME} -f ./Dockerfile

docker-push: ## Pushes docker image to container registry
	docker push ${IMG_NAME}

docker-push-fips: ## Pushes FIPS docker image to container registry
	docker push ${FIPS_IMG_NAME}

docker-rmi: ## Remove the local docker image
	docker rmi ${IMG_NAME}

docker-rmi-fips: ## Remove the local FIPS docker image
	docker rmi ${FIPS_IMG_NAME}
