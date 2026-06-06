CONTROLLER_GEN ?= $(shell which controller-gen 2>/dev/null)

.PHONY: generate
generate:
ifeq (, $(CONTROLLER_GEN))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
	$(eval CONTROLLER_GEN := $(shell go env GOPATH)/bin/controller-gen)
endif
	$(CONTROLLER_GEN) object paths="./api/..."

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: all
all: tidy fmt vet generate
