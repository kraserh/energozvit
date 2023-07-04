VERSION := $(shell git describe --tags --match 'v*' 2>/dev/null | sed 's/^v//')
ifeq ($(strip ${VERSION}),)
	VERSION := 0.0.0
endif
HASH := $(shell git rev-parse --short HEAD  2>/dev/null)
ifneq ($(strip ${HASH}),)
	HASH := -${HASH}
endif
BUILDTIME := $(shell date '+%Y%m%d%H%M%S')
GOVARS = -X main.Version=${VERSION}-${BUILDTIME}${HASH} 

LDFLAGS = -ldflags '-s -w ${GOVARS}'
BUILDFLAGS = -trimpath -tags libsqlite3

GOBIN = $(GOPATH)/bin
BIN = bin
DEMODB = examples/demo_db.sqlite
DEMODATA = examples/demo_data.sql

export PATH := $(abspath ${BIN}):${PATH}

help:
	@echo 'usage: make [target]'
	@echo
	@echo 'targets:'
	@echo '  build'
	@echo '  run'
	@echo '  test'
	@echo '  vet'
	@echo '  fmt'
	@echo '  install'
	@echo '  uninstall'
	@echo '  clean'
	@echo '  help'


build:
	go build -o ${BIN}/energozvit ${BUILDFLAGS} ${LDFLAGS} \
		./cmd/energozvit
	go build -o ${BIN}/energozvit-tmpl ${BUILDFLAGS} ${LDFLAGS} \
		./cmd/energozvit-tmpl


run: build ${DEMODB}
	energozvit ${DEMODB}


${DEMODB}: ${DEMODATA} internal/storage/storage.go internal/storage/schema.sql
	rm -f "${DEMODB}"
	energozvit ${DEMODB} --create 1970-01
	sqlite3 ${DEMODB} < ${DEMODATA}


test:
	go test -v ${BUILDFLAGS} ${LDFLAGS} ./internal/...


vet:
	go vet ./internal/...
	go vet ./cmd/...


fmt:
	find -name '*.go' -exec go fmt {} \;


install:
	go install ${BUILDFLAGS} ${LDFLAGS} \
		./cmd/energozvit
	go install ${BUILDFLAGS} ${LDFLAGS} \
		./cmd/energozvit-tmpl
	

uninstall:
	rm -f "${GOBIN}/energozvit"
	rm -f "${GOBIN}/energozvit-tmpl"


clean:
	rm -f  ${BIN}/energozvit
	rm -f  ${BIN}/energozvit-tmpl
	rm -f  ${DEMODB}
	rm -fr $(dir ${DEMODB})Output/


.PHONY: help build run test vet fmt
