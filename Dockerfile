FROM golang:1.21 as builder

COPY . /srv

RUN cd /srv && \
    export GO111MODULE=on CGO_ENABLED=0 GOOS=$GOOS && \
    go build -ldflags '-s -w' -v \
        -o /root/.local/bin/hstream-exporter \
        github.com/hstreamdb/hstream-exporter && \
    rm -rf /srv

# -----------------------------------------------------------------------------

FROM ubuntu:jammy

COPY --from=builder /root/.local/bin/hstream-exporter /usr/local/bin/
