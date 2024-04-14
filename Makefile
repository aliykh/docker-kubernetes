APP_SERVER_MAIN_PACKAGE ?= app/cmd
APP_SERVER_BINARY ?= app

-include toolchain.mk

.PHONY: app-build
app-build:
	@$(GO) build -o $(APP_SERVER_BINARY) $(APP_SERVER_MAIN_PACKAGE)

app-build-prod:
	@$(GO) build -ldflags="-w -s" -a -installsuffix cgo -o $(APP_SERVER_BINARY) $(APP_SERVER_MAIN_PACKAGE)

.PHONY: run-test
run-test:
	@$(GO) test -v ./...

.PHONY: lint-go
lint-go:
	golangci-lint run