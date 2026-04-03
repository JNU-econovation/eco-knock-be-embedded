# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.26 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=arm64

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY application.yaml /app/application.yaml

ENTRYPOINT ["/app/server"]
