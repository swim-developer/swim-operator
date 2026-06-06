# Minikube Testing Guide

This guide describes how to set up a local environment using Minikube to test the `swim-kubernetes-operator`.

## 1. Prepare Minikube

Start Minikube with sufficient resources to run the middleware operators (Kafka, Artemis, etc.):

```bash
minikube start --cpus 4 --memory 8192 --addons ingress
```

Verify the ingress controller is running:

```bash
kubectl get pods -n ingress-nginx
```

## 2. Install Dependencies (Upstream Operators)

The `swim-kubernetes-operator` manages resources that depend on other operators. Install them before starting the SWIM operator.

### a) Cert-Manager
Required for generating mTLS certificates.

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.0/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=available --timeout=120s deployment/cert-manager -n cert-manager
kubectl wait --for=condition=available --timeout=120s deployment/cert-manager-webhook -n cert-manager
```

Create a self-signed ClusterIssuer (required by samples):

```bash
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
EOF
```

### b) Strimzi Kafka Operator
Required for the Consumer component (Kafka-based event streaming).

```bash
kubectl create namespace strimzi-system --dry-run=client -o yaml | kubectl apply -f -
kubectl create -f 'https://strimzi.io/install/latest?namespace=strimzi-system' -n strimzi-system

# Wait for operator
kubectl wait --for=condition=available --timeout=120s deployment/strimzi-cluster-operator -n strimzi-system
```

### c) ArtemisCloud Operator
Required for the Provider and Consumer Validator (AMQP broker).

```bash
# Install CRDs first
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/crds/broker_activemqartemis_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/crds/broker_activemqartemisaddress_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/crds/broker_activemqartemisscaledown_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/crds/broker_activemqartemissecurity_crd.yaml

# Install operator (watches all namespaces)
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/operator.yaml
```

## 3. Install swim-kubernetes-operator

### a) Install CRDs
```bash
make install
```

### b) Run Locally (Development Mode)
You can run the operator directly on your machine pointing to the Minikube context:

```bash
make run
```

*Or, if you prefer to deploy it into the cluster:*
```bash
make image-build IMG=swim-operator:latest
minikube image load swim-operator:latest
make deploy IMG=swim-operator:latest
```

## 4. Test Components (Samples)

Create a namespace for testing:

```bash
kubectl create namespace swim-test
```

Apply the examples configured in `config/samples/`:

```bash
kubectl apply -k config/samples/ -n swim-test
```

## 5. Verification

Verify that the resources were created and pods are coming up:

```bash
# List SWIM custom resources
kubectl get swimdigitalnotamconsumer,swimdigitalnotamprovider,swimdnotamconsumervalidator,swimdnotamprovidervalidator -n swim-test

# Check middleware resources created by the operator
kubectl get kafka,activemqartemis -n swim-test

# Check application resources
kubectl get deployment,statefulset,ingress,certificate -n swim-test

# Check pods status
kubectl get pods -n swim-test
```

## 6. Accessing Services

Since we are using `Ingress`, add the hosts to your `/etc/hosts`:

```bash
# Get Minikube IP
MINIKUBE_IP=$(minikube ip)

# Add entries (requires sudo)
echo "$MINIKUBE_IP swim-provider.local" | sudo tee -a /etc/hosts
echo "$MINIKUBE_IP provider-validator.local" | sudo tee -a /etc/hosts
echo "$MINIKUBE_IP artemis.local" | sudo tee -a /etc/hosts
```

Test the endpoints:

```bash
# Test provider (if deployed)
curl -k https://swim-provider.local/q/health

# Or use kubectl port-forward as alternative
kubectl port-forward svc/swim-dnotam-provider 8080:8080 -n swim-test
```

## 7. Cleanup

Remove all test resources:

```bash
kubectl delete namespace swim-test

# Optionally remove operators
kubectl delete -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/operator.yaml
kubectl delete -f 'https://strimzi.io/install/latest?namespace=strimzi-system' -n strimzi-system
kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.0/cert-manager.yaml
```

## Troubleshooting

### Pods not starting
Check operator logs:
```bash
kubectl logs -l control-plane=controller-manager -n swim-kubernetes-operator-system
```

### Certificate issues
Check cert-manager logs:
```bash
kubectl logs -l app=cert-manager -n cert-manager
```

Verify ClusterIssuer exists:
```bash
kubectl get clusterissuer selfsigned-issuer
```

### Ingress not accessible
Verify ingress controller is running:
```bash
kubectl get pods -n ingress-nginx
```

Check ingress resources:
```bash
kubectl get ingress -n swim-test
kubectl describe ingress -n swim-test
```
