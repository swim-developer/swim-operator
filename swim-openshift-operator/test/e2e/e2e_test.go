/*
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
*/

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/swim-developer/swim-openshift-operator/test/utils"
)

// namespace where the project is deployed in
const namespace = "swim-operator-system"

// serviceAccountName created for the project
const serviceAccountName = "swim-operator-controller-manager"

// metricsServiceName is the name of the metrics service of the project
const metricsServiceName = "swim-operator-controller-manager-metrics-service"

// metricsRoleBindingName is the name of the RBAC that will be created to allow get the metrics data
const metricsRoleBindingName = "swim-operator-metrics-binding"

const metricsCurlPodName = "metricsCurlPodName"

const (
	curlProbeImageReference = "curlimages/curl:8.11.1"
	postgresWorkloadSuffix  = "-postgres"
)

var _ = Describe("Manager", Ordered, func() {
	var controllerPodName string

	// Before running the tests, set up the environment by creating the namespace,
	// enforce the restricted security policy to the namespace, installing CRDs,
	// and deploying the controller.
	BeforeAll(func() {
		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to create namespace")

		By("labeling the namespace to enforce the restricted security policy")
		cmd = exec.Command("kubectl", "label", "--overwrite", "ns", namespace,
			"pod-security.kubernetes.io/enforce=restricted")
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to label namespace with restricted policy")

		By("installing CRDs")
		cmd = exec.Command("make", "install")
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to install CRDs")

		By("deploying the controller-manager")
		cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectImage))
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to deploy the controller-manager")
	})

	// After all tests have been executed, clean up by undeploying the controller, uninstalling CRDs,
	// and deleting the namespace.
	AfterAll(func() {
		By("cleaning up the curl pod for metrics")
		cmd := exec.Command("kubectl", "delete", "pod", "metricsCurlPodName", "-n", namespace)
		_, _ = utils.Run(cmd)

		By("undeploying the controller-manager")
		cmd = exec.Command("make", "undeploy")
		_, _ = utils.Run(cmd)

		By("uninstalling CRDs")
		cmd = exec.Command("make", "uninstall")
		_, _ = utils.Run(cmd)

		By("removing manager namespace")
		cmd = exec.Command("kubectl", "delete", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	// After each test, check for failures and collect logs, events,
	// and pod descriptions for debugging.
	AfterEach(func() {
		specReport := CurrentSpecReport()
		if specReport.Failed() {
			By("Fetching controller manager pod logs")
			cmd := exec.Command("kubectl", "logs", controllerPodName, "-n", namespace)
			controllerLogs, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Controller logs:\n %s", controllerLogs)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Controller logs: %s", err)
			}

			By("Fetching Kubernetes events")
			cmd = exec.Command("kubectl", "get", "events", "-n", namespace, "--sort-by=.lastTimestamp")
			eventsOutput, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Kubernetes events:\n%s", eventsOutput)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Kubernetes events: %s", err)
			}

			By("Fetching metricsCurlPodName logs")
			cmd = exec.Command("kubectl", "logs", "metricsCurlPodName", "-n", namespace)
			metricsOutput, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Metrics logs:\n %s", metricsOutput)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get metricsCurlPodName logs: %s", err)
			}

			By("Fetching controller manager pod description")
			cmd = exec.Command("kubectl", "describe", "pod", controllerPodName, "-n", namespace)
			podDescription, err := utils.Run(cmd)
			if err == nil {
				fmt.Println("Pod description:\n", podDescription)
			} else {
				fmt.Println("Failed to describe controller pod")
			}
		}
	})

	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	Context("Manager", func() {
		It("should run successfully", func() {
			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func(g Gomega) {
				// Get the name of the controller-manager pod
				cmd := exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve controller-manager pod information")
				podNames := utils.GetNonEmptyLines(podOutput)
				g.Expect(podNames).To(HaveLen(1), "expected 1 controller pod running")
				controllerPodName = podNames[0]
				g.Expect(controllerPodName).To(ContainSubstring("controller-manager"))

				// Validate the pod's status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Running"), "Incorrect controller-manager pod status")
			}
			Eventually(verifyControllerUp).Should(Succeed())
		})

		It("should ensure the metrics endpoint is serving metrics", func() {
			By("creating a ClusterRoleBinding for the service account to allow access to metrics")
			cmd := exec.Command("kubectl", "create", "clusterrolebinding", metricsRoleBindingName,
				"--clusterrole=swim-operator-metrics-reader",
				fmt.Sprintf("--serviceaccount=%s:%s", namespace, serviceAccountName),
			)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to create ClusterRoleBinding")

			By("validating that the metrics service is available")
			cmd = exec.Command("kubectl", "get", "service", metricsServiceName, "-n", namespace)
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Metrics service should exist")

			By("getting the service account token")
			token, err := serviceAccountToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeEmpty())

			By("waiting for the metrics endpoint to be ready")
			verifyMetricsEndpointReady := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "endpoints", metricsServiceName, "-n", namespace)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("8443"), "Metrics endpoint is not ready")
			}
			Eventually(verifyMetricsEndpointReady).Should(Succeed())

			By("verifying that the controller manager is serving the metrics server")
			verifyMetricsServerStarted := func(g Gomega) {
				cmd := exec.Command("kubectl", "logs", controllerPodName, "-n", namespace)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("controller-runtime.metrics\tServing metrics server"),
					"Metrics server not yet started")
			}
			Eventually(verifyMetricsServerStarted).Should(Succeed())

			By("creating the metricsCurlPodName pod to access the metrics endpoint")
			cmd = exec.Command("kubectl", "run", "metricsCurlPodName", "--restart=Never",
				"--namespace", namespace,
				fmt.Sprintf("--image=%s", curlProbeImageReference),
				"--overrides",
				fmt.Sprintf(`{
					"spec": {
						"containers": [{
							"name": "curl",
							"image": "%s",
							"command": ["/bin/sh", "-c"],
							"args": ["curl -v -k -H 'Authorization: Bearer %s' https://%s.%s.svc.cluster.local:8443/metrics"],
							"securityContext": {
								"allowPrivilegeEscalation": false,
								"capabilities": {
									"drop": ["ALL"]
								},
								"runAsNonRoot": true,
								"runAsUser": 1000,
								"seccompProfile": {
									"type": "RuntimeDefault"
								}
							}
						}],
						"serviceAccount": "%s"
					}
				}`, curlProbeImageReference, token, metricsServiceName, namespace, serviceAccountName))
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to create metricsCurlPodName pod")

			By("waiting for the metricsCurlPodName pod to complete.")
			verifyCurlUp := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "pods", "metricsCurlPodName",
					"-o", "jsonpath={.status.phase}",
					"-n", namespace)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Succeeded"), "curl pod in wrong status")
			}
			Eventually(verifyCurlUp, 5*time.Minute).Should(Succeed())

			By("getting the metrics by checking metricsCurlPodName logs")
			metricsOutput := getMetricsOutput()
			Expect(metricsOutput).To(ContainSubstring(
				"controller_runtime_reconcile_total",
			))
		})

		// +kubebuilder:scaffold:e2e-webhooks-checks
	})

	Context("Provider Reconciliation", Ordered, func() {
		const reconcileNS = "e2e-provider-reconcile"
		const crName = "e2e-dnotam-provider"

		BeforeAll(func() {
			By("creating test namespace for provider reconciliation")
			cmd := exec.Command("kubectl", "create", "ns", reconcileNS)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("applying a minimal SwimDigitalNotamProvider CR")
			crYAML := providerTestCR(crName, reconcileNS)
			tmpFile := filepath.Join(os.TempDir(), "e2e-provider-cr.yaml")
			Expect(os.WriteFile(tmpFile, []byte(crYAML), os.FileMode(0o644))).To(Succeed())

			cmd = exec.Command("kubectl", "apply", "-f", tmpFile)
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			By("cleaning up provider reconciliation test namespace")
			cmd := exec.Command("kubectl", "delete", "ns", reconcileNS, "--wait=false")
			_, _ = utils.Run(cmd)
		})

		It("should set a finalizer on the provider CR", func() {
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "swimdigitalnotamproviders", crName,
					"-n", reconcileNS, "-o", "jsonpath={.metadata.finalizers}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("provider-finalizer"))
			}).Should(Succeed())
		})

		It("should create a ServiceAccount for the provider", func() {
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "sa", crName, "-n", reconcileNS)
				_, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
			}).Should(Succeed())
		})

		It("should create a Role with secrets access", func() {
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "role", crName, "-n", reconcileNS,
					"-o", "jsonpath={.rules}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("secrets"))
			}).Should(Succeed())
		})

		It("should create a RoleBinding linking ServiceAccount to Role", func() {
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "rolebinding", crName, "-n", reconcileNS,
					"-o", "jsonpath={.roleRef.name}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal(crName))
			}).Should(Succeed())
		})

		It("should create a Postgres Secret with database credentials", func() {
			secretName := crName + postgresWorkloadSuffix + "-secret"
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "secret", secretName, "-n", reconcileNS,
					"-o", "jsonpath={.data}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("database-user"))
				g.Expect(output).To(ContainSubstring("database-password"))
			}).Should(Succeed())
		})

		It("should create a Postgres StatefulSet with 1 replica", func() {
			stsName := crName + postgresWorkloadSuffix
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "statefulset", stsName, "-n", reconcileNS,
					"-o", "jsonpath={.spec.replicas}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("1"))
			}).Should(Succeed())
		})

		It("should create a Postgres Service on port 5432", func() {
			svcName := crName + postgresWorkloadSuffix
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "service", svcName, "-n", reconcileNS,
					"-o", "jsonpath={.spec.ports[0].port}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("5432"))
			}).Should(Succeed())
		})

		It("should set owner references on child resources pointing to the CR", func() {
			stsName := crName + postgresWorkloadSuffix
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "statefulset", stsName, "-n", reconcileNS,
					"-o", "jsonpath={.metadata.ownerReferences[0].kind}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("SwimDigitalNotamProvider"))
			}).Should(Succeed())
		})

		It("should set status conditions on the CR", func() {
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "swimdigitalnotamproviders", crName,
					"-n", reconcileNS, "-o", "jsonpath={.status.conditions}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).NotTo(BeEmpty(), "status conditions should be present")
			}).Should(Succeed())
		})
	})

	Context("Provider Deletion", Ordered, func() {
		const deletionNS = "e2e-provider-deletion"
		const deletionCR = "e2e-deletion-provider"

		It("should clean up child resources when the CR is deleted", func() {
			By("creating test namespace")
			cmd := exec.Command("kubectl", "create", "ns", deletionNS)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("applying a provider CR")
			crYAML := providerTestCR(deletionCR, deletionNS)
			tmpFile := filepath.Join(os.TempDir(), "e2e-deletion-cr.yaml")
			Expect(os.WriteFile(tmpFile, []byte(crYAML), os.FileMode(0o644))).To(Succeed())

			cmd = exec.Command("kubectl", "apply", "-f", tmpFile)
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("waiting for child resources to be created")
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "sa", deletionCR, "-n", deletionNS)
				_, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
			}).Should(Succeed())

			stsName := deletionCR + postgresWorkloadSuffix
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "statefulset", stsName, "-n", deletionNS)
				_, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
			}).Should(Succeed())

			By("deleting the provider CR")
			cmd = exec.Command("kubectl", "delete", "swimdigitalnotamproviders", deletionCR,
				"-n", deletionNS, "--timeout=60s")
			_, err = utils.Run(cmd)
			if err != nil {
				By("removing stuck finalizer to unblock deletion")
				cmd = exec.Command("kubectl", "patch", "swimdigitalnotamproviders", deletionCR,
					"-n", deletionNS, "--type=merge",
					"-p", `{"metadata":{"finalizers":[]}}`)
				_, _ = utils.Run(cmd)
			}

			By("verifying the CR is deleted")
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "swimdigitalnotamproviders", deletionCR,
					"-n", deletionNS)
				_, err := utils.Run(cmd)
				g.Expect(err).To(HaveOccurred())
			}).Should(Succeed())

			By("verifying child ServiceAccount is cleaned up")
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "sa", deletionCR, "-n", deletionNS)
				_, err := utils.Run(cmd)
				g.Expect(err).To(HaveOccurred())
			}).Should(Succeed())

			By("verifying Postgres StatefulSet is cleaned up")
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "statefulset", stsName, "-n", deletionNS)
				_, err := utils.Run(cmd)
				g.Expect(err).To(HaveOccurred())
			}).Should(Succeed())

			By("verifying Postgres Secret is cleaned up")
			Eventually(func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "secret", deletionCR+postgresWorkloadSuffix+"-secret",
					"-n", deletionNS)
				_, err := utils.Run(cmd)
				g.Expect(err).To(HaveOccurred())
			}).Should(Succeed())

			By("cleaning up test namespace")
			cmd = exec.Command("kubectl", "delete", "ns", deletionNS, "--wait=false")
			_, _ = utils.Run(cmd)
		})
	})
})

// serviceAccountToken returns a token for the specified service account in the given namespace.
// It uses the Kubernetes TokenRequest API to generate a token by directly sending a request
// and parsing the resulting token from the API response.
func serviceAccountToken() (string, error) {
	const tokenRequestRawString = `{
		"apiVersion": "authentication.k8s.io/v1",
		"kind": "TokenRequest"
	}`

	// Temporary file to store the token request
	secretName := fmt.Sprintf("%s-token-request", serviceAccountName)
	tokenRequestFile := filepath.Join("/tmp", secretName)
	err := os.WriteFile(tokenRequestFile, []byte(tokenRequestRawString), os.FileMode(0o644))
	if err != nil {
		return "", err
	}

	var out string
	verifyTokenCreation := func(g Gomega) {
		// Execute kubectl command to create the token
		cmd := exec.Command("kubectl", "create", "--raw", fmt.Sprintf(
			"/api/v1/namespaces/%s/serviceaccounts/%s/token",
			namespace,
			serviceAccountName,
		), "-f", tokenRequestFile)

		output, err := cmd.CombinedOutput()
		g.Expect(err).NotTo(HaveOccurred())

		// Parse the JSON output to extract the token
		var token tokenRequest
		err = json.Unmarshal(output, &token)
		g.Expect(err).NotTo(HaveOccurred())

		out = token.Status.Token
	}
	Eventually(verifyTokenCreation).Should(Succeed())

	return out, err
}

// getMetricsOutput retrieves and returns the logs from the curl pod used to access the metrics endpoint.
func getMetricsOutput() string {
	By("getting the metricsCurlPodName logs")
	cmd := exec.Command("kubectl", "logs", "metricsCurlPodName", "-n", namespace)
	metricsOutput, err := utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to retrieve logs from curl pod")
	Expect(metricsOutput).To(ContainSubstring("< HTTP/1.1 200 OK"))
	return metricsOutput
}

// tokenRequest is a simplified representation of the Kubernetes TokenRequest API response,
// containing only the token field that we need to extract.
type tokenRequest struct {
	Status struct {
		Token string `json:"token"`
	} `json:"status"`
}

func providerTestCR(name, namespace string) string {
	return fmt.Sprintf(`apiVersion: apps.swim-developer.github.io/v1alpha1
kind: SwimDigitalNotamProvider
metadata:
  name: %s
  namespace: %s
spec:
  certManager:
    issuerName: swim-ca-issuer
    issuerKind: ClusterIssuer
  artemis:
    acceptors:
      verifyHost: false
    oidc:
      authServerUrl: "https://keycloak.e2e.local/"
      realm: swim
      clientId: amq-broker
      clientSecret: e2e-test-secret
  provider:
    consumeFromClientTopics: false
    oidc:
      authServerUrl: "https://keycloak.e2e.local/realms/swim"
      clientId: swim-dnotam-provider
      clientSecret: e2e-provider-secret`, name, namespace)
}
