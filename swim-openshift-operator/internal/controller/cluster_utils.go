package controller

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	cachedClusterDomain string
	clusterDomainOnce   sync.Once
)

func detectClusterDomain(ctx context.Context, c client.Client, namespace string) string {
	clusterDomainOnce.Do(func() {
		logger := log.FromContext(ctx)

		if domain, err := getClusterDomainFromIngressController(ctx, c); err == nil {
			logger.Info("Cluster domain detected from IngressController", "domain", domain)
			cachedClusterDomain = domain
			return
		}

		if domain, err := getClusterDomainFromRoute(ctx, c, namespace); err == nil {
			logger.Info("Cluster domain detected from Route", "domain", domain)
			cachedClusterDomain = domain
			return
		}

		logger.Info("Using default cluster domain", "domain", "apps.cluster.local")
		cachedClusterDomain = "apps.cluster.local"
	})

	return cachedClusterDomain
}

// getClusterDomainFromIngressController extracts domain from OpenShift IngressController
func getClusterDomainFromIngressController(ctx context.Context, c client.Client) (string, error) {
	ingressController := &unstructured.Unstructured{}
	ingressController.SetAPIVersion("operator.openshift.io/v1")
	ingressController.SetKind("IngressController")

	err := c.Get(ctx, client.ObjectKey{
		Name:      "default",
		Namespace: "openshift-ingress-operator",
	}, ingressController)

	if err != nil {
		return "", fmt.Errorf("failed to get IngressController: %w", err)
	}

	// Extract domain from status
	domain, found, err := unstructured.NestedString(ingressController.Object, "status", "domain")
	if err != nil || !found {
		return "", fmt.Errorf("domain not found in IngressController status")
	}

	return domain, nil
}

// getClusterDomainFromRoute creates a temporary route to discover the cluster domain
func getClusterDomainFromRoute(ctx context.Context, c client.Client, namespace string) (string, error) {
	// Create a temporary route for domain discovery
	route := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "route.openshift.io/v1",
			"kind":       "Route",
			"metadata": map[string]interface{}{
				"name":      "swim-domain-discovery",
				"namespace": namespace,
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":      "swim-operator",
					"app.kubernetes.io/component": "domain-discovery",
					"app.kubernetes.io/temporary": "true",
				},
			},
			"spec": map[string]interface{}{
				"to": map[string]interface{}{
					"kind": "Service",
					"name": "nonexistent-service", // Service doesn't need to exist
				},
			},
		},
	}

	// Create the route
	err := c.Create(ctx, route)
	if err != nil {
		return "", fmt.Errorf("failed to create discovery route: %w", err)
	}

	// Ensure cleanup
	defer func() {
		if deleteErr := c.Delete(ctx, route); deleteErr != nil {
			log.FromContext(ctx).Error(deleteErr, "Failed to cleanup discovery route")
		}
	}()

	// Wait for OpenShift to populate the host
	time.Sleep(3 * time.Second)

	// Get the updated route with status
	err = c.Get(ctx, client.ObjectKeyFromObject(route), route)
	if err != nil {
		return "", fmt.Errorf("failed to get updated route: %w", err)
	}

	// Extract host from route status
	ingress, found, err := unstructured.NestedSlice(route.Object, "status", "ingress")
	if err != nil || !found || len(ingress) == 0 {
		return "", fmt.Errorf("no ingress found in route status")
	}

	ingressMap, ok := ingress[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid ingress format in route status")
	}

	host, found, err := unstructured.NestedString(ingressMap, "host")
	if err != nil || !found {
		return "", fmt.Errorf("host not found in route ingress status")
	}

	// Parse domain from host
	// Expected format: swim-domain-discovery-namespace.apps.ocp4.masales.cloud
	// The first part is the route name, everything after is the cluster domain
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("unable to parse domain from host: %s", host)
	}

	// Extract domain (everything after the first part which is route-namespace)
	domain := strings.Join(parts[1:], ".")
	return domain, nil
}

func getOrDetectClusterDomain(ctx context.Context, c client.Client, specDomain, namespace string) string {
	if specDomain != "" {
		return specDomain
	}
	return detectClusterDomain(ctx, c, namespace)
}
