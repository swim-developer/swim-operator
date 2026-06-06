package controller

import (
	"context"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/helpers"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	KafkaAPIVersion      = constants.KafkaAPIVersion
	KafkaGroup           = constants.KafkaGroup
	DatabasePasswordKey  = constants.DatabasePasswordKey
	DatabaseUserKey      = constants.DatabaseUserKey
	ConsumerValidatorApp = constants.ConsumerValidatorApp
	ProviderValidatorApp = constants.ProviderValidatorApp
	MTLSCertsVolume      = constants.MTLSCertsVolume
	CurlMetricsContainer = constants.CurlMetricsContainer

	ArtemisSuffix            = constants.ArtemisSuffix
	KeystorePasswordSuffix   = constants.KeystorePasswordSuffix
	MongoDBCredentialsSuffix = constants.MongoDBCredentialsSuffix
	MongoDBSuffix            = constants.MongoDBSuffix
	MongoDBDataSuffix        = constants.MongoDBDataSuffix
	ServerTLSSuffix          = constants.ServerTLSSuffix
	HostnameSuffix           = constants.HostnameSuffix
	MTLSHostnameSuffix       = constants.MTLSHostnameSuffix
	SSOJAASConfigSuffix      = constants.SSOJAASConfigSuffix
	PostgresSuffix           = constants.PostgresSuffix
	PostgresSecretSuffix     = constants.PostgresSecretSuffix
	SSLSecretSuffix          = constants.SSLSecretSuffix

	ErrUnableToCreateController = constants.ErrUnableToCreateController
	ErrFailedToWriteOutput      = constants.ErrFailedToWriteOutput
	InfoFoundActiveSWIMCR       = constants.InfoFoundActiveSWIMCR
)

func ensureCRLabels(ctx context.Context, c client.Client, obj client.Object, component string) error {
	return labels.EnsureCRLabels(ctx, c, obj, component, sharedManagedByValue)
}

func standardLabels(appName, component, crName string) map[string]string {
	return labels.StandardLabels(appName, component, crName, sharedManagedByValue)
}

func shouldRequeue(result ctrl.Result) bool {
	return result.RequeueAfter > 0
}

func providerRequeueResult(result ctrl.Result) bool {
	return result.RequeueAfter > 0
}

func isPodReady(ctx context.Context, c client.Client, namespace string, lbls map[string]string) bool {
	return helpers.IsPodReady(ctx, c, namespace, lbls)
}

func isKafkaClusterReady(ctx context.Context, c client.Client, namespace, name string) bool {
	return helpers.IsKafkaClusterReady(ctx, c, namespace, name)
}

