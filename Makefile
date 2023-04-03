SHELL:=/bin/bash

changelog_args=-o CHANGELOG.md -tag-filter-pattern '^v'

lint:
	golangci-lint run

changelog:
ifdef version
	$(eval changelog_args=--next-tag $(version) $(changelog_args))
endif
	git-chglog $(changelog_args)