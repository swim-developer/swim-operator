# SWIM Operator - Test Resources

Test resources for validating operator functionality in OpenShift cluster.

## SwimDnotamConsumerValidator

### `swim-consumer-validator-min.yaml`
Minimal ConsumerValidator deployment for testing.
- ConsumerValidator HTTP API
- Artemis AMQP broker with mTLS
- Suitable for quick testing and development

---

## SwimDigitalNotamConsumer

### `swim-digital-notam-consumer-min.yaml`
Minimal client deployment.
- PostgreSQL for audit persistence
- Kafka managed (1 replica)
- Connects to ConsumerValidator for testing
- **Use case**: Development and quick testing

### `swim-digital-notam-consumer-full.yaml`
Production-ready client deployment.
- 2 client replicas
- 3 Kafka replicas with 10Gi storage
- 5Gi PostgreSQL storage
- Custom AMQP credentials
- **Use case**: Production or performance testing

### `swim-digital-notam-consumer-external-kafka.yaml`
Client with external Kafka cluster.
- No Kafka deployment by operator
- Requires external Kafka bootstrap servers
- Kafka credentials in spec (username/password)
- **Use case**: Integration with existing Kafka infrastructure

**Note**: Provider must also use external Kafka or manage its own if deploying both.

---

## SwimDigitalNotamProvider

### `swim-digital-notam-provider-min.yaml`
Minimal provider deployment.
- PostgreSQL 5Gi
- Artemis broker with mTLS + JMX
- Kafka managed (shares with Client if deployed)
- Provider app with 2 routes (Edge + Passthrough)
- **Use case**: Development and quick testing

### `swim-digital-notam-provider-full.yaml`
Production-ready provider deployment.
- PostgreSQL 10Gi with 2 replicas
- Artemis 2 replicas with custom broker properties
- Kafka 3 replicas with 10Gi storage
- Provider app 2 replicas
- Custom routes and resources
- **Use case**: Production deployment

### `swim-digital-notam-provider-external-kafka.yaml`
Provider with external Kafka cluster.
- No Kafka deployment by operator
- Requires external Kafka bootstrap servers
- PostgreSQL and Artemis managed by operator
- **Use case**: Integration with existing Kafka infrastructure

### Provider Configuration: `consumeFromClientTopics`

All Provider examples include the **required** field `provider.consumeFromClientTopics` (default: `false`):

**`consumeFromClientTopics: false` (Default)**:
- Provider consumes from its own topic: `dnotam-events-all-topic`
- `KAFKA_TOPIC=dnotam-events-all-topic`
- `KAFKA_PATTERN=false`
- **Use case**: Provider has dedicated topic, isolated from client topics

**`consumeFromClientTopics: true`**:
- Provider consumes from all client topics using regex pattern: `dnotam-events-.*`
- `KAFKA_TOPIC=dnotam-events-.*`
- `KAFKA_PATTERN=true`
- **Use case**: Provider needs to consume events from multiple client-generated topics

**Example**:
```yaml
spec:
  provider:
    consumeFromClientTopics: true  # Consume from client topics
```

---

## Complete Stack

### `swim-digital-notam-complete.yaml`
Full SWIM DNOTAM stack with Client + Provider.
- **Single Kafka cluster shared between Client and Provider**
- Client consumes from Provider's Artemis broker
- Complete end-to-end DNOTAM flow
- **Use case**: Integration testing and demonstration

**Architecture:**
```
Provider (Artemis AMQPS) → Kafka Topics → Client (PostgreSQL)
```

### `swim-digital-notam-kafka-sharing-test.yaml`
Test for Kafka sharing in reverse order.
- **Provider deployed FIRST** (creates Kafka)
- **Client deployed SECOND** (finds existing Kafka)
- Validates bidirectional Kafka detection logic
- **Use case**: Testing Provider → Client deployment order

**Architecture:**
```
1. Provider creates Kafka cluster
2. Client detects existing Kafka and skips creation
3. Both use the same Kafka cluster for topics
```

---

## Testing Order

**1. ConsumerValidator first:**
```bash
oc apply -f swim-consumer-validator-min.yaml
oc get swimdnotamconsumer-validator -n swim-sandbox
```

**2. Client only:**
```bash
oc apply -f swim-digital-notam-consumer-min.yaml
oc get swimdigitalnotamconsumer -n swim-sandbox
```

**3. Provider only:**
```bash
oc apply -f swim-digital-notam-provider-min.yaml
oc get swimdigitalnotamprovider -n swim-sandbox
```

**4. Complete stack:**
```bash
oc apply -f swim-digital-notam-complete.yaml
oc get swimdigitalnotamconsumer,swimdigitalnotamprovider -n swim-sandbox
```

---

## Verification Commands

**Check all SWIM resources:**
```bash
oc get swimdnotamconsumer-validator,swimdigitalnotamconsumer,swimdigitalnotamprovider -n swim-sandbox
```

**Check pods:**
```bash
oc get pods -n swim-sandbox
```

**Check Kafka (if managed):**
```bash
oc get kafka,kafkatopic -n swim-sandbox
```

**Check Artemis brokers:**
```bash
oc get activemqartemis -n swim-sandbox
```

**Check certificates:**
```bash
oc get certificate -n swim-sandbox
```

**Check routes:**
```bash
oc get route -n swim-sandbox
```

---

## Implementation Status

**What is implemented:**
- ✅ PostgreSQL managed by operator (StatefulSet)
- ✅ Artemis managed by operator (AMQ Operator)
- ✅ Kafka managed by operator (Strimzi)
- ✅ Kafka external mode (provide bootstrap servers, username, password in spec)
- ✅ PostgreSQL managed by operator for Client (StatefulSet)

**What is NOT implemented:**
- ❌ PostgreSQL external mode (no fields for external host/port/credentials)
- ❌ Artemis external mode (provider requires operator-managed Artemis)
- ❌ PostgreSQL external mode for Client (client requires operator-managed PostgreSQL)

**If you need external PostgreSQL:**
- Set `enabled: false` in spec
- Configure connection in provider/client application's ConfigMap manually
- This is a workaround until external mode is implemented

---

## Important Notes

**Kafka Sharing (Bidirectional Detection):**
- Only ONE Kafka cluster is deployed when both Client and Provider use `deploymentMode: managed`
- **Client → Provider**: Provider detects Kafka created by Client and skips creation
- **Provider → Client**: Client detects Kafka created by Provider and skips creation
- **Order doesn't matter**: Deploy Client first OR Provider first, the second one finds the existing Kafka
- Topics are created by both controllers (idempotent operation)
- The Kafka cluster name is always `kafka` in the namespace

**Kafka Sharing Flow:**
```
Scenario 1: Client First
1. Client reconciles → Kafka not found → Creates Kafka cluster
2. Provider reconciles → Kafka found → Skips creation, uses existing

Scenario 2: Provider First
1. Provider reconciles → Kafka not found → Creates Kafka cluster
2. Client reconciles → Kafka found → Skips creation, uses existing

Result: Same Kafka cluster used by both, regardless of deployment order
```

**Artemis Brokers:**
- ConsumerValidator has its own Artemis: `consumer-validator-artemis`
- Shared Artemis broker: `swim-artemis` (shared by all SWIM services)
- Client connects to Provider's Artemis for production flow
- No coupling between ConsumerValidator and Provider/Client artemis instances

**Certificates:**
- All mTLS certificates issued by `swim-ca-issuer` (ClusterIssuer)
- JKS keystores automatically generated by cert-manager
- SSL secrets transformed inline by operator (no Jobs required)

**Storage:**
- All PVCs use default StorageClass
- Customize `storageSize` per component as needed
- StatefulSets for Postgres (persistence critical)
- Deployments for apps (stateless)

---

## Cleanup

**Delete specific resource:**
```bash
oc delete swimdigitalnotamconsumer dnotam-client -n swim-sandbox
oc delete swimdigitalnotamprovider dnotam-provider -n swim-sandbox
```

**Delete all SWIM resources:**
```bash
oc delete swimdnotamconsumer-validator,swimdigitalnotamconsumer,swimdigitalnotamprovider --all -n swim-sandbox
```

**Finalizers ensure clean deletion of external resources (Kafka, Topics).**

