# Build the manager binary
FROM quay.io/masales/golang:1.25 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

FROM registry.access.redhat.com/ubi9-micro:9.5

LABEL org.opencontainers.image.licenses="GPL-3.0-or-later"
LABEL org.opencontainers.image.description="SWIM Operator - A Kubernetes/OpenShift operator that automates deployment and management of SWIM (System Wide Information Management) services for aviation, including DNOTAM Consumer, DNOTAM Provider, Consumer Validator, and Provider Validator. It abstracts the complexity of configuring resources like Deployments, Services, ConfigMaps, Secrets, Kafka, PostgreSQL, MongoDB, MariaDB, Keycloak, and Artemis through simple Custom Resources."
LABEL org.opencontainers.image.source="https://github.com/swim-developer/swim-openshift-operator"
LABEL org.opencontainers.image.version="0.0.1"

WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
