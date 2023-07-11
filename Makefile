BUILD_DIR = build

export PROJECT_PKG = envoy-xds-server
export VERSION ?=$(shell git describe --tags --exact-match 2>/dev/null || git symbolic-ref -q --short HEAD)
export COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)
export BUILD_DATE ?= $(shell date +%FT%T%z)

# remove debug info from the binary & make it smaller
LDFLAGS += -s -w
# inject build info
LDFLAGS += -X ${PROJECT_PKG}/internal/app/build.Version=${VERSION} -X ${PROJECT_PKG}/internal/app/build.CommitHash=${COMMIT_HASH} -X ${PROJECT_PKG}/internal/app/build.BuildDate=${BUILD_DATE}

test-bs:
	go test -v ./...

.PHONY: devbuild
devbuild:
	go build ${GOARGS} -tags "${GOTAGS}" -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/app ./cmd/server
	
.PHONY: build
build:
	skaffold -v="info" build --file-output=tags.json

.PHONY: deploy
deploy:
	skaffold -v="info" deploy --build-artifacts=tags.json

.PHONY: undeploy
undeploy:
	skaffold delete

.PHONY: config
config:
	kubectl create configmap xds-config --from-file=./config/config.yaml --from-file=./config/auth.yaml

.PHONY:delconfig
delconfig:
	kubectl delete configmap xds-config

