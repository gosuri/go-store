test:
	@go test -v ./...

updatedeps:
	@go get -t -v ./...

benchmark:
	@go test -bench=.

.PHONY: test benchmark
