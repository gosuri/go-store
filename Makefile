test:
	@go test -v ./...

get:
	@go get -t -v ./...

benchmark:
	@go test -bench=.

.PHONY: test get benchmark
