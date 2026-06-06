.DEFAULT_GOAL := help

OPENSHIFT_DIR := swim-openshift-operator
KUBERNETES_DIR := swim-kubernetes-operator

.PHONY: help sync pull build test lint

help:
	@echo "Targets:"
	@echo "  build        Build both operators (go build)"
	@echo "  test         Run unit tests for both operators"
	@echo "  lint         Run go fmt + go vet on both operators"
	@echo ""
	@echo "For deploy, bundle, or image targets run make inside each operator directory:"
	@echo "  cd $(OPENSHIFT_DIR) && make help"
	@echo "  cd $(KUBERNETES_DIR) && make help"
	@echo ""
	@echo "  sync          Pull this project from remote (git pull --ff-only)"
	@echo "  pull          Pull this project from remote (git pull --ff-only)"

sync: pull

pull:
	@git pull --ff-only

build:
	cd $(OPENSHIFT_DIR) && $(MAKE) build
	cd $(KUBERNETES_DIR) && $(MAKE) build

test:
	cd $(OPENSHIFT_DIR) && $(MAKE) test
	cd $(KUBERNETES_DIR) && $(MAKE) test

lint:
	cd $(OPENSHIFT_DIR) && $(MAKE) fmt vet
	cd $(KUBERNETES_DIR) && $(MAKE) fmt vet
