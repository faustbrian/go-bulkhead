GOLIB ?= golib

.PHONY: source-docs-check check check-full ci cohesion docs inventory repository-check

check: source-docs-check
	$(GOLIB) check --local --all

check-full: source-docs-check
	$(GOLIB) check --all

ci: source-docs-check
	$(GOLIB) repository check
	$(GOLIB) cohesion check
	$(GOLIB) check --local --all

cohesion:
	$(GOLIB) cohesion check

docs:
	./scripts/check-docs.sh

inventory repository-check:
	$(GOLIB) repository check

source-docs-check:
	$(GOLIB) source-docs check --json
