FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /src

COPY go.work /src/go.work
COPY packages/semantic/go.mod packages/semantic/go.sum /src/packages/semantic/
COPY packages/orchestrator/go.mod packages/orchestrator/go.sum /src/packages/orchestrator/
COPY packages/correctness/go.mod /src/packages/correctness/

RUN go -C /src/packages/semantic mod download
RUN go -C /src/packages/orchestrator mod download

COPY . .

RUN bash -lc 'source /src/scripts/env.sh && assert_no_legacy_artifacts'

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
	go -C /src/packages/orchestrator build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/go-fmt ./cmd/fmt

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
	go build -trimpath -ldflags="-s -w" -o /out/gofmt cmd/gofmt

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOPATH=/tmp/go \
	go install -trimpath -ldflags="-s -w" golang.org/x/tools/cmd/goimports@v0.43.0 && \
	find /tmp/go/bin -name goimports -exec cp {} /out/goimports \;

FROM golang:1.25-alpine AS gosdk

FROM alpine:3.21

RUN apk add --no-cache git

WORKDIR /work

COPY --from=gosdk /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}" \
	GOCACHE="/work/storage/.cache/go-build" \
	GOPATH="/work/storage/.cache/gopath" \
	GOMODCACHE="/work/storage/.cache/gopath/pkg/mod" \
	TURBO_CACHE_DIR="/work/storage/.cache/turbo"

COPY --from=builder /out/go-fmt /usr/local/bin/go-fmt
COPY --from=builder /out/gofmt /usr/local/bin/gofmt
COPY --from=builder /out/goimports /usr/local/bin/goimports

ENTRYPOINT ["/usr/local/bin/go-fmt"]
CMD ["help"]
