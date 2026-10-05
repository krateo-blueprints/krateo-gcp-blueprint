---
title: Installing Config Connector
description: How Config Connector is installed for the Krateo GCP blueprints.
---

# Installing Config Connector

The blueprints in this repo render native Config Connector resources. Config Connector itself
is installed as a **Krateo blueprint**, from [`operator/`](../operator) — not by a hand
`helm install`.

Two compositions, in this order:

1. `krateo-gcp-configconnector-crds` — the `ConfigConnector` and `ConfigConnectorContext` CRDs.
2. `krateo-gcp-configconnector` — the operator, and the cluster-scoped `ConfigConnector` CR.

The CRDs ship as a dedicated composition so they land before anything creates a
`ConfigConnector`. The ~635 Google resource CRDs are **not** in either chart: the operator
installs those itself at runtime, once a `ConfigConnector` is applied.

## Choosing a mode

| Mode | Identity | Requires |
|---|---|---|
| `cluster` | one service account for the whole cluster | a key Secret (`credentialSecretName`) **or** Workload Identity (`googleServiceAccount`) |
| `namespaced` | a `ConfigConnectorContext` per namespace | **Workload Identity only** |

Namespaced mode has no key-based option: `ConfigConnectorContext` accepts only
`googleServiceAccount`. On a cluster without Workload Identity, cluster mode with a key Secret
is the only supported configuration.

`credentialSecretName` and `googleServiceAccount` are mutually exclusive. The chart fails the
render rather than letting the operator reject the object later, where the error is far less
visible.
