PLUGIN_NAME         = terraform-provider-hush
OUT_DIR             = bin
MOCK_FIXTURES_S3    = s3://hush-knowledgebase/openapi/mock_api_fixtures.json
MOCK_FIXTURES_LOCAL = testdata/mock_api_fixtures.json

.PHONY: all
all: build

.PHONY: build
build:
	@$(MAKE) mod-tidy
	@$(MAKE) build-bin

# No `go mod tidy`, so lint can depend on the binary without mutating go.mod.
.PHONY: build-bin
build-bin:
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
	@make validate-examples

.PHONY: format
format:
	@golangci-lint fmt

.PHONY: lint-examples
lint-examples:
	@terraform fmt -check=true -diff=true -recursive examples/

.PHONY: validate-examples
validate-examples: build-bin
	@./scripts/validate-examples.sh

.PHONY: generate
generate:
	go generate -v -x

.PHONY: test
test:
	go test -v $$(go list ./... | grep -v /acc_tests)

.PHONY: test-acc
test-acc:
	go test -v ./internal/provider/acc_tests/... -parallel 10 -count=1 -timeout 30m

.PHONY: docs
docs:
	@tfplugindocs generate
	@find docs -name '*.md' -exec sed -i 's/^subcategory: ".*"/subcategory: ""/' {} +
	@sed -i '/^subcategory:/d' docs/index.md

.PHONY: validate-docs
validate-docs:
	@tfplugindocs validate

.PHONY: fetch-mock-fixtures
fetch-mock-fixtures:
	@mkdir -p testdata
	@aws s3 cp $(MOCK_FIXTURES_S3) $(MOCK_FIXTURES_LOCAL)
