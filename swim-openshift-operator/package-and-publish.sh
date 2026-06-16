#!/bin/bash

export VERSION=0.0.1
export IMG=quay.io/masales/swim-openshift-operator:v${VERSION}
export BUNDLE_IMG=quay.io/masales/swim-openshift-operator-bundle:v${VERSION}
export CATALOG_IMG=quay.io/masales/swim-openshift-operator-catalog:v${VERSION}

make image-build IMG=$IMG
make push IMG=$IMG

make bundle IMG=$IMG VERSION=$VERSION
make bundle-build BUNDLE_IMG=$BUNDLE_IMG
make bundle-push BUNDLE_IMG=$BUNDLE_IMG

make catalog-build catalog-push CATALOG_IMG=$CATALOG_IMG BUNDLE_IMGS=$BUNDLE_IMG