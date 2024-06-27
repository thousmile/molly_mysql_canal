FROM golang:1.22.4-bookworm AS builder
ENV GO111MODULE=on
ENV CGO_ENABLED=0
WORKDIR /build
ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -o /build/app

FROM debian:stable-slim
WORKDIR /work/
COPY --from=builder --chown=1001:root /build/app /work/app
COPY --from=builder --chown=1001:root /build/config.yaml /work/config.yaml
CMD ["./app"]
