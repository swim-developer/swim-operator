# swim-operator — Knowledge Base

## What This Is

**Kubernetes Operator (Go, Operator SDK) for lifecycle management of all SWIM services on OpenShift.** A user deploys a SWIM service by creating a single Custom Resource — the operator handles everything else (Artemis, databases, certificates, secrets, routes).

**Primary operator** (OLM bundle, CSV — use for OpenShift demos): `swim-openshift-operator/`
**Secondary operator** (parallel Go implementation): `swim-kubernetes-operator/`

## Supported Custom Resources

| CRD | Service Managed |
|-----|-----------------|
| `SwimDnotamConsumerValidator` | DNOTAM Consumer Validator (mock AISP: Artemis + SM API + event generator) |
| `SwimDigitalNotamConsumer` | DNOTAM Consumer (ANSP role: MongoDB + Kafka + XML validation) |
| `SwimDigitalNotamProvider` | DNOTAM Provider (AISP role: PostgreSQL + Kafka + Artemis) |
| `SwimDnotamProviderValidator` | DNOTAM Provider Validator (conformance testing) |

> CRDs renamed April 2026: `SwimDnotamMockServer` → `SwimDnotamConsumerValidator`, `SwimDnotamMockClient` → `SwimDnotamProviderValidator`. Old CRs are not recognized.

## What the Operator Automates

- **cert-manager** integration: mTLS secrets (EACP PKI) created automatically
- **Database provisioning**: MongoDB, PostgreSQL, MariaDB instances
- **Route exposure**: OpenShift Routes + ServiceMonitors
- **Self-healing**: Kubernetes controller reconciliation loop
- **Config injection**: ConfigMaps + Secrets mounted into pods

## Deploy a Service (example)

```bash
# Install operator (OLM)
oc apply -f swim-openshift-operator/bundle/

# Deploy DNOTAM Consumer Validator
oc apply -f swim-openshift-operator/config/samples/SwimDnotamConsumerValidator.yaml

# Deploy DNOTAM Consumer
oc apply -f swim-openshift-operator/config/samples/SwimDigitalNotamConsumer.yaml
```

## Shared Infrastructure (Helm)

Before deploying services, install shared infra:

```bash
# Deploy using the swim-infra Helm chart (included in this repo under deploy/openshift/helm/swim-infra/)
helm install swim-infra deploy/openshift/helm/swim-infra/ -n swim-demo
# or use the Makefile targets if available in your setup:
# make infra-install / make infra-status / make infra-uninstall
```

## CI/CD

Tekton pipeline: `operator-ci` — builds 3 images, OLM reinstall, applies sample CR.

## Build

```bash
make docker-build    # builds operator image — NEVER run this; ask user to run on their machine
make generate        # regenerate CRD manifests after API changes
make manifests       # regenerate RBAC + CRD manifests
```
