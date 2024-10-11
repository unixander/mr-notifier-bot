# Change these variables as necessary.
service_main_package_path = ./cmd/service
servce_binary_name = review-reminder-bot
cli_main_package_path = ./cmd/cli
cli_binary_name = review-reminder-cli
version = 1.0.0

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## tidy: tidy modfiles and format .go files
.PHONY: tidy
tidy:
	go mod tidy -v
	go fmt ./...


## build-service: build the application as the service
.PHONY: build-service
build-service:
	go build -ldflags="-X 'main.Version=$(version)'" -o=./bin/$(servce_binary_name) $(service_main_package_path)

## build-service-linux: build for linux as the service
.PHONY: build-service-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=$(version)'" -o=./bin/$(service_binary_name)-linux $(service_main_package_path)

## run-service: run the  application as the service
.PHONY: run-service
run: build-service
	./bin/$(service_binary_name)

## build-cli: build the application as cli
.PHONY: build-cli
build-cli:
	go build -ldflags="-X 'main.Version=$(version)'" -o=./bin/$(cli_binary_name) $(cli_main_package_path)

## build-cli-linux: build for linux as cli
.PHONY: build-cli-linux
build-cli-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=$(version)'" -o=./bin/$(cli_binary_name)-linux $(cli_main_package_path)

## run-cli: run the  application as cli
.PHONY: run-cli
run-cli: build-cli
	./bin/$(cli_binary_name)

## audit: run quality control checks
.PHONY: audit
audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)" 
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out