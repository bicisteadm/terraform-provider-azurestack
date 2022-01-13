TEST?=$$(go list ./... |grep -v 'vendor')
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=azurestack
TESTTIMEOUT=180m

.EXPORT_ALL_VARIABLES:
  TF_SCHEMA_PANIC_ON_ERROR=1

default: build

tools:
	@echo "==> installing required tooling..."
	@sh "$(CURDIR)/scripts/gogetcookie.sh"
	go install github.com/client9/misspell/cmd/misspell@latest
	go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@latest
	go install github.com/bflad/tfproviderdocs@latest
	go install github.com/katbyte/terrafmt@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install mvdan.cc/gofumpt@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH || $$GOPATH)/bin v1.41.1

build: fmtcheck
	go install

fmt:
	@echo "==> Fixing source code with gofmt..."
	# This logic should match the search logic in scripts/gofmtcheck.sh
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

fumpt:
	@echo "==> Fixing source code with Gofumpt..."
	# This logic should match the search logic in scripts/gofmtcheck.sh
	find . -name '*.go' | grep -v vendor | xargs gofumpt -s -w

# Currently required by tf-deploy compile, duplicated by linters
fmtcheck:
	@sh "$(CURDIR)/scripts/gofmtcheck.sh"
	@sh "$(CURDIR)/scripts/timeouts.sh"
	@sh "$(CURDIR)/scripts/check-test-package.sh"

terrafmt:
	@echo "==> Fixing acceptance test terraform blocks code with terrafmt..."
	@find azurestack | egrep "_test.go" | sort | while read f; do terrafmt fmt -f $$f; done
	@echo "==> Fixing website terraform blocks code with terrafmt..."
	@find . | egrep html.markdown | sort | while read f; do terrafmt fmt $$f; done

goimports:
	@echo "==> Fixing imports code with goimports..."
	@find . -name '*.go' | grep -v vendor | grep -v generator-resource-id | while read f; do ./scripts/goimport-file.sh "$$f"; done

lint:
	./scripts/run-lint.sh

depscheck:
	@echo "==> Checking dependencies.."
	@./scripts/track2-check.sh
	@echo "==> Checking source code with go mod tidy..."
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'go mod tidy' command or revert any go.mod/go.sum changes and commit."; exit 1)
	@echo "==> Checking source code with go mod vendor..."
	@go mod vendor
	@git diff --compact-summary --exit-code -- vendor || \
		(echo; echo "Unexpected difference in vendor/ directory. Run 'go mod vendor' command or revert any go.mod/go.sum/vendor changes and commit."; exit 1)

tflint:
	./scripts/run-tflint.sh

whitespace:
	@echo "==> Fixing source code with whitespace linter..."
	golangci-lint run ./... --no-config --disable-all --enable=whitespace --fix

test: fmtcheck
	@TEST=$(TEST) ./scripts/run-gradually-deprecated.sh
	@TEST=$(TEST) ./scripts/run-test.sh

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout $(TESTTIMEOUT) -ldflags="-X=github.com/hashicorp/terraform-provider-azurestack/version.ProviderVersion=acc"

acctests: fmtcheck
	TF_ACC=1 go test -v ./internal/services/$(SERVICE) $(TESTARGS) -timeout $(TESTTIMEOUT) -ldflags="-X=github.com/hashicorp/terraform-provider-azurestack/version.ProviderVersion=acc"

debugacc: fmtcheck
	TF_ACC=1 dlv test $(TEST) --headless --listen=:2345 --api-version=2 -- -test.v $(TESTARGS)

website-lint:
	@echo "==> Checking documentation for .html.markdown extension present"
	@if ! find website/docs -type f -not -name "*.html.markdown" -print -exec false {} +; then \
		echo "ERROR: file extension should be .html.markdown"; \
		exit 1; \
	fi
	@echo "==> Checking documentation spelling..."
	@misspell -error -source=text -i hdinsight,exportfs website/
	@echo "==> Checking documentation for errors..."
	@tfproviderdocs check -provider-name=azurestack -require-resource-subcategory \
		-allowed-resource-subcategories-file website/allowed-subcategories
	@sh -c "'$(CURDIR)/scripts/terrafmt-website.sh'"

website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

teamcity-test:
	@$(MAKE) -C .teamcity tools
	@$(MAKE) -C .teamcity test


pr-check: generate build test lint tflint website-lint

.PHONY: build test testacc vet fmt fmtcheck errcheck pr-check test-compile website website-test validate-examples
