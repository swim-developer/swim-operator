# swim-operator, Kubernetes Portability Migration Guide

> **Purpose**: This document is the complete technical brief for an AI agent tasked with migrating
> `swim-operator` from OpenShift-only to fully portable Kubernetes (vanilla / Minikube / EKS / GKE / AKS).
>
> **Date of analysis**: April 2026  
> **Analyst**: Automated code review of the full project tree  
> **Scope**: Everything under `applications/swim-operator/`

---

## 1. Executive Summary

The operator binary **compiles and can run** on any Kubernetes cluster. However, **reconciliation
of Providers, Validators, and Artemis broker exposure relies on OpenShift-specific APIs** that do
not exist on vanilla Kubernetes. On Minikube or any plain k8s cluster without OpenShift Route CRDs:

| CRD | Reconciliation result |
|-----|-----------------------|
| `SwimDigitalNotamProvider` | ❌ Fails: creates `Route` objects |
| `SwimDnotamProviderValidator` | ❌ Fails: creates `Route` + domain detection via Route |
| `SwimDnotamConsumerValidator` | ❌ Fails: creates `Route` |
| `SwimEd254Provider` | ❌ Fails: creates `Route` |
| `SwimEd254ConsumerValidator` | ❌ Fails: creates `Route` |
| `SwimDigitalNotamConsumer` | ✅ Works (no Route creation in its reconciler) |
| `SwimEd254Consumer` | ✅ Works (no Route creation in its reconciler) |

Artemis (`ActiveMQArtemis`) uses `exposeMode: "route"` in all configurations, this is the
AMQ Broker Operator / Red Hat mode that instructs Artemis to create OpenShift Routes automatically.
On upstream `activemq-artemis-operator` this mode either does not exist or behaves differently.

---

## 2. Full Inventory of OpenShift Dependencies

### 2.1 Go module dependency

**File**: `go.mod`, line 8

```
github.com/openshift/api v3.9.0+incompatible
```

This is the **only direct OpenShift Go dependency**. It provides the typed structs for
`routev1.Route`, `routev1.RouteSpec`, `routev1.TLSConfig`, `routev1.TLSTerminationEdge`,
`routev1.TLSTerminationPassthrough`, `routev1.InsecureEdgeTerminationPolicyRedirect`, etc.

There is **no** `github.com/openshift/client-go` in `go.mod`.

**Migration action**: Replace all `routev1.*` types with `networkingv1.Ingress` equivalents,
then remove `github.com/openshift/api` from `go.mod` and run `go mod tidy`.

---

### 2.2 Controller files that import `routev1`

All of the following files contain `import routev1 "github.com/openshift/api/route/v1"` and
must be refactored:

| File | Functions that build/reconcile Route objects |
|------|----------------------------------------------|
| `internal/controller/reconcilers.go` | `reconcileRoute()`: generic shared reconciler |
| `internal/controller/consumervalidator_resources.go` | `ConsumerValidatorRoute()`: Edge and Passthrough variants |
| `internal/controller/providervalidator_resources.go` | `ProviderValidatorRoute()`: Edge termination |
| `internal/controller/provider_app_resources.go` | `ProviderAppRouteEdge()`, `ProviderAppRoutePassthrough()` |
| `internal/controller/ed254_provider_app_resources.go` | `Ed254ProviderAppRouteEdge()`, `Ed254ProviderAppRoutePassthrough()` |
| `internal/controller/ed254_provider_reconcilers.go` | `reconcileRoute()`: ED-254 provider variant |
| `internal/controller/ed254consumervalidator_app_resources.go` | `Ed254CVRoute()` |
| `internal/controller/ed254consumervalidator_reconcilers.go` | calls `reconcileRoute()` |
| `internal/controller/swimdigitalnotamprovider_controller.go` | `reconcileRoute()` call + `Owns(&routev1.Route{})` |
| `internal/controller/swimdnotamconsumervalidator_controller.go` | `reconcileRoute()` call + `Owns(&routev1.Route{})` |
| `internal/controller/swimdnotamprovidervalidator_controller.go` | `reconcileRoute()` call + `Owns(&routev1.Route{})` |
| `internal/controller/swimed254provider_controller.go` | `reconcileRoute()` call + `Owns(&routev1.Route{})` |
| `internal/controller/swimed254consumervalidator_controller.go` | `reconcileRoute()` call + `Owns(&routev1.Route{})` |

---

### 2.3 `cluster_utils.go`, OpenShift-specific domain detection

**File**: `internal/controller/cluster_utils.go`

This is the **most architecturally invasive** dependency. The cluster domain is detected via:

**Path 1, `getClusterDomainFromIngressController()`** (lines ~43–64):
```go
ingressController.SetAPIVersion("operator.openshift.io/v1")
ingressController.SetKind("IngressController")
// Reads: openshift-ingress-operator / default / status.domain
```
This **only works on OpenShift** (OCP/OKD). On vanilla k8s, `operator.openshift.io` API group
does not exist → returns error.

**Path 2, `getClusterDomainFromRoute()`** (lines ~67–120):
```go
"apiVersion": "route.openshift.io/v1",
"kind":       "Route",
// Creates a temporary Route, reads its status.ingress[0].host to extract domain
```
This **only works on OpenShift**. On vanilla k8s, `route.openshift.io` API group does not exist
→ returns error.

**Fallback** (line ~123):
```go
cachedClusterDomain = "apps.cluster.local"
```
This fallback is used on vanilla k8s but is almost always **wrong** for cert-manager,
DNS routing, and TLS hostname generation.

**`getOrDetectClusterDomain()`**, called by all controllers. When `specDomain != ""` (set in the
CR), it bypasses detection entirely. **This is the mechanism to support vanilla k8s today**, if
the user sets `global.clusterDomain` in their CR, detection is skipped.

**Migration action**: Replace `getClusterDomainFromIngressController` and `getClusterDomainFromRoute`
with a new function `getClusterDomainFromIngressClass()` that reads `IngressClass` objects
(standard k8s API `networking.k8s.io/v1`). The fallback should also try reading an existing
`Ingress` object's `.status.loadBalancer.ingress[0].hostname`.

---

### 2.4 Artemis `exposeMode: "route"`, 4 occurrences

All four Artemis resource builders hardcode `"exposeMode": "route"`:

| File | Line |
|------|------|
| `internal/controller/artemis_resources.go` | ~133 |
| `internal/controller/provider_artemis_resources.go` | ~256 |
| `internal/controller/ed254_provider_artemis_resources.go` | ~266 |
| `internal/controller/ed254consumervalidator_artemis.go` | ~120 |

Example from `artemis_resources.go`:
```go
acceptors := []interface{}{
    map[string]interface{}{
        "name":       "amqps",
        "expose":     true,
        "exposeMode": "route",        // ← OpenShift-specific
        "ingressHost": ingressHost,
    },
}
```

The `exposeMode` field is a property of the `ActiveMQArtemis` CRD (AMQ Broker Operator).
Valid values depend on the operator version:
- OpenShift / Red Hat AMQ: `"route"` (uses OpenShift Routes)
- Vanilla Kubernetes upstream `activemq-artemis-operator`: `"ingress"` (uses k8s Ingress)

**Migration action**: Replace hardcoded `"route"` with a variable resolved from cluster type or
from the CR spec. Add an `artemisExposeMode` field (or derive from `clusterType` field) that
defaults to `"ingress"` on vanilla and `"route"` on OCP.

---

### 2.5 HAProxy annotation

**File**: `internal/controller/providervalidator_resources.go`, line ~498

```go
Annotations: map[string]string{
    "haproxy.router.openshift.io/timeout": "300s",
},
```

This annotation is **silently ignored** outside of OpenShift's HAProxy router. It is a **soft**
dependency, nothing will break, but the timeout will not be respected.

**Migration action**: Either remove this annotation entirely or make it conditional on OpenShift
mode. The equivalent for vanilla nginx-ingress is: `nginx.ingress.kubernetes.io/proxy-read-timeout: "300"`.

---

### 2.6 RBAC, OpenShift API groups

**File**: `config/rbac/role.yaml`, lines 156–194

```yaml
- apiGroups:
  - operator.openshift.io
  resources:
  - ingresscontrollers
  verbs:
  - get
  - list

- apiGroups:
  - route.openshift.io
  resources:
  - routes
  - routes/custom-host
  verbs:
  - get;list;watch;create;update;patch;delete
```

On vanilla Kubernetes, these rules cause **no error** at apply time (RBAC just permits
non-existent resources), but they are misleading and must be updated.

**Migration action**: Replace `route.openshift.io` rules with `networking.k8s.io` / `ingresses`
rules. Replace `operator.openshift.io/ingresscontrollers` with `networking.k8s.io/ingressclasses`.

Also update the `+kubebuilder:rbac` annotations in the following controller files (these
generate `role.yaml`):
- `swimdigitalnotamprovider_controller.go`, lines 46–48
- `swimdigitalnotamconsumer_controller.go`, lines 48–49
- `swimdnotamconsumervalidator_controller.go`, lines 42–46
- `swimdnotamprovidervalidator_controller.go`, lines 38–40
- `swimed254provider_controller.go`, lines 44–46
- `swimed254consumer_controller.go`, lines 48–49

After editing the annotations, regenerate with: `make generate manifests`

---

### 2.7 OLM Bundle, `openshift-marketplace`

**File**: `install-swim-catalog.yaml`, line ~5

```yaml
namespace: openshift-marketplace
```

**File**: `Makefile`, multiple targets (~392–416, ~469–527) use:
- `oc` CLI commands
- `openshift-marketplace` namespace
- OLM CatalogSource / OperatorHub flow

On vanilla Kubernetes, OLM can be installed separately (https://operatorhub.io/), but the
`openshift-marketplace` namespace does not exist.

**Migration action**: Add a plain `kubectl apply -f` installation path that does not require OLM.
The existing `make deploy` (which uses `kustomize`) is already vanilla-compatible.

---

### 2.8 Dockerfile, UBI base image

**File**: `Dockerfile`, line ~17

```dockerfile
FROM registry.access.redhat.com/ubi9-micro
```

UBI (Universal Base Image) is publicly available at `registry.access.redhat.com` **without
a Red Hat subscription**. Pull is possible on any environment.

**Verdict**: **No action required**. UBI is portable.

---

### 2.9 External Operators Required (portable, not OpenShift-specific)

The following CRDs must be installed on **any** target cluster, including OpenShift:

| Operator | CRD groups | Install on vanilla k8s |
|----------|-----------|------------------------|
| **cert-manager** | `cert-manager.io/v1` | `kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml` |
| **Strimzi** (Kafka) | `kafka.strimzi.io` | `kubectl apply -f https://strimzi.io/install/latest?namespace=<ns>` |
| **StreamsHub Console** | `console.streamshub.github.com` | Helm or operator manifests from StreamsHub |
| **AMQ Broker / ActiveMQ Artemis Operator** | `broker.amq.io` | On vanilla: use upstream `activemq-artemis-operator` from https://github.com/artemiscloud/activemq-artemis-operator |
| **prometheus-operator** | `monitoring.coreos.com` | `kube-prometheus-stack` Helm chart or standalone |

**Note on Artemis**: Red Hat's AMQ Broker Operator (`registry.redhat.io/amq7/...`) requires
a Red Hat pull secret. On vanilla k8s, use the upstream `artemiscloud/activemq-artemis-operator`
which uses the same `broker.amq.io` CRD group but different image references. The `exposeMode`
values also differ (see 2.4).

---

## 3. API Types, No Go-level OpenShift Imports

All files under `api/v1alpha1/` use **only standard Kubernetes types**:
- `k8s.io/api/core/v1` (resources, probes, env vars)
- `k8s.io/apimachinery/pkg/apis/meta/v1` (ObjectMeta, Condition)

The OpenShift coupling is **entirely in the controller layer**, not in the CRD types themselves.
This means CRD schemas are fully portable, no changes to `api/` files are needed.

---

## 4. What Works Today Without Any Changes

The following CRs reconcile successfully on vanilla Kubernetes (provided external operators
are installed):

### `SwimDigitalNotamConsumer`
- Creates: ServiceAccount, Role, RoleBinding, Secrets (MongoDB, Keystore, Providers),
  PVC (MongoDB), Deployments (MongoDB + Client), Services, Certificate (cert-manager),
  HPA, ServiceMonitor, Kafka, KafkaTopic, Console (StreamsHub)
- Does **not** create any Route
- Uses `detectClusterDomain()` only for Kafka Console hostname, if `global.clusterDomain`
  is set in the CR, detection is bypassed entirely
- **Status**: ✅ Fully portable when `global.clusterDomain` is set

### `SwimEd254Consumer`
- Same pattern as `SwimDigitalNotamConsumer`
- **Status**: ✅ Fully portable when `global.clusterDomain` is set

---

## 5. Detailed Migration Plan, Step by Step

### Step 1, Add `IngressClass`-based domain detection to `cluster_utils.go`

Replace `getClusterDomainFromIngressController` and `getClusterDomainFromRoute` with:

```go
// getClusterDomainFromIngressClass reads the default IngressClass and returns
// the domain associated with it, if annotated.
func getClusterDomainFromIngressClass(ctx context.Context, c client.Client) (string, error) {
    ingressClassList := &networkingv1.IngressClassList{}
    if err := c.List(ctx, ingressClassList); err != nil {
        return "", err
    }
    for _, ic := range ingressClassList.Items {
        if ic.Annotations["ingressclass.kubernetes.io/is-default-class"] == "true" {
            // Some controllers annotate the class with the domain
            if domain, ok := ic.Annotations["swim-operator/cluster-domain"]; ok {
                return domain, nil
            }
        }
    }
    return "", fmt.Errorf("no default IngressClass with domain annotation found")
}
```

Also add an attempt to read domain from any existing `Ingress` in the namespace:

```go
func getClusterDomainFromIngress(ctx context.Context, c client.Client, namespace string) (string, error) {
    ingressList := &networkingv1.IngressList{}
    if err := c.List(ctx, ingressList, client.InNamespace(namespace)); err != nil {
        return "", err
    }
    for _, ing := range ingressList.Items {
        for _, rule := range ing.Spec.Rules {
            if rule.Host != "" {
                parts := strings.SplitN(rule.Host, ".", 2)
                if len(parts) == 2 {
                    return parts[1], nil
                }
            }
        }
    }
    return "", fmt.Errorf("no Ingress with host rule found in namespace %s", namespace)
}
```

Update `detectClusterDomain()` to call these new functions and keep `"apps.cluster.local"` as
last resort.

---

### Step 2, Replace `routev1.Route` with `networkingv1.Ingress`

#### 2a, Generic reconciler in `reconcilers.go`

Change the signature of `reconcileRoute`:
```go
// Before
func (r *SwimDnotamConsumerValidatorReconciler) reconcileRoute(
    ctx context.Context, owner *appsv1alpha1.SwimDnotamConsumerValidator,
    desired *routev1.Route) error

// After
func (r *SwimDnotamConsumerValidatorReconciler) reconcileIngress(
    ctx context.Context, owner client.Object,
    desired *networkingv1.Ingress) error
```

The body logic (get → compare → create/update) is nearly identical; only the typed object changes.

#### 2b, Resource builder functions

For each Route builder, create an equivalent Ingress builder. The TLS termination mapping:

| Route TLS Mode | Ingress equivalent |
|---|---|
| `TLSTerminationEdge` | TLS termination at ingress controller; `tls:` block in Ingress spec |
| `TLSTerminationPassthrough` | `nginx.ingress.kubernetes.io/ssl-passthrough: "true"` annotation |

Example replacement for `ConsumerValidatorRoute` → `ConsumerValidatorIngress`:

```go
func (r *SwimDnotamConsumerValidatorReconciler) ConsumerValidatorIngress(
    cr *appsv1alpha1.SwimDnotamConsumerValidator, host, portName, secretName string,
    passthrough bool) *networkingv1.Ingress {

    pathType := networkingv1.PathTypePrefix
    annotations := map[string]string{}
    if passthrough {
        annotations["nginx.ingress.kubernetes.io/ssl-passthrough"] = "true"
    }

    ing := &networkingv1.Ingress{
        ObjectMeta: metav1.ObjectMeta{
            Name:        cr.Name,
            Namespace:   cr.Namespace,
            Labels:      standardLabels(cr.Name, "ingress", cr.Name),
            Annotations: annotations,
        },
        Spec: networkingv1.IngressSpec{
            Rules: []networkingv1.IngressRule{{
                Host: host,
                IngressRuleValue: networkingv1.IngressRuleValue{
                    HTTP: &networkingv1.HTTPIngressRuleValue{
                        Paths: []networkingv1.HTTPIngressPath{{
                            Path:     "/",
                            PathType: &pathType,
                            Backend: networkingv1.IngressBackend{
                                Service: &networkingv1.IngressServiceBackend{
                                    Name: cr.Name,
                                    Port: networkingv1.ServiceBackendPort{Name: portName},
                                },
                            },
                        }},
                    },
                },
            }},
        },
    }
    if !passthrough && secretName != "" {
        ing.Spec.TLS = []networkingv1.IngressTLS{{
            Hosts:      []string{host},
            SecretName: secretName,
        }}
    }
    return ing
}
```

Apply the same pattern to:
- `ProviderValidatorRoute` → `ProviderValidatorIngress`
- `ProviderAppRouteEdge` / `ProviderAppRoutePassthrough` → `ProviderAppIngressEdge` / `ProviderAppIngressPassthrough`
- `Ed254ProviderAppRouteEdge` / `Ed254ProviderAppRoutePassthrough` → equivalents
- `Ed254CVRoute` → `Ed254CVIngress`

#### 2c, Update `Owns()` in SetupWithManager

Replace in all 5 affected controllers:
```go
// Before
Owns(&routev1.Route{}).
// After
Owns(&networkingv1.Ingress{}).
```

---

### Step 3, Replace `exposeMode: "route"` with configurable mode

#### 3a, Add field to shared spec or detect automatically

Option A (recommended): Add a `global.artemisExposeMode` field to `GlobalSpec` in
`api/v1alpha1/swimdigitalnotamconsumer_types.go` (or a shared types file):

```go
type GlobalSpec struct {
    ClusterDomain string `json:"clusterDomain,omitempty"`
    // +kubebuilder:default="ingress"
    // +kubebuilder:validation:Enum=route;ingress
    ArtemisExposeMode string `json:"artemisExposeMode,omitempty"`
}
```

Option B: Auto-detect by checking if `route.openshift.io` API group exists at startup:
```go
func detectArtemisExposeMode(ctx context.Context, c client.Client) string {
    // Try to list routes, if API exists, use "route"; otherwise "ingress"
    routeList := &unstructured.UnstructuredList{}
    routeList.SetAPIVersion("route.openshift.io/v1")
    routeList.SetKind("RouteList")
    if err := c.List(ctx, routeList, client.InNamespace("default"), client.Limit(1)); err == nil {
        return "route"
    }
    return "ingress"
}
```

#### 3b, Update the 4 Artemis resource builders

In each of the 4 files (see section 2.4), replace:
```go
"exposeMode": "route",
```
with:
```go
"exposeMode": exposeMode,  // passed as parameter from controller
```

---

### Step 4, Remove HAProxy annotation or make it conditional

**File**: `internal/controller/providervalidator_resources.go`, line ~498

Option A, Remove entirely (safe, it is only a timeout hint):
```go
// Delete this annotation block
Annotations: map[string]string{
    "haproxy.router.openshift.io/timeout": "300s",
},
```

Option B, Make conditional and add nginx equivalent:
```go
func timeoutAnnotations(isOpenShift bool) map[string]string {
    if isOpenShift {
        return map[string]string{"haproxy.router.openshift.io/timeout": "300s"}
    }
    return map[string]string{"nginx.ingress.kubernetes.io/proxy-read-timeout": "300"}
}
```

---

### Step 5, Update RBAC annotations and regenerate

In each of the 6 controller files, replace the OpenShift RBAC markers:

```go
// Remove these:
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes/custom-host,verbs=create;update;patch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=ingresscontrollers,verbs=get;list

// Add these:
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses,verbs=get;list;watch
```

Then regenerate:
```bash
make generate
make manifests
```

This regenerates `config/rbac/role.yaml` and all bundle manifests.

---

### Step 6, Remove `github.com/openshift/api` from go.mod

After all `routev1.*` usages are replaced:

```bash
go mod tidy
```

Verify:
```bash
grep "openshift" go.mod   # Must return nothing
grep "openshift" go.sum   # Must return nothing
```

---

### Step 7, Update OLM install manifest

**File**: `install-swim-catalog.yaml`

Change `namespace: openshift-marketplace` to `namespace: olm` (standard namespace when OLM
is installed on vanilla k8s via `operator-sdk olm install`):

```yaml
# Before
namespace: openshift-marketplace

# After
namespace: olm  # or remove entirely and provide kubectl apply path
```

Add a plain `kustomize`-based install section to `README.md` for vanilla k8s.

---

### Step 8, Update `bundle/metadata/annotations.yaml`

Remove or replace OCP-specific channel annotations:
```yaml
# Remove if present:
operators.operatorframework.io/builder: operator-sdk-v1.x.x
com.redhat.openshift.versions: "v4.12-v4.17"
```

---

## 6. Files Requiring No Changes

The following files are fully portable and **must not be modified** during this migration:

| File / Directory | Reason |
|-----------------|--------|
| `api/v1alpha1/*.go` | Only standard k8s types; no OpenShift imports |
| `api/v1alpha1/zz_generated.deepcopy.go` | Auto-generated; re-run `make generate` after changes |
| `internal/controller/consumer_resources.go` | No Route usage |
| `internal/controller/consumer_kafka_resources.go` | No Route usage |
| `internal/controller/ed254_consumer_resources.go` | No Route usage |
| `internal/controller/ed254_consumer_kafka_resources.go` | No Route usage |
| `internal/controller/swimdigitalnotamconsumer_controller.go` | No Route creation (only RBAC annotation needs updating) |
| `internal/controller/swimed254consumer_controller.go` | No Route creation (only RBAC annotation needs updating) |
| `internal/controller/artemis_ssl_secret.go` | PKCS12/JKS keystore logic: fully portable |
| `internal/controller/labels.go` | Pure label helpers |
| `internal/controller/shared_helpers.go` | Pure utility functions |
| `internal/controller/shared_reconcilers.go` | Generic k8s reconcile helpers; no OCP |
| `internal/controller/constants.go` | Constants only |
| `config/crd/` | CRD schemas: no OCP types |
| `config/manager/manager.yaml` | Standard Deployment |
| `config/network-policy/` | Standard NetworkPolicy |
| `config/prometheus/monitor.yaml` | `ServiceMonitor`: portable with prometheus-operator |
| `Dockerfile` | UBI base is publicly available |
| `test/e2e/e2e_test.go` | Uses `kubectl`: portable |
| `internal/controller/suite_test.go` | Already fixed to skip gracefully without envtest |
| `internal/controller/swimdigitalnotamconsumer_resources_test.go` | Pure unit tests: no infra |

---

## 7. Testing Strategy After Migration

### Unit tests (zero infra, run anywhere)
```bash
go test ./internal/controller/ -run "^Test" -v
```
Add new unit tests for:
- `ConsumerValidatorIngress()`, verify Ingress spec, TLS block, annotations
- `ProviderAppIngressEdge()` / `ProviderAppIngressPassthrough()`, verify passthrough annotation
- `detectClusterDomainFromIngressClass()`, mock client returning IngressClass
- Artemis builder with `exposeMode` parameter

### Integration tests (requires envtest or real cluster)
```bash
KUBEBUILDER_ASSETS=./bin/k8s/$(ls bin/k8s | head -1) go test ./internal/controller/ -v
```
These use `suite_test.go` which already skips gracefully without envtest binaries.

### E2E on Minikube
```bash
minikube start --driver=docker
# Install dependencies:
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
kubectl apply -f https://strimzi.io/install/latest?namespace=swim-system
# Install upstream Artemis operator
kubectl apply -f https://raw.githubusercontent.com/artemiscloud/activemq-artemis-operator/main/deploy/operator.yaml

# Deploy swim-operator
make deploy IMG=quay.io/masales/swim-operator:latest

# Apply a minimal consumer CR (works today without migration)
kubectl apply -f config/samples/apps_v1alpha1_swimdigitalnotamconsumer_minimal.yaml

# Apply a minimal provider CR (requires migration to be complete)
kubectl apply -f config/samples/apps_v1alpha1_swimdigitalnotamprovider_minimal.yaml
```

---

## 8. API Changes Summary for CR Users

After migration, the following CR spec changes should be documented:

### `GlobalSpec`, new field
```yaml
global:
  clusterDomain: "apps.mycluster.example.com"   # Required on vanilla k8s
  artemisExposeMode: "ingress"                    # New field; default "ingress" on vanilla
```

### TLS annotation for passthrough (nginx)
On OpenShift, TLS passthrough for Artemis AMQPS was handled by `Route`. On vanilla nginx:
```yaml
# Ingress annotation added automatically by the operator:
nginx.ingress.kubernetes.io/ssl-passthrough: "true"
```
The user may need to ensure nginx-ingress is deployed with `--enable-ssl-passthrough`.

---

## 9. Dependency Version Reference

| Module | Current version | Notes |
|--------|----------------|-------|
| `k8s.io/api` | `v0.35.3` | Includes `networking/v1.Ingress`, `networking/v1.IngressClass` |
| `k8s.io/apimachinery` | `v0.35.3` | |
| `k8s.io/client-go` | `v0.35.3` | |
| `sigs.k8s.io/controller-runtime` | `v0.23.3` | |
| `github.com/cert-manager/cert-manager` | `v1.20.1` | Portable |
| `github.com/openshift/api` | `v3.9.0+incompatible` | **TO REMOVE** |
| `github.com/prometheus-operator/...` | `v0.90.1` | Portable |
| `sigs.k8s.io/gateway-api` | `v1.5.0` | Already in `go.sum`: could be used as alternative to Ingress for AMQPS TCP routing |

**Note on Gateway API**: `sigs.k8s.io/gateway-api` is already a transitive dependency.
For AMQP/TCP passthrough, `TCPRoute` (Gateway API) is more appropriate than `Ingress` (which
is HTTP-centric). Consider using `gateway.networking.k8s.io/v1alpha2/TCPRoute` for the Artemis
AMQPS acceptor exposure instead of Ingress. This requires a Gateway API-capable controller
(e.g., Envoy Gateway, Traefik v3, nginx-gateway-fabric).

---

## 10. Checklist for the Migrating Agent

Use this checklist to track completion:

### Code changes
- [ ] `cluster_utils.go`, replace OCP-specific detection with `IngressClass`/`Ingress`-based detection
- [ ] `reconcilers.go`, rename `reconcileRoute` → `reconcileIngress`, change type to `*networkingv1.Ingress`
- [ ] `consumervalidator_resources.go`, replace `ConsumerValidatorRoute` with `ConsumerValidatorIngress`
- [ ] `providervalidator_resources.go`, replace `ProviderValidatorRoute` with `ProviderValidatorIngress`; remove HAProxy annotation
- [ ] `provider_app_resources.go`, replace Route builders with Ingress builders
- [ ] `ed254_provider_app_resources.go`, replace Route builders with Ingress builders
- [ ] `ed254_provider_reconcilers.go`, update to use Ingress reconciler
- [ ] `ed254consumervalidator_app_resources.go`, replace `Ed254CVRoute` with `Ed254CVIngress`
- [ ] `ed254consumervalidator_reconcilers.go`, update to use Ingress reconciler
- [ ] `swimdigitalnotamprovider_controller.go`, update `reconcileRoute` call, `Owns()`, RBAC annotations
- [ ] `swimdnotamconsumervalidator_controller.go`, update `reconcileRoute` call, `Owns()`, RBAC annotations
- [ ] `swimdnotamprovidervalidator_controller.go`, update `reconcileRoute` call, `Owns()`, RBAC annotations
- [ ] `swimed254provider_controller.go`, update `reconcileRoute` call, `Owns()`, RBAC annotations
- [ ] `swimed254consumervalidator_controller.go`, update `reconcileRoute` call, `Owns()`, RBAC annotations
- [ ] `swimdigitalnotamconsumer_controller.go`, update RBAC annotations only (no Route creation)
- [ ] `swimed254consumer_controller.go`, update RBAC annotations only (no Route creation)
- [ ] `artemis_resources.go`, replace hardcoded `"exposeMode": "route"` with parameter
- [ ] `provider_artemis_resources.go`, replace hardcoded `"exposeMode": "route"` with parameter
- [ ] `ed254_provider_artemis_resources.go`, replace hardcoded `"exposeMode": "route"` with parameter
- [ ] `ed254consumervalidator_artemis.go`, replace hardcoded `"exposeMode": "route"` with parameter
- [ ] `api/v1alpha1/*.go`, add `artemisExposeMode` to `GlobalSpec` (or shared spec)

### Build / manifests
- [ ] `go mod tidy`, remove `github.com/openshift/api`
- [ ] `make generate`, regenerate deepcopy
- [ ] `make manifests`, regenerate CRDs and RBAC from updated annotations
- [ ] Verify `config/rbac/role.yaml` no longer contains `route.openshift.io` or `operator.openshift.io`
- [ ] Verify `bundle/manifests/swim-operator.clusterserviceversion.yaml` RBAC is updated

### Testing
- [ ] `go test ./internal/controller/ -run "^Test" -v`, all unit tests pass
- [ ] Add unit tests for new Ingress builder functions
- [ ] Add unit test for `detectClusterDomain` with mocked `IngressClass`
- [ ] E2E test on Minikube with all 5 CR types

### Documentation
- [ ] `README.md`, add vanilla k8s installation section
- [ ] `config/samples/`, update sample CRs to include `global.clusterDomain` and `artemisExposeMode: ingress`
- [ ] `install-swim-catalog.yaml`, update namespace from `openshift-marketplace` to `olm`
