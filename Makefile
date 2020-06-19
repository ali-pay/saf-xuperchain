ifeq ($(OS),Windows_NT)
  PLATFORM="Windows"
else
  ifeq ($(shell uname),Darwin)
    PLATFORM="MacOS"
  else
    PLATFORM="Linux"
  endif
endif

all: build 
export GO111MODULE=on
export GOFLAGS=-mod=vendor
XCHAIN_ROOT := ${PWD}/core
export XCHAIN_ROOT
PATH := ${PWD}/core/xvm/compile/wabt/build:$(PATH)

build:
	PLATFORM=$(PLATFORM) ./build.sh

test:
	go test -coverprofile=coverage.txt -covermode=atomic ./...
	# test wasm sdk
	GOOS=js GOARCH=wasm go build github.com/xuperchain/xuperchain/core/contractsdk/go/driver
	cd core/xvm/spectest && go run main.go core

contractsdk:
	make -C core/contractsdk/cpp build
	make -C core/contractsdk/cpp test

clean:
	rm -rf output
	rm -f xchain-cli
	rm -f xchain
	rm -f dump_chain
	rm -f event_client

.PHONY: all test clean

cli:
	PLATFORM=$(PLATFORM) ./build-cli.sh
export GO111MODULE=on
export GOFLAGS=-mod=vendor
XCHAIN_ROOT := ${PWD}/core
export XCHAIN_ROOT
PATH := ${PWD}/core/xvm/compile/wabt/build:$(PATH)

http:
	PLATFORM=$(PLATFORM) ./build-gateway.sh
export GO111MODULE=on
export GOFLAGS=-mod=vendor
XCHAIN_ROOT := ${PWD}/core
export XCHAIN_ROOT
PATH := ${PWD}/core/xvm/compile/wabt/build:$(PATH)
