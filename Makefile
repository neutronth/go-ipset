all: test go-ipset-test

test:
	CGO_ENABLED=0 go test -v ./...

go-ipset-test:
	GOOS=linux CGO_ENABLED=0 go build -o testing/bin/go-ipset-test testing/main.go

.PHONY: test go-ipset-test
