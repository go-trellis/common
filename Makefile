.PHONY: build unittest

build:
	go build ./...

unittest:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out -o coverage.html
	gocov convert coverage.out | gocov-xml > coverage.xml
