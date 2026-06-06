# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Kubernetes Operators (Go, Operator SDK / Kubebuilder) for deploying SWIM (System Wide Information Management) aviation services. A single Custom Resource deploys the full stack: application, database, message broker, certificates, routes/ingress, and observability.

API group: `apps.swim-developer.github.io/v1alpha1`

## Build and Test Commands

```bash
# Root-level (both operators)
make build          # go build both operators
make test           # unit tests for both operators
make lint           # go fmt + go vet on both operators

# Per-operator (from swim-openshift-operator/ or swim-kubernetes-operator/)
make build          # build + manifests + generate + fmt + vet
make test           # unit tests via envtest (excludes e2e)
make manifests      # regenerate CRD + RBAC manifests
make generate       # regenerate DeepCopy methods
make lint           # golangci-lint
make fmt            # go fmt
make vet            # go vet

# Run a single test file
cd swim-openshift-operator && go test ./internal/controller/ -run TestReconcile -v

# E2E tests (requires Kind cluster)
cd swim-kubernetes-operator && make test-e2e

# Common module
cd swim-operator-common && make all   # tidy + fmt + vet + generate
```

## Module Structure

Three Go modules with separate `go.mod` files:

- **`swim-operator-common/`** — Shared library, no `main`. Contains all domain types (`api/v1alpha1/`), resource builders (`pkg/resources/`), reconciler primitives (`pkg/reconciler/`), and controller phase logic (`pkg/controller/`). Both operators import this module.

- **`swim-openshift-operator/`** — OpenShift-specific operator. Uses OLM bundles, OpenShift Routes (`routev1`), Red Hat AMQ Broker/Streams. Has its own CRD type definitions in `api/v1alpha1/` that embed common types.

- **`swim-kubernetes-operator/`** — Vendor-neutral Kubernetes operator. Uses Helm/Kustomize, Kubernetes Ingress, upstream community images. Helm chart at `charts/`.

## Architecture

### Reconciliation Pattern

Each CRD controller follows a **phase-based reconciliation** pattern:

1. **Controller** (`internal/controller/`) — Platform-specific: handles the CR lifecycle, delegates to common phases, and manages platform-specific resources (Routes vs Ingress).
2. **Phases** (`swim-operator-common/pkg/controller/{consumer,provider,cv,pv}/phases*.go`) — Ordered reconciliation steps (RBAC → Secrets → Certificates → Database → Broker → App → HPA → Monitoring). Each phase calls builders and then reconciler primitives.
3. **Builders** (`pkg/controller/*/builders*.go`) — Pure functions that construct desired Kubernetes resources from `BuildParams`.
4. **Reconciler primitives** (`pkg/reconciler/reconciler.go`) — Generic get-or-create-or-update functions for each resource kind (Secret, ConfigMap, Deployment, StatefulSet, Certificate, Ingress, etc.).

### Provider Strategy Pattern

Provider controllers use a `ProviderStrategy` interface (`pkg/controller/provider/strategy.go`) with implementations per service type (`strategy_dnotam.go`, `strategy_ed254.go`). The strategy supplies service-specific config: images, Artemis names, Kafka topics, ConfigMap data, OIDC secrets, PostgreSQL params.

### Shared Infrastructure Cleanup

When multiple CRs share Artemis/Kafka in the same namespace, the operator tracks active CRs (`pkg/reconciler/active_crs.go`) and only deletes shared infra when the last CR is removed (`CleanupSharedInfraIfLast`).

### CRD Kinds

| Controller | Service |
|---|---|
| `SwimDigitalNotamConsumer` | DNOTAM Consumer (MongoDB + Kafka + mTLS) |
| `SwimDigitalNotamProvider` | DNOTAM Provider (PostgreSQL + Artemis + Kafka + OIDC) |
| `SwimDnotamConsumerValidator` | DNOTAM Consumer Validator (MariaDB + Artemis + mTLS) |
| `SwimDnotamProviderValidator` | DNOTAM Provider Validator (MariaDB + mTLS) |
| `SwimEd254Consumer` | ED-254 Consumer |
| `SwimEd254Provider` | ED-254 Provider |
| `SwimEd254ConsumerValidator` | ED-254 Consumer Validator |

## Key Conventions

- After changing API types in `api/v1alpha1/`, always run `make generate && make manifests` in the affected module.
- Tests use controller-runtime's **envtest** (not mocks). The test suite boots a real API server via `setup-envtest`.
- The openshift operator uses `github.com/openshift/api` for Route types; the kubernetes operator uses standard `networkingv1.Ingress`.
- Unstructured objects are used for third-party CRDs (ActiveMQArtemis, KafkaTopic) to avoid importing their full Go types.

## Non-Negotiable Rules

- **Consumer connects to Consumer Validator, NEVER to Provider** — this is an absolute architectural rule enforced in operator reconciliation.
