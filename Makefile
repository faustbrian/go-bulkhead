GOLIB ?= golib

.PHONY: check check-full ci cohesion docs inventory repository-check

check:
	$(GOLIB) check --local --all

check-full:
	$(GOLIB) check --all

ci:
	$(GOLIB) repository check
	$(GOLIB) cohesion check
	$(GOLIB) check --local --all

cohesion:
	$(GOLIB) cohesion check

docs:
	./scripts/check-docs.sh

inventory repository-check:
	$(GOLIB) repository check
