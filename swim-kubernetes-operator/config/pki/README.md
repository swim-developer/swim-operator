# SWIM PKI - Certificate Management

This directory contains the PKI (Public Key Infrastructure) configuration for SWIM services using cert-manager with self-signed certificates.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         SWIM PKI Hierarchy                          │
└─────────────────────────────────────────────────────────────────────┘

    selfsigned-bootstrap (ClusterIssuer)
              │
              │  generates
              ▼
       ┌──────────────┐
       │   swim-ca    │  Certificate (namespace: cert-manager)
       │  Root CA     │  Secret: swim-ca-secret
       │  10 years    │  RSA 4096-bit
       └──────┬───────┘
              │
              │  signs via
              ▼
    swim-ca-issuer (ClusterIssuer)
              │
    ┌─────────┼─────────┬─────────────┬────────────┐
    │         │         │             │            │
    ▼         ▼         ▼             ▼            ▼
Keycloak  Provider  ConsumerValidator  Consumer    ProviderValidator
  TLS       TLS       mTLS        mTLS         TLS
```

## Why Self-Signed Instead of Let's Encrypt?

**Let's Encrypt does NOT work with nip.io** because:
- Let's Encrypt requires domain validation (HTTP-01 or DNS-01 challenge)
- You don't control the nip.io domain
- Cannot prove domain ownership

**Self-signed CA works perfectly** for:
- Development environments
- Minikube/Kind clusters
- Testing mTLS configurations
- Internal services

## Components

### 1. selfsigned-bootstrap (ClusterIssuer)
Bootstrap issuer that can generate self-signed certificates. Used only to create the root CA.

### 2. swim-ca (Certificate)
- **Namespace**: cert-manager
- **Secret**: swim-ca-secret
- **Duration**: 10 years
- **Key**: RSA 4096-bit
- **Purpose**: Root CA for all SWIM certificates

### 3. swim-ca-issuer (ClusterIssuer)
Uses the swim-ca secret to issue certificates for all SWIM services.

## Usage

### Apply PKI Resources

```bash
kubectl apply -f swim-pki.yaml
```

### Verify CA is Ready

```bash
kubectl get certificate swim-ca -n cert-manager
kubectl get clusterissuer swim-ca-issuer
```

### Export CA Certificate

Export the CA to trust in browsers or clients:

```bash
kubectl get secret swim-ca-secret -n cert-manager \
  -o jsonpath='{.data.ca\.crt}' | base64 -d > swim-ca.crt
```

## Trusting the CA Certificate (Remove Browser Warnings)

By default, browsers will show "Your connection is not private" warnings because the CA is self-signed. To remove these warnings, import the CA certificate into your system or browser.

### Linux (Fedora/RHEL/CentOS)

```bash
sudo cp swim-ca.crt /etc/pki/ca-trust/source/anchors/swim-ca.crt
sudo update-ca-trust

# Verify
trust list | grep -i swim
```

**For Ubuntu/Debian:**

```bash
sudo cp swim-ca.crt /usr/local/share/ca-certificates/swim-ca.crt
sudo update-ca-certificates

# Verify
ls /etc/ssl/certs | grep swim
```

### macOS

**Option 1: Command line (recommended)**

```bash
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain swim-ca.crt
```

**Option 2: Keychain Access GUI**

1. Open **Keychain Access** (Applications → Utilities → Keychain Access)
2. Select **System** keychain in the left sidebar
3. File → Import Items → Select `swim-ca.crt`
4. Double-click the imported certificate
5. Expand **Trust** section
6. Set **When using this certificate** to **Always Trust**
7. Close and enter your password

**To remove:**

```bash
sudo security delete-certificate -c "swim-ca" -t /Library/Keychains/System.keychain
```

### Windows

**Option 1: PowerShell (as Administrator)**

```powershell
Import-Certificate -FilePath "swim-ca.crt" -CertStoreLocation Cert:\LocalMachine\Root
```

**Option 2: GUI**

1. Double-click `swim-ca.crt`
2. Click **Install Certificate...**
3. Select **Local Machine** → Next
4. Select **Place all certificates in the following store**
5. Click **Browse...** → Select **Trusted Root Certification Authorities**
6. Click Next → Finish
7. Confirm the security warning

**To verify (PowerShell):**

```powershell
Get-ChildItem Cert:\LocalMachine\Root | Where-Object { $_.Subject -like "*swim*" }
```

**To remove (PowerShell as Administrator):**

```powershell
Get-ChildItem Cert:\LocalMachine\Root | Where-Object { $_.Subject -like "*swim*" } | Remove-Item
```

### Browser-Specific Import

Some browsers use their own certificate store instead of the system store.

#### Firefox (All Platforms)

Firefox uses its own certificate store:

1. Open Firefox → Settings → Privacy & Security
2. Scroll to **Certificates** → Click **View Certificates...**
3. Go to **Authorities** tab
4. Click **Import...** → Select `swim-ca.crt`
5. Check **Trust this CA to identify websites**
6. Click OK

#### Chrome/Edge (Windows/macOS)

Chrome and Edge use the system certificate store. After importing the CA at the system level (see above), restart the browser.

#### Chrome (Linux)

Chrome on Linux uses NSS database:

```bash
# Install certutil if needed
sudo dnf install nss-tools  # Fedora
sudo apt install libnss3-tools  # Ubuntu

# Import CA
certutil -d sql:$HOME/.pki/nssdb -A -t "C,," -n "SWIM CA" -i swim-ca.crt

# Verify
certutil -d sql:$HOME/.pki/nssdb -L

# Remove (if needed)
certutil -d sql:$HOME/.pki/nssdb -D -n "SWIM CA"
```

### Using curl/wget with the CA

If you don't want to import the CA system-wide:

```bash
# curl
curl --cacert swim-ca.crt https://keycloak.192.168.49.2.nip.io

# wget
wget --ca-certificate=swim-ca.crt https://keycloak.192.168.49.2.nip.io

# Skip verification (not recommended, but useful for quick tests)
curl -k https://keycloak.192.168.49.2.nip.io
```

### Postman

1. Settings → Certificates
2. **CA Certificates** → Toggle ON
3. Select `swim-ca.crt` file

### Java Applications

```bash
# Import to Java truststore
keytool -importcert -trustcacerts -alias swim-ca \
  -file swim-ca.crt \
  -keystore $JAVA_HOME/lib/security/cacerts \
  -storepass changeit -noprompt

# Verify
keytool -list -keystore $JAVA_HOME/lib/security/cacerts -storepass changeit | grep swim
```

## Creating Certificates for Services

Services use the `swim-ca-issuer` ClusterIssuer. Example:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: my-service-tls
  namespace: my-namespace
spec:
  secretName: my-service-tls-secret
  duration: 8760h    # 1 year
  renewBefore: 720h  # 30 days
  privateKey:
    algorithm: RSA
    size: 2048
  usages:
    - server auth
    - client auth  # for mTLS
  dnsNames:
    - my-service.192.168.49.2.nip.io
    - my-service.my-namespace.svc.cluster.local
    - my-service
  issuerRef:
    name: swim-ca-issuer
    kind: ClusterIssuer
    group: cert-manager.io
```

## Ingress TLS Configuration

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-service
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - my-service.192.168.49.2.nip.io
      secretName: my-service-tls-secret
  rules:
    - host: my-service.192.168.49.2.nip.io
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-service
                port:
                  number: 8080
```

## mTLS Configuration

For mutual TLS, create client certificates:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: client-cert
  namespace: my-namespace
spec:
  secretName: client-cert-secret
  duration: 8760h
  renewBefore: 720h
  privateKey:
    algorithm: RSA
    size: 2048
  usages:
    - client auth
  commonName: my-client
  issuerRef:
    name: swim-ca-issuer
    kind: ClusterIssuer
```

Configure Ingress for mTLS:

```yaml
annotations:
  nginx.ingress.kubernetes.io/auth-tls-verify-client: "on"
  nginx.ingress.kubernetes.io/auth-tls-secret: "cert-manager/swim-ca-secret"
```

## Troubleshooting

### Certificate not ready

```bash
kubectl describe certificate swim-ca -n cert-manager
kubectl logs -n cert-manager -l app=cert-manager
```

### ClusterIssuer not ready

```bash
kubectl describe clusterissuer swim-ca-issuer
```

### Secret not created

```bash
kubectl get secret swim-ca-secret -n cert-manager
kubectl describe secret swim-ca-secret -n cert-manager
```

## Integration with SWIM Operator

SWIM CRDs support cert-manager integration:

```yaml
spec:
  certManager:
    issuerName: "swim-ca-issuer"
    issuerKind: "ClusterIssuer"
```

The operator will automatically create Certificate resources for each component.

