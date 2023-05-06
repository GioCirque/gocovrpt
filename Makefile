.PHONY: all

BINARY_VERSION=0.0.1
BINARY_NAME=gocovrpt
BINARY_PATH=./.build
SRC_PATH=./src
MAIN_CMD=${SRC_PATH}/cmd/main.go
PKG_SPEC := $(if $(PKG_SPEC),${SRC_PATH}/$(PKG_SPEC)...,${SRC_PATH}/...)
GO=$(shell command -v go)
FOUND_OS := $(if $(FOUND_OS),$(FOUND_OS),$(shell go env FOUND_OS))
FOUND_ARCH := $(if $(FOUND_ARCH),$(FOUND_ARCH),$(shell go env FOUND_ARCH))
EXEC_SUFFIX=
ifeq ($(FOUND_OS),windows)
	EXEC_SUFFIX=.exe
endif
GIT_HASH_SHORT=$(shell git rev-parse --short HEAD)
LDFLAGS_HASH=github.com/giocirque/reportarr/src/globals.Hash=${GIT_HASH_SHORT}
LDFLAGS_VERSION=github.com/giocirque/reportarr/src/globals.Version=${BINARY_VERSION}
LDFLAGS_BRANCH=github.com/giocirque/reportarr/src/globals.Branch=$(shell git rev-parse --abbrev-ref HEAD)
GOOUTPUT=${BINARY_PATH}/${BINARY_NAME}${EXEC_SUFFIX}
GOBUILD=$(GO) build -o ${GOOUTPUT} -ldflags="-X '${LDFLAGS_VERSION}' -X '${LDFLAGS_BRANCH}' -X '${LDFLAGS_HASH}'" ${MAIN_CMD}
GOINSTALL=$(GO) install
GOTEST=$(GO) test -count=1 -coverprofile=.build/coverage.raw
GOCOVER=$(GO) ${BINARY_PATH}/${BINARY_NAME} -html=.build/coverage.raw -o .build/coverage.html
GOCLEAN=$(GO) clean
GOAPIGEN=$(GO) run github.com/ogen-go/ogen/cmd/ogen -loglevel error -allow-remote -clean -convenient-errors=off -debug.noerr -generate-tests -infer-types -no-server -no-webhook-server
ONDONE=echo "✅ Done!\n"
MKEXE=chmod +x
PROOF=file ${GOOUTPUT}
MKBUILD=mkdir -p .build

export CGO_ENABLED=0
export GOARM=7
export GOOS=${FOUND_OS}
export GOARCH=${FOUND_ARCH}

all: clean test build

test:
	@echo "🧪 Testing ${BINARY_NAME}@${PKG_SPEC} ..."
	@$(MKBUILD)
	@$(GOTEST) ${PKG_SPEC}
	@${GOCOVER}
	@$(ONDONE)

clean:
	@echo "🧼 Cleaning ${BINARY_NAME} ..."
	@$(GOCLEAN) ${PKG_SPEC}
	@rm -rf ${BINARY_PATH}/*
	@$(MKBUILD)
	@$(ONDONE)

build: clean
	@echo "🧱 Building ${BINARY_NAME} @ ${GIT_HASH_SHORT} ..."
	@$(GOBUILD)
	@${MKEXE} ${GOOUTPUT}
	@${PROOF}
	@$(ONDONE)

run: build
	@echo "💨 Running ${BINARY_NAME} ..."
	@${BINARY_PATH}/${BINARY_NAME} --config-dir=./.confg --webux-dir=./src/webux/default
	@$(ONDONE)

run-summary: test
	@echo "💨 Running ${BINARY_NAME} ..."
	@${BINARY_PATH}/${BINARY_NAME} summary
	@$(ONDONE)

cli-cmd:
	cobra-cli add ${NAME} --config .cobra.yaml
