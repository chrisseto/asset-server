BINARY=asset-server
VERSION=$(shell git describe master 2> /dev/null || echo "$${VERSION:-Unknown}")
BUILD_TIME=`date +%FT%T%z`
LDFLAGS=-ldflags "-X main.Version=${VERSION}"
BUILD_DIR=./build/

build:
	go build ${LDFLAGS} -o "${BUILD_DIR}${BINARY}"

clean:
	git clean -Xdf

run:
	rm -f ./db/server.db
	goose up
	make build
	./build/asset-server

test:
	python scripts/test.py

stress:
	python scripts/gen_bodies.py > scripts/targets.txt
	vegeta attack -targets=scripts/targets.txt -rate 150 -duration 15s > results.bin
	cat results.bin | vegeta report
	rm results.bin

.PHONY: build
