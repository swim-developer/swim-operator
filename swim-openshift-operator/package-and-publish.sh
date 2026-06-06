#!/bin/bash

export VERSION=0.0.1
export IMG=quay.io/masales/swim-operator:v${VERSION}
export BUNDLE_IMG=quay.io/masales/swim-operator-bundle:v${VERSION}
export CATALOG_IMG=quay.io/masales/swim-operator-catalog:v${VERSION}

make docker-build IMG=$IMG
make docker-push IMG=$IMG

make bundle IMG=$IMG VERSION=$VERSION
make bundle-build BUNDLE_IMG=$BUNDLE_IMG
make bundle-push BUNDLE_IMG=$BUNDLE_IMG

make catalog-build catalog-push CATALOG_IMG=$CATALOG_IMG BUNDLE_IMGS=$BUNDLE_IMG