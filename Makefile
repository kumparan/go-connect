SHELL:=/bin/bash

changelog_args=-o CHANGELOG.md -tag-filter-pattern '^v'

lint:
	golangci-lint run

changelog:
ifdef version
	$(eval changelog_args=--next-tag $(version) $(changelog_args))
endif
	git-chglog $(changelog_args)

check-cognitive-complexity:
	find . -type f -name '*.go' -not -name "*.pb.go" -not -name "mock*.go" -not -name "*_test.go" \
      -exec gocognit -over 15 {} +

test: lint test-only

check-gotest:
ifeq (, $(shell which richgo))
	$(warning "richgo is not installed, falling back to plain go test")
	$(eval TEST_BIN=go test)
else
	$(eval TEST_BIN=richgo test)
endif

ifdef test_run
	$(eval TEST_ARGS := -run $(test_run))
endif
	$(eval test_command=$(TEST_BIN) ./... $(TEST_ARGS) -v --cover)

test-only: check-gotest
	SVC_ENV=test SVC_DISABLE_CACHING=true $(test_command) -timeout 60s