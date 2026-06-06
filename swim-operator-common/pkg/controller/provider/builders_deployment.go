package provider

import (
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
)

func BuildProviderAppDeployment(p ProviderBuildParams, managedBy string, configHash string) *appsv1.Deployment {
	name := p.Name
	replicas := p.Strategy.AppReplicas()
	return resources.BuildProviderAppDeployment(resources.ProviderAppDeploymentParams{
		Name:                  name,
		Namespace:             p.Namespace,
		Labels:                labels.StandardLabels(name, "provider", name, managedBy),
		Image:                 p.Strategy.AppImage(),
		Replicas:              replicas,
		ContainerResources:    p.Strategy.AppResources(),
		ConfigHash:            configHash,
		ServerTLSSecretName:   fmt.Sprintf(constants.ServerTLSSuffix, name),
		CABundleConfigMapName: fmt.Sprintf("%s-ca-bundle", name),
	})
}
