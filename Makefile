.PHONY: mocks test coverage

mocks:
	mockery

test:
	go test ./... -tags=test

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
