test:
	@go test ./...

get:
	@go get -t -v ./...

benchmark:
	@go test -bench=.

.PHONY: test get benchmark
