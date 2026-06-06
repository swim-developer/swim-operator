# SWIM Kubernetes Operator - Installation Guide

This guide shows how to install the SWIM Kubernetes Operator on any Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (Minikube, EKS, GKE, AKS, or any standard Kubernetes)
- `kubectl` configured to access your cluster
- Clone this repository:

```bash
git clone https://github.com/swim-developer/rhone.git
cd rhone/deploy/swim-kubernetes-operator
```

## Quick Start

### 1. Install Dependencies

```bash
# Cert-Manager (required)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.0/cert-manager.yaml
kubectl wait --for=condition=available --timeout=120s deployment/cert-manager -n cert-manager

# Create ClusterIssuer
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
EOF

# Strimzi Kafka Operator (required for Consumer)
kubectl create namespace strimzi-system --dry-run=client -o yaml | kubectl apply -f -
kubectl create -f 'https://strimzi.io/install/latest?namespace=strimzi-system' -n strimzi-system

# ArtemisCloud Operator (required for Provider/Consumer Validator)
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/crds/broker_activemqartemis_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/crds/broker_activemqartemisaddress_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/crds/broker_activemqartemisscaledown_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/crds/broker_activemqartemissecurity_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/operator.yaml
```

### 2. Install SWIM Operator

**Option A: Single YAML file (recommended)**

```bash
kubectl apply -f dist/install.yaml
```

**Option B: Using Helm**

```bash
helm install swim-operator charts/swim-kubernetes-operator \
  --namespace swim-operator-system \
  --create-namespace
```

Wait for the operator to be ready:

```bash
kubectl wait --for=condition=available --timeout=120s deployment/swim-kubernetes-operator-controller-manager -n swim-kubernetes-operator-system
```

### 3. Deploy SWIM Components

Components are deployed in separate namespaces for isolation:

| Namespace | Components |
|-----------|------------|
| `swim-consumervalidator` | Consumer Validator + Artemis + MariaDB |
| `swim-backend` | Consumer + Provider + Kafka + MongoDB + PostgreSQL |
| `swim-providervalidator` | Provider Validator (Test UI) |

```bash
# Create namespaces
kubectl create namespace swim-consumervalidator
kubectl create namespace swim-backend
kubectl create namespace swim-providervalidator

# Deploy in order (Consumer Validator first, then backend, then provider validator)
kubectl apply -f config/samples/multi-namespace/consumervalidator.yaml
kubectl apply -f config/samples/multi-namespace/provider.yaml
kubectl apply -f config/samples/multi-namespace/consumer.yaml
kubectl apply -f config/samples/multi-namespace/providervalidator.yaml
```

### 4. Verify Installation

```bash
# Check operator
kubectl get pods -n swim-kubernetes-operator-system

# Check SWIM resources per namespace
kubectl get swimdnotamconsumervalidator -n swim-consumervalidator
kubectl get swimdigitalnotamprovider,swimdigitalnotamconsumer -n swim-backend
kubectl get swimdnotamprovidervalidator -n swim-providervalidator

# Check pods
kubectl get pods -n swim-consumervalidator
kubectl get pods -n swim-backend
kubectl get pods -n swim-providervalidator

# Check ingress
kubectl get ingress -A
```

## Container Images

All images are available on Quay.io:

| Component | Image |
|-----------|-------|
| Operator | `quay.io/masales/swim-kubernetes-operator:latest` |
| Provider | `quay.io/masales/swim-dnotam-provider:latest` |
| Consumer | `quay.io/masales/swim-dnotam-consumer:latest` |
| Consumer validator | `quay.io/masales/swim-dnotam-consumer-validator:latest` |
| Provider validator | `quay.io/masales/swim-dnotam-provider-validator:latest` |

## Uninstall

```bash
# Remove SWIM namespaces (removes all resources)
kubectl delete namespace swim-consumervalidator swim-backend swim-providervalidator

# Remove operator
kubectl delete -f dist/install.yaml

# Remove dependencies (optional)
kubectl delete -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/operator.yaml
kubectl delete -f 'https://strimzi.io/install/latest?namespace=strimzi-system' -n strimzi-system
kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.0/cert-manager.yaml
```

## Next Steps

- Configure Ingress with your domain
- Set up proper TLS certificates (replace self-signed)
- Configure observability (Prometheus/Grafana)
- Review HPA settings for production

## Support

For issues and contributions, visit: https://github.com/swim-developer/rhone

