.PHONY: dep
dep:
	go mod download
	go install golang.org/x/tools/cmd/goimports
	go mod tidy

.PHONY: fmt
fmt:
	gofmt -w ./
	goimports -w ./

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o kube-wan-dns-refresh ./cmd/...

test:
	go test -v ./...

docker-build:
	docker build . -t kube-wan-dns-refresh --platform=linux/arm64
