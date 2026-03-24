FROM golang:1.25-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /src

COPY semantic/go.mod semantic/go.sum /src/semantic/
RUN go -C /src/semantic mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
	go -C /src/semantic build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/fmt ./cmd/fmt

FROM golang:1.25-alpine

WORKDIR /work

COPY --from=builder /out/fmt /usr/local/bin/fmt

ENTRYPOINT ["/usr/local/bin/fmt"]
CMD ["help"]
