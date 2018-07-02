.PHONY: test

test:
	go test -race -v -timeout 5s ./...

fmt:
	gofmt -w -e ./..

check: test fmt
