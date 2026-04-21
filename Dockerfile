# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.26 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG GO_TAGS

RUN set -eu; \
    build_tags=""; \
    if [ -n "${GO_TAGS:-}" ]; then build_tags="-tags=${GO_TAGS}"; fi; \
    CGO_ENABLED=0 GOOS="${TARGETOS:-linux}" GOARCH="${TARGETARCH:-amd64}" \
    go build ${build_tags} -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY application.yaml /app/application.yaml

ENTRYPOINT ["/app/server"]
