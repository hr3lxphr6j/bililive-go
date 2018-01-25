help:
	@echo "test             run test"
	@echo "lint             run lint"

.PHONY: test
test:
	go test -v -cover

.PHONY: lint
lint:
	gofmt -s -w . _example
	golint .
	golint _example
	go vet
