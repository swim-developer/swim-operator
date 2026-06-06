package resources

import (
	"encoding/json"

	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
)

type ProviderJSON struct {
	ProviderId          string                 `json:"providerId"`
	SubscriptionManager SubscriptionMgrJSON    `json:"subscriptionManager"`
	AmqpBroker          AmqpBrokerJSON         `json:"amqpBroker"`
}

type SubscriptionMgrJSON struct {
	URL        string          `json:"url"`
	TLS        *TLSJSON        `json:"tls"`
	Resilience *ResilienceJSON `json:"resilience,omitempty"`
}

type AmqpBrokerJSON struct {
	Host       string   `json:"host"`
	Port       int32    `json:"port"`
	SSLEnabled bool     `json:"sslEnabled"`
	Username   *string  `json:"username"`
	Password   *string  `json:"password"`
	TLS        *TLSJSON `json:"tls"`
}

type TLSJSON struct {
	TrustStorePath     string `json:"trustStorePath,omitempty"`
	TrustStorePassword string `json:"trustStorePassword,omitempty"`
	KeyStorePath       string `json:"keyStorePath,omitempty"`
	KeyStorePassword   string `json:"keyStorePassword,omitempty"`
}

type ResilienceJSON struct {
	ConnectTimeoutMs int32 `json:"connectTimeoutMs"`
	ReadTimeoutMs    int32 `json:"readTimeoutMs"`
	RetryMaxAttempts int32 `json:"retryMaxAttempts"`
	RetryDelayMs     int64 `json:"retryDelayMs"`
}

func SerializeProviders(providers []commonapi.ProviderSpec, keystorePassword string) string {
	defaultTLS := &TLSJSON{
		TrustStorePath:     "/secrets/truststore.p12",
		TrustStorePassword: keystorePassword,
		KeyStorePath:       "/secrets/keystore.p12",
		KeyStorePassword:   keystorePassword,
	}

	var result []ProviderJSON
	for _, p := range providers {
		pj := ProviderJSON{
			ProviderId: p.ProviderId,
			SubscriptionManager: SubscriptionMgrJSON{
				URL: p.SubscriptionManager.URL,
			},
			AmqpBroker: AmqpBrokerJSON{
				Host:       p.AmqpBroker.Host,
				Port:       p.AmqpBroker.Port,
				SSLEnabled: p.AmqpBroker.SSLEnabled,
			},
		}

		if p.SubscriptionManager.TLS != nil {
			pj.SubscriptionManager.TLS = &TLSJSON{
				TrustStorePath:     p.SubscriptionManager.TLS.TrustStorePath,
				TrustStorePassword: p.SubscriptionManager.TLS.TrustStorePassword,
				KeyStorePath:       p.SubscriptionManager.TLS.KeyStorePath,
				KeyStorePassword:   p.SubscriptionManager.TLS.KeyStorePassword,
			}
		} else {
			pj.SubscriptionManager.TLS = defaultTLS
		}

		if p.SubscriptionManager.Resilience != nil {
			pj.SubscriptionManager.Resilience = &ResilienceJSON{
				ConnectTimeoutMs: Int32Default(p.SubscriptionManager.Resilience.ConnectTimeoutMs, 5000),
				ReadTimeoutMs:    Int32Default(p.SubscriptionManager.Resilience.ReadTimeoutMs, 30000),
				RetryMaxAttempts: Int32Default(p.SubscriptionManager.Resilience.RetryMaxAttempts, 3),
				RetryDelayMs:     Int64Default(p.SubscriptionManager.Resilience.RetryDelayMs, 1000),
			}
		} else {
			pj.SubscriptionManager.Resilience = &ResilienceJSON{
				ConnectTimeoutMs: 5000,
				ReadTimeoutMs:    30000,
				RetryMaxAttempts: 3,
				RetryDelayMs:     1000,
			}
		}

		if p.AmqpBroker.Username != "" {
			u := p.AmqpBroker.Username
			pj.AmqpBroker.Username = &u
		}
		if p.AmqpBroker.Password != "" {
			pw := p.AmqpBroker.Password
			pj.AmqpBroker.Password = &pw
		}

		if p.AmqpBroker.TLS != nil {
			pj.AmqpBroker.TLS = &TLSJSON{
				TrustStorePath:     p.AmqpBroker.TLS.TrustStorePath,
				TrustStorePassword: p.AmqpBroker.TLS.TrustStorePassword,
				KeyStorePath:       p.AmqpBroker.TLS.KeyStorePath,
				KeyStorePassword:   p.AmqpBroker.TLS.KeyStorePassword,
			}
		} else if p.AmqpBroker.SSLEnabled {
			pj.AmqpBroker.TLS = defaultTLS
		}

		result = append(result, pj)
	}
	data, _ := json.Marshal(result)
	return string(data)
}

func SerializeDnotamSubscriptions(subs []commonapi.DnotamSubscriptionSpec) string {
	data, _ := json.Marshal(subs)
	return string(data)
}

func SerializeEd254Subscriptions(subs []commonapi.Ed254SubscriptionSpec) string {
	data, _ := json.Marshal(subs)
	return string(data)
}
