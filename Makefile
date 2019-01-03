.PHONY: integration
integration:
	@go test -v -tags=integration ./integration

test:
	@go test ./...

fmt:
	@gofmt -w -s ./*/*.go

lint:
	@golint ./...
