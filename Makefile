RESOURCEDIR=./resource
EXAMPLEDIR=./example
DOCDIR=./doc
SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY=energozvit
VERSION=`git describe --tags --match 'v*' 2>/dev/null | sed 's/^v//'`
BUILDTIME=`date +%FT%T%z`
LDFLAGS=-ldflags "-w -X main.Version=${VERSION} -X main.BuildTime=${BUILDTIME}"

.DEFAULT: $(BINARY)

$(BINARY): $(SOURCES)
	go build ${LDFLAGS} -o ${BINARY} main.go

.PHONY: build
build: $(BINARY)

.PHONY: install
install: $(SOURCES)
	go install ${LDFLAGS}
	
.PHONY: uninstall
uninstall:
	rm  ${GOPATH}/bin/$(BINARY)

.PHONY: run
run:
	if [ ! -f "${EXAMPLEDIR}/myexample.ez.db" ] ; \
	then \
		cp "${EXAMPLEDIR}/example.ez.db" "${EXAMPLEDIR}/myexample.ez.db" ; \
	fi
	go run ${LDFLAGS} main.go ${EXAMPLEDIR}/myexample.ez.db

.PHONY: fmt
fmt: $(SOURCES)
	find -name '*.go' -exec go fmt {} \;

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; \
		then rm ${BINARY} ; fi
	if [ -f "${EXAMPLEDIR}/myexample.ez.db" ] ; \
		then rm "${EXAMPLEDIR}/myexample.ez.db" ; fi
