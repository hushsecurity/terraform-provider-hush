PLUGIN_NAME = terraform-provider-hush
OUT_DIR = bin

.PHONY: all
all: build

.PHONY: build
build:
	@go mod tidy
	@mkdir -p $(OUT_DIR)
	@CGO_ENABLED=0 go build -o $(OUT_DIR)/$(PLUGIN_NAME)

.PHONY: clean
clean:
	rm -rf $(OUT_DIR)

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: lint
lint:
	@golangci-lint fmt -d
	@golangci-lint run
	@make lint-examples

.PHONY: format
format:
	@golangci-lint fmt

.PHONY: lint-examples
lint-examples:
	@terraform fmt -check=true -diff=true examples/

.PHONY: generate
generate:
	go generate -v -x

.PHONY: docs
docs:
	@tfplugindocs generate
