#!/bin/bash

oc delete swimdnotamconsumervalidator --all -A
oc delete swimdigitalnotamconsumer --all -A
oc delete swimdigitalnotamprovider --all -A

oc delete -f install-swim-catalog.yaml

oc delete crd swimdnotamconsumervalidators.apps.swim-developer.github.io
oc delete crd swimdigitalnotamconsumers.apps.swim-developer.github.io
oc delete crd swimdigitalnotamproviders.apps.swim-developer.github.io
