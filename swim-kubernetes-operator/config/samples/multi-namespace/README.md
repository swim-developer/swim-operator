# Multi-Namespace Deployment

This folder contains samples configured for a multi-namespace deployment:

## Namespace Layout

| Namespace | Components |
|-----------|------------|
| `swim-consumervalidator` | ConsumerValidator + Artemis + MariaDB |
| `swim-backend` | Consumer + Provider + Kafka + MongoDB + Artemis + PostgreSQL |
| `swim-providervalidator` | ProviderValidator |

## Connection Map

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│  swim-consumervalidator                                                 │
│  ┌──────────────────┐   ┌─────────────────────────────┐   ┌──────────────────┐     │
│  │ConsumerValidator │◄──│ swim-consumervalidator-artemis │  │     MariaDB      │     │
│  │   :8080/:8443    │   │         :5672               │   │      :3306       │     │
│  └────────┬──────────┘   └─────────────────────────────┘   └──────────────────┘     │
│           │                     ▲                                       │
└───────────┼─────────────────────┼───────────────────────────────────────┘
            │                     │
            │ HTTP API            │ AMQP (events)
            ▼                     │
┌───────────────────────────────────────────────────────────────────────────┐
│  swim-backend                   │                                         │
│                                 │                                         │
│  ┌─────────────────┐   ┌───────┴──────┐   ┌──────────────────┐           │
│  │    Consumer     │◄──│    Kafka     │◄──│    Provider      │           │
│  │                 │   │    :9092     │   │   :8080/:8443    │           │
│  └────────┬────────┘   └──────────────┘   └────────┬─────────┘           │
│           │                                        │                      │
│           ▼                                        ▼                      │
│  ┌─────────────────┐                      ┌──────────────────┐           │
│  │    MongoDB      │                      │ swim-artemis │           │
│  │     :27017      │                      │      :5672       │           │
│  └─────────────────┘                      └──────────────────┘           │
│                                                    │                      │
│                                           ┌───────┴──────────┐           │
│                                           │   PostgreSQL     │           │
│                                           │     :5432        │           │
│                                           └──────────────────┘           │
└───────────────────────────────────────────────────────────────────────────┘
            ▲
            │ HTTP API
            │
┌───────────┴─────────────────────────────────────────────────────────────┐
│  swim-providervalidator                                                 │
│  ┌─────────────────┐                                                    │
│  │ ProviderValidator│                                                    │
│  │   (Angular UI)  │                                                    │
│  └─────────────────┘                                                    │
└─────────────────────────────────────────────────────────────────────────┘
```

## Deploy Order

1. Create namespaces
2. Deploy ConsumerValidator (swim-consumervalidator)
3. Deploy Provider + Consumer (swim-backend)
4. Deploy ProviderValidator (swim-providervalidator)

## Commands

```bash
# Create namespaces
kubectl create namespace swim-consumervalidator
kubectl create namespace swim-backend
kubectl create namespace swim-providervalidator

# Deploy in order
kubectl apply -f consumervalidator.yaml
kubectl apply -f provider.yaml
kubectl apply -f consumer.yaml
kubectl apply -f providervalidator.yaml

# Verify
kubectl get pods -n swim-consumervalidator
kubectl get pods -n swim-backend
kubectl get pods -n swim-providervalidator
```

## Cross-Namespace DNS

All connections use Kubernetes DNS FQDNs:

| Service | FQDN |
|---------|------|
| ConsumerValidator API | `swim-consumervalidator.swim-consumervalidator.svc.cluster.local:8080` |
| ConsumerValidator Artemis | `swim-consumervalidator-artemis-hdls-svc.swim-consumervalidator.svc.cluster.local:5672` |
| Provider API | `swim-dnotam-provider.swim-backend.svc.cluster.local:8080` |
| Provider Artemis | `swim-artemis-hdls-svc.swim-backend.svc.cluster.local:5672` |
| Kafka | `kafka-kafka-bootstrap.swim-backend.svc.cluster.local:9092` |

