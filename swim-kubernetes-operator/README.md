# swim-kubernetes-operator

Kubernetes Operator for SWIM (System Wide Information Management) Digital NOTAM services. This operator manages the lifecycle of DNOTAM Provider, Consumer, Consumer Validator, and Provider Validator components on any standard Kubernetes cluster.

## Description

This operator provides a vendor-neutral alternative to `swim-operator`, using standard Kubernetes resources (Ingress instead of OpenShift Routes) and upstream container images. It enables ANSPs to deploy and manage SWIM services with minimal configuration through Custom Resources.

**Key Features:**
- Standard Kubernetes Ingress (no OpenShift dependency)
- Horizontal Pod Autoscaler (HPA) support
- Cert-manager integration for TLS
- Prometheus ServiceMonitor for observability
- Configurable cluster domain
- Minikube-ready with Ansible automation

## Quick Start (Minikube)

This is the fastest way to get a full SWIM environment running locally. The Ansible playbook provisions a Minikube cluster with all prerequisites and deploys sample SWIM components.

### Prerequisites

| Tool | Install |
|---|---|
| Minikube | `brew install minikube` (macOS) |
| Podman | `brew install podman && podman machine init && podman machine start` |
| Ansible | `brew install ansible` (macOS) or `pip install ansible` |
| Go 1.25+ | `brew install go` |
| kubectl | Installed automatically by the playbook if missing |

### Step 1, Build and push the operator image

```bash
cd applications/swim-kubernetes-operator

make image-build image-push IMG=quay.io/masales/swim-kubernetes-operator:latest
```

> If you want to use a local image without pushing, see [Loading images into Minikube](#loading-images-into-minikube).

### Step 2, Generate the install manifest

```bash
make build-installer IMG=quay.io/masales/swim-kubernetes-operator:latest
```

This creates `dist/install.yaml`, a single YAML with all CRDs, RBAC, and the operator Deployment.

### Step 3, Run the Ansible playbook

```bash
cd ansible

# Full setup: Minikube + prerequisites + operator + sample SWIM components
ansible-playbook swim-setup.yml -e deploy_samples=true

# OR: Setup without samples (just infrastructure + operator)
ansible-playbook swim-setup.yml
```

The playbook will:
1. Create a Minikube profile `swim` (8 CPUs, 16 GB RAM, Ingress addon)
2. Install **cert-manager** and configure SWIM PKI (CA + ClusterIssuer)
3. Install **Strimzi Kafka Operator** (all-namespace mode)
4. Install **ArtemisCloud Operator** (all-namespace mode)
5. Install the **SWIM Kubernetes Operator** from `dist/install.yaml`
6. Deploy **Keycloak** with dynamic hostname via `nip.io`
7. (If `deploy_samples=true`) Deploy Consumer Validator → Consumer → Provider → Provider Validator

### Step 4, Verify

```bash
# Check all pods
kubectl get pods -A

# Check SWIM operator
kubectl get pods -n swim-kubernetes-operator-system

# Check SWIM CRs
kubectl get swimdigitalnotamproviders -A
kubectl get swimdigitalnotamconsumers -A
kubectl get swimdnotamconsumervalidators -A
kubectl get swimdnotamprovidervalidators -A
```

### Ansible playbook options

| Command | Effect |
|---|---|
| `ansible-playbook swim-setup.yml` | Setup infra + operator (no samples) |
| `ansible-playbook swim-setup.yml -e deploy_samples=true` | Setup + deploy sample SWIM components |
| `ansible-playbook swim-setup.yml -e install_prometheus=true` | Also install Prometheus Operator |
| `ansible-playbook swim-setup.yml -e restart=true` | Restart existing Minikube profile |
| `ansible-playbook swim-setup.yml -e cleanup=true` | Delete the Minikube profile entirely |

### Loading images into Minikube

If you don't want to push images to a registry:

```bash
# Point your shell to Minikube's container runtime
eval $(minikube -p swim docker-env)

# Build the image (it will be available inside Minikube)
make image-build IMG=quay.io/masales/swim-kubernetes-operator:latest CONTAINER_TOOL=docker

# Then set imagePullPolicy: Never in the deployment
```

---

## Manual Setup (any Kubernetes cluster)

If you're not using Minikube or the Ansible playbook, install the prerequisites manually.

### Prerequisites

Install these operators/components on your cluster before deploying the SWIM operator:

1. **cert-manager**
   ```bash
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.0/cert-manager.yaml
   kubectl wait --for=condition=available --timeout=120s deployment/cert-manager -n cert-manager
   ```

2. **Strimzi Kafka Operator** (for Kafka-based event streaming)
   ```bash
   kubectl create namespace strimzi-system
   kubectl apply -f 'https://strimzi.io/install/latest?namespace=strimzi-system' -n strimzi-system
   ```

3. **ArtemisCloud Operator** (for ActiveMQ Artemis AMQP broker)
   ```bash
   # Install CRDs
   kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/crds/broker_activemqartemis_crd.yaml

   # Install operator (see ArtemisCloud docs for full setup)
   ```

4. **SWIM PKI** (CA and ClusterIssuer for mTLS)
   ```bash
   kubectl apply -f config/pki/swim-pki.yaml
   kubectl wait --for=condition=ready certificate/swim-ca -n cert-manager --timeout=120s
   ```

5. **(Optional) Prometheus Operator** for ServiceMonitor support

### Deploy the operator

Choose one of the three deployment methods:

#### Option A, Kustomize (development)

```bash
make install                                                        # Install CRDs
make deploy IMG=quay.io/masales/swim-kubernetes-operator:latest     # Deploy operator
```

#### Option B, Helm (recommended for production)

```bash
helm install swim-operator ./charts/swim-kubernetes-operator \
  --namespace swim-kubernetes-operator-system \
  --create-namespace \
  --set image.repository=quay.io/masales/swim-kubernetes-operator \
  --set image.tag=latest
```

#### Option C, Plain YAML (no tools required)

```bash
make build-installer IMG=quay.io/masales/swim-kubernetes-operator:latest
kubectl apply -f dist/install.yaml
```

### Deploy a sample CR

```bash
# Single-namespace samples (all in the default namespace)
kubectl apply -f config/samples/apps_v1alpha1_swimdigitalnotamprovider.yaml
kubectl apply -f config/samples/apps_v1alpha1_swimdigitalnotamconsumer.yaml

# Multi-namespace samples (realistic topology)
kubectl apply -f config/samples/multi-namespace/
```

---

## Custom Resources

The operator manages four CRD types:

| CRD | Purpose |
|---|---|
| `SwimDigitalNotamProvider` | Publishes DNOTAM events via AMQP + REST API |
| `SwimDigitalNotamConsumer` | Subscribes to DNOTAM events, persists to MongoDB |
| `SwimDnotamConsumerValidator` | Simulates a SWIM external provider (Artemis + Subscription Manager) |
| `SwimDnotamProviderValidator` | Simulates a SWIM consumer UI for testing |

### Minimal Provider example

```yaml
apiVersion: apps.swim-developer.github.io/v1alpha1
kind: SwimDigitalNotamProvider
metadata:
  name: my-provider
spec:
  certManager:
    issuerName: swim-ca-issuer
    issuerKind: ClusterIssuer
  provider:
    image: quay.io/masales/swim-dnotam-provider:latest
    replicas: 1
    ingress:
      enabled: true
      host: swim-provider.example.com
  postgres:
    database: providerdb
    user: provider
    password: password123
    storageSize: 1Gi
  artemis:
    name: swim-artemis
    size: 1
```

### Minimal Consumer example

```yaml
apiVersion: apps.swim-developer.github.io/v1alpha1
kind: SwimDigitalNotamConsumer
metadata:
  name: my-consumer
spec:
  kafka:
    enabled: true
    bootstrapServers: kafka-kafka-bootstrap:9092
  client:
    image: quay.io/masales/swim-dnotam-consumer:latest
    replicas: 1
    mongo:
      database: swim-dnotam
      storageSize: 1Gi
    config:
      swimServiceBaseURL: http://consumervalidator:8080
      amqpBrokerHost: consumervalidator-artemis-hdls-svc
      amqpBrokerPort: 5672
  certManager:
    issuerName: swim-ca-issuer
    issuerKind: ClusterIssuer
```

### Consumer Validator, Ingress

O Consumer Validator (`SwimDnotamConsumerValidator`, `SwimEd254ConsumerValidator`) cria **três Ingress** automaticamente quando `spec.ingress.enabled: true`:

| Ingress | Hostname padrão | Modo nginx | Finalidade |
|---|---|---|---|
| `{name}-https` | derivado de `host` | `backend-protocol: HTTPS` | UI: TLS sem exigência de cert cliente |
| `{name}-api` | derivado de `host` + sufixo `-api` | `ssl-passthrough` | Subscription Manager API: mTLS direto no Quarkus |
| `{name}-artemis` | derivado de `host` + prefixo `{name}-artemis` | gerenciado pelo Artemis reconciler | Broker AMQP: mTLS |

#### Parâmetros `spec.ingress`

| Campo | Obrigatório | Descrição |
|---|---|---|
| `enabled` | não (default: true) | Habilita ou desabilita todos os Ingress |
| `host` | recomendado | Hostname base da UI. Se omitido, auto-gerado como `{name}-{namespace}.{clusterDomain}` |
| `apiHost` | não | Hostname do endpoint mTLS da Subscription Manager API. Se omitido, derivado de `host` adicionando `-api` ao primeiro segmento |
| `artemisHost` | não | Hostname do broker Artemis AMQP. Se omitido, derivado de `host` substituindo o primeiro segmento pelo nome do Artemis (`{name}-artemis`) |
| `tlsSecretName` | não | Secret TLS do servidor. Se omitido, gerado pelo cert-manager como `{name}-server-tls` |
| `annotations` | não | Anotações extras propagadas para todos os Ingress |

#### Exemplo de derivação automática

```
host: dnotam-consumer-validator.127.0.0.1.nip.io
  → UI   : dnotam-consumer-validator.127.0.0.1.nip.io       (backend-protocol: HTTPS)
  → API  : dnotam-consumer-validator-api.127.0.0.1.nip.io   (ssl-passthrough, mTLS)
  → Artms: dnotam-consumer-validator-artemis.127.0.0.1.nip.io
```

#### Pré-requisito: ssl-passthrough no nginx

O Ingress `{name}-api` usa `ssl-passthrough`, que requer o flag `--enable-ssl-passthrough` no controller. No Minikube, habilite com:

```bash
kubectl patch deployment ingress-nginx-controller -n ingress-nginx \
  --type='json' \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--enable-ssl-passthrough"}]'
```

#### Exemplo mínimo (Minikube)

```yaml
apiVersion: apps.swim-developer.github.io/v1alpha1
kind: SwimDnotamConsumerValidator
metadata:
  name: dnotam-consumer-validator
  namespace: default
spec:
  certManager:
    enabled: true
    issuerName: swim-ca-issuer
    issuerKind: ClusterIssuer
  appConfig:
    amqp:
      username: admin
      password: admin
      port: 5672
  ingress:
    enabled: true
    host: dnotam-consumer-validator.127.0.0.1.nip.io
```

---

## Development

### Build from source

```bash
# Generate CRDs and Go code
make generate manifests

# Build binary
make build

# Run locally (against current kubeconfig)
make run
```

### Running tests

```bash
# Unit tests (envtest, no cluster required)
make test

# E2E tests (requires Kind)
make test-e2e
```

### Building OCI images

```bash
# Build (podman by default)
make image-build IMG=quay.io/masales/swim-kubernetes-operator:latest

# Push
make image-push IMG=quay.io/masales/swim-kubernetes-operator:latest

# Multi-arch (linux/amd64 + linux/arm64)
make image-buildx IMG=quay.io/masales/swim-kubernetes-operator:latest

# Using Docker instead of Podman
make image-build CONTAINER_TOOL=docker IMG=quay.io/masales/swim-kubernetes-operator:latest
```

### Useful commands

```bash
make help           # List all make targets
make lint           # Run golangci-lint
make lint-fix       # Auto-fix lint issues
make manifests      # Regenerate CRDs and RBAC
make generate       # Regenerate DeepCopy methods
```

---

## Cleanup

### Remove samples and operator (Kustomize)

```bash
make cleanup-full
```

### Remove Minikube environment

```bash
cd ansible
ansible-playbook swim-setup.yml -e cleanup=true
```

### Uninstall Helm release

```bash
helm uninstall swim-operator -n swim-kubernetes-operator-system
kubectl delete -f config/crd/bases/  # Remove CRDs
```

---

## Project Structure

```
├── ansible/                # Ansible playbook for Minikube setup
├── api/v1alpha1/           # CRD types and common specs
├── charts/                 # Helm chart for deployment
├── cmd/                    # Main entrypoint
├── config/
│   ├── crd/bases/          # Generated CRD YAMLs
│   ├── keycloak/           # Keycloak deployment manifest
│   ├── manager/            # Operator Deployment (Kustomize)
│   ├── manifests/          # OLM bundle manifests
│   ├── pki/                # SWIM PKI (CA + ClusterIssuer)
│   ├── rbac/               # RBAC roles (Kustomize)
│   ├── samples/            # Example CRs (single and multi-namespace)
│   └── scorecard/          # OLM scorecard tests
├── dist/                   # Generated install.yaml (make build-installer)
├── internal/controller/    # Reconciliation logic and resource builders
└── test/                   # E2E and integration tests
```

## Troubleshooting

### Operator pod is CrashLoopBackOff

Check logs:
```bash
kubectl logs -n swim-kubernetes-operator-system -l control-plane=controller-manager -f
```

Common causes:
- cert-manager not installed (CRD registration fails)
- Missing RBAC for watched resources
- Image pull errors (check `imagePullPolicy` and registry access)

### Artemis broker not starting

```bash
# Check ArtemisCloud operator logs
kubectl logs -n activemq-artemis-operator -l name=activemq-artemis-controller-manager -f

# Check the ActiveMQArtemis CR status
kubectl get activemqartemis -A -o yaml
```

### Certificate not ready

```bash
# Check cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager -f

# Check certificate status
kubectl get certificates -A
kubectl describe certificate <name> -n <namespace>
```

### Kafka cluster not starting

```bash
# Check Strimzi operator logs
kubectl logs -n strimzi-system -l name=strimzi-cluster-operator -f

# Check Kafka CR status
kubectl get kafka -A -o yaml
```

---

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
