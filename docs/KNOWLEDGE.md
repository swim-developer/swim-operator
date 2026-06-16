# swim-operator — Knowledge Base

## What This Is

**Kubernetes Operator (Go, Operator SDK) for lifecycle management of all SWIM services on OpenShift.** A user deploys a SWIM service by creating a single Custom Resource — the operator handles everything else (Artemis, databases, certificates, secrets, routes).

**Primary operator** (OLM bundle, CSV — use for OpenShift demos): `swim-openshift-operator/`
**Secondary operator** (parallel Go implementation): `swim-kubernetes-operator/`

## Supported Custom Resources

### DNOTAM (Digital NOTAM — CP1 mandatory)

| CRD | Service Managed |
|-----|-----------------|
| `SwimDnotamConsumerValidator` | DNOTAM Consumer Validator (mock AISP: Artemis + SM API + event generator) |
| `SwimDigitalNotamConsumer` | DNOTAM Consumer (ANSP role: MongoDB + Kafka + XML validation) |
| `SwimDigitalNotamProvider` | DNOTAM Provider (AISP role: PostgreSQL + Kafka + Artemis) |
| `SwimDnotamProviderValidator` | DNOTAM Provider Validator (conformance testing) |

> CRDs renamed April 2026: `SwimDnotamMockServer` → `SwimDnotamConsumerValidator`, `SwimDnotamMockClient` → `SwimDnotamProviderValidator`. Old CRs are not recognized.

### ED-254 (Extended AMAN — EUROCAE ED-254)

| CRD | Service Managed |
|-----|-----------------|
| `SwimEd254ConsumerValidator` | ED-254 Consumer Validator (mock AISP: Artemis + SM API + event generator) |
| `SwimEd254Consumer` | ED-254 Consumer (ANSP role: MongoDB + Kafka + XML validation) |
| `SwimEd254Provider` | ED-254 Provider (AISP role: PostgreSQL + Kafka + Artemis) |
| `SwimEd254ProviderValidator` | ED-254 Provider Validator (conformance testing) |

### FF-ICE (Flight and Flow Information for a Collaborative Environment)

| CRD | Service Managed |
|-----|-----------------|
| `SwimFficeConsumerValidator` | FF-ICE Consumer Validator (mock provider: Artemis + SM API + event generator) |
| `SwimFficeConsumer` | FF-ICE Consumer (MongoDB + Kafka + XML validation) |
| `SwimFficeProvider` | FF-ICE Provider (PostgreSQL + Kafka + Artemis) |
| `SwimFficeProviderValidator` | FF-ICE Provider Validator (conformance testing) |

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
oc apply -f swim-openshift-operator/config/samples/apps_v1alpha1_swimdnotamconsumervalidator.yaml

# Deploy DNOTAM Consumer
oc apply -f swim-openshift-operator/config/samples/apps_v1alpha1_swimdigitalnotamconsumer_minimal.yaml

# Deploy ED-254 Consumer Validator
oc apply -f swim-openshift-operator/config/samples/apps_v1alpha1_swimed254consumervalidator.yaml

# Deploy ED-254 Consumer
oc apply -f swim-openshift-operator/config/samples/apps_v1alpha1_swimed254consumer_minimal.yaml

# Deploy FF-ICE Consumer Validator
oc apply -f swim-openshift-operator/config/samples/apps_v1alpha1_swimfficeconsumervalidator.yaml

# Deploy FF-ICE Consumer
oc apply -f swim-openshift-operator/config/samples/apps_v1alpha1_swimfficeconsumer_minimal.yaml
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

## Regenerating CRDs

`make manifests` requires `controller-gen` installed locally. If not in `$(LOCALBIN)`, use:

```bash
./bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
```

After regenerating:
1. Update `config/crd/kustomization.yaml` to add new CRD entries
2. Copy new CRD YAMLs to `charts/swim-kubernetes-operator/crds/` (k8s-operator only)
3. Copy new CRD YAMLs to `bundle/manifests/` and update `swim-operator.clusterserviceversion.yaml` (ocp-operator only)
4. Add new resources to `config/rbac/role.yaml` and individual `*_admin/editor/viewer_role.yaml` files
5. Create sample CRs in `config/samples/`
