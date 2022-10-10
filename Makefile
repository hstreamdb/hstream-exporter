PACKAGE := github.com/hstreamdb/hstream-exporter

export GO_BUILD=GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) go build -ldflags '-s -w'

all: exporter

fmt:
	gofmt -s -w -l `find . -name '*.go' -print`

exporter:
	$(GO_BUILD) -o bin/hstream-exporter $(PACKAGE)

.PHONY: fmt, exporter, all