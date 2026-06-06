# SWIM Kubernetes Operator - Ansible Playbook

Automated setup of SWIM Kubernetes Operator on Minikube or any Kubernetes cluster.

## Prerequisites

- Ansible installed (`pip install ansible` or `brew install ansible`)
- Minikube installed (will fail with instructions if not found)
- kubectl installed (playbook will auto-install if not found)

## Requirements

This playbook uses **local files** from the cloned repository. Execute from the `ansible/` directory:

## Quick Start

```bash
# Full setup (creates minikube profile "swim" from scratch)
ansible-playbook swim-setup.yml

# Full setup with samples deployed
ansible-playbook swim-setup.yml -e deploy_samples=true

# Restart existing environment (no reinstall)
ansible-playbook swim-setup.yml -e restart=true

# Cleanup (delete minikube profile)
ansible-playbook swim-setup.yml -e cleanup=true

# Custom resources
ansible-playbook swim-setup.yml -e minikube_cpus=6 -e minikube_memory=16384
```

## Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `minikube_profile` | `swim` | Minikube profile name |
| `minikube_cpus` | `4` | CPUs allocated to minikube |
| `minikube_memory` | `8192` | Memory (MB) allocated to minikube |
| `minikube_driver` | `podman` | Minikube driver |
| `cert_manager_version` | `v1.16.0` | cert-manager version |
| `deploy_samples` | `false` | Deploy sample SWIM components |
| `cleanup` | `false` | Delete minikube profile and exit |
| `restart` | `false` | Restart existing environment (no reinstall) |
| `ns_consumervalidator` | `swim-consumervalidator` | Namespace for Consumer Validator |
| `ns_backend` | `swim-backend` | Namespace for Consumer + Provider |
| `ns_providervalidator` | `swim-providervalidator` | Namespace for Provider Validator |

## What Gets Installed

1. **Minikube** profile with ingress addon
2. **cert-manager** with self-signed ClusterIssuer
3. **Strimzi Kafka Operator** (for Consumer)
4. **ArtemisCloud Operator** (for Provider/Consumer Validator)
5. **Keycloak** with realm `swim` pre-configured (namespace: keycloak)
6. **SWIM Kubernetes Operator**
7. (Optional) Sample components: Consumer Validator, Provider, Consumer, Provider Validator

### Keycloak Clients Pre-configured

| Client | Type | Secret | Usage |
|--------|------|--------|-------|
| `amq-broker` | Confidential | `amq-broker-secret` | Artemis OIDC |
| `swim-dnotam-provider` | Confidential | `swim-dnotam-provider-secret` | Provider API |
| `swim-public-client` | Public | - | Provider Validator frontend |

### Keycloak Users and Roles

| Username | Password | AMQ Role |
|----------|----------|----------|
| `admin` | `admin` | `admin-swim-dnotam-v1-amq-role` |
| `user` | `user` | `user-swim-dnotam-v1-amq-role` |
| `marcelo` | `marcelo` | `marcelo-swim-dnotam-v1-amq-role` |
| `john` | `john` | `john-swim-dnotam-v1-amq-role` |

Role pattern: `<username>-swim-dnotam-v1-amq-role`

## Manual Sample Deployment

If you skip `deploy_samples`, deploy manually from the `swim-kubernetes-operator/` directory:

```bash
# Create namespaces
kubectl create namespace swim-consumervalidator
kubectl create namespace swim-backend
kubectl create namespace swim-providervalidator

# Consumer Validator (EUROCONTROL simulator)
kubectl apply -f config/samples/multi-namespace/consumervalidator.yaml

# Provider (DNOTAM publisher)
kubectl apply -f config/samples/multi-namespace/provider.yaml

# Consumer (ANSP client)
kubectl apply -f config/samples/multi-namespace/consumer.yaml

# Provider Validator (Test UI)
kubectl apply -f config/samples/multi-namespace/providervalidator.yaml
```

## Usage on Existing Cluster

If you already have a Kubernetes cluster (not Minikube), you can skip the minikube tasks:

```bash
# Set kubectl context to your cluster first
kubectl config use-context my-cluster

# Run only the operator installation tasks
ansible-playbook swim-setup.yml --start-at-task="Install cert-manager"
```

## Cleanup

**Using the playbook (recommended):**

```bash
ansible-playbook swim-setup.yml -e cleanup=true
```

**Or manually:**

```bash
# Delete minikube profile (removes everything)
minikube delete --profile swim

# Or remove only SWIM resources (from swim-kubernetes-operator/ directory)
kubectl delete namespace swim-consumervalidator swim-backend swim-providervalidator
kubectl delete -f dist/install.yaml
```

## Troubleshooting

### Minikube not starting
```bash
minikube delete --profile swim
ansible-playbook swim-setup.yml
```

### Operators not ready
```bash
kubectl get pods -A
kubectl describe pod <pod-name> -n <namespace>
```

### Check SWIM operator logs
```bash
kubectl logs -l control-plane=controller-manager -n swim-kubernetes-operator-system
```

