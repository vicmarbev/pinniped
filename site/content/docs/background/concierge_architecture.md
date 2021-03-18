---
title: Concierge Architecture
description: Dive into the overall design and implementation details of Pinniped.
cascade:
  layout: docs
menu:
  docs:
    name: Concierge
    weight: 100
    parent: background
---

The Pinniped Concierge is a credential exchange API which takes as input a
credential from an identity source (e.g., Pinniped Supervisor, proprietary IDP),
authenticates the user via that credential, and returns another credential which is
understood by the host Kubernetes cluster or by an impersonation proxy which acts
on behalf of the user.

## Cluster Integration Strategies
The Pinniped Concierge will issue a cluster credential by leveraging cluster-specific
functionality. In the longer term,
Pinniped hopes to contribute and leverage upstream Kubernetes extension points that
cleanly enable this integration.

### Token Credential Request API
Pinniped hosts a credential exchange API endpoint via a Kubernetes aggregated API server.
This API returns a new cluster-specific credential.
When possible, this short-lived certificate will be created using the cluster's signing keypair.
Otherwise, it will be signed by Pinniped's own keys.
If the cert was made with the cluster's signing keypair, it can be used with or without the impersonation
proxy, while otherwise it is only meant to be used with the impersonation proxy, see below.
(In the future, when the Kubernetes CSR API
provides a way to issue short-lived certificates, then the Pinniped credential exchange API
will use that instead of using the cluster's signing keypair.)

![concierge-with-webhook-architecture-diagram](/docs/img/pinniped_architecture_concierge_webhook.svg)

### Impersonation Proxy
Pinniped hosts an [impersonation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation)
proxy that sends requests to the Kubernetes API server with user information and permissions attached in impersonation
headers.
All requests can be routed through the impersonation proxy as if they were going directly to the Kubernetes API.
After the TokenCredentialRequest is made, and the concierge has returned an x509 cert, it will pass that cert on to
the impersonation proxy.
The impersonation parses the certificate and passes through the request with impersonate headers that
indicate to the Kubernetes API server that we are performing the request on the user's behalf, with their permissions.
  
![Concierge With Impersonation Proxy Sketch](/docs/img/pinniped_architecture_concierge_webhook_impersonator.svg)





