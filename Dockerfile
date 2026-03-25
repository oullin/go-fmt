FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /src

COPY semantic/go.mod semantic/go.sum /src/semantic/
RUN go -C /src/semantic mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
	go -C /src/semantic build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/fmt ./cmd/fmt

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
	go build -trimpath -ldflags="-s -w" -o /out/gofmt cmd/gofmt

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOPATH=/tmp/go \
	go install -trimpath -ldflags="-s -w" golang.org/x/tools/cmd/goimports@v0.43.0 && \
	if [ -f /tmp/go/bin/${TARGETOS}_${TARGETARCH}/goimports ]; then \
		cp /tmp/go/bin/${TARGETOS}_${TARGETARCH}/goimports /out/goimports; \
	elif [ -f /tmp/go/bin/goimports ]; then \
		cp /tmp/go/bin/goimports /out/goimports; \
	else \
		echo "goimports binary not found after install" >&2; exit 1; \
	fi

FROM alpine:3.21

WORKDIR /work

COPY --from=builder /out/fmt /usr/local/bin/fmt
COPY --from=builder /out/gofmt /usr/local/bin/gofmt
COPY --from=builder /out/goimports /usr/local/bin/goimports

ENTRYPOINT ["/usr/local/bin/fmt"]
CMD ["help"]
