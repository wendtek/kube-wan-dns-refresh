# Build stage container
FROM golang:1.22 as builder
RUN go version
COPY . /build
WORKDIR /build

ENV CGO_ENABLED=0
RUN go mod download && go build -o kube-wan-dns-refresh cmd/main.go

# Application Container
FROM alpine:3.19
WORKDIR /
COPY --from=builder /build/kube-wan-dns-refresh /kube-wan-dns-refresh
ENTRYPOINT ["/kube-wan-dns-refresh"]
