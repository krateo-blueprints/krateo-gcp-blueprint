---
title: Authentication
description: Google Cloud credentials for Config Connector.
---

# Authentication

Config Connector authenticates to Google Cloud as a Google service account. Grant that account
only the roles the blueprints you publish actually need — it is the identity every Composition
in the cluster provisions through.

## Workload Identity (preferred)

On a GKE cluster with Workload Identity enabled, set `configConnector.googleServiceAccount` and
bind the Google service account to the Kubernetes service account. No long-lived key exists,
which is why this is the default recommendation.

## Service-account key

On a cluster without Workload Identity, cluster mode takes a key Secret instead:

```sh
gcloud iam service-accounts keys create key.json \
  --iam-account <name>@<project>.iam.gserviceaccount.com
kubectl create secret generic gcp-key --from-file key.json -n cnrm-system
```

Then set `configConnector.credentialSecretName` to that Secret's name.

`cnrm-system` is created by the operator when it reconciles the `ConfigConnector`, so the Secret
is created after the operator is running. The controller manager recovers once it appears.

A key is a long-lived credential: rotate it, and prefer Workload Identity wherever the cluster
supports it.
