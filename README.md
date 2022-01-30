# tunnel operator

This project provides a kubernetes operator which handles [Cloudfare tunnels](https://www.cloudflare.com/products/tunnel/)

## Custom Resources

### Tunnel

The `Tunnel` custom resource definition allows to create cloudflare tunnels and DNS CNAME records for their routes:
```yaml
apiVersion: tunnel.zeeweb.xyz/v1alpha1
kind: Tunnel
metadata:
  name: example1

spec:
  name: example1
  # output secret.
  # cloudflared credentials.json and config.yaml will be created in this secret
  # secret:
  #   name: mysecret
  #   namepsace: mynamespace
  run: true
  ingress:
  - hostname: example1.zeeweb.xyz
    service: tcp://localhost:10000
  # forward kd.zeeweb.xyz to the local kubernetes cluster API
  - hostname: kd.zeeweb.xyz
    service: https://kubernetes.default
    # all customization from https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/configuration/configuration-file/ingress are available
    originRequest:
      noTLSVerify: true
  # - hostname: example12.zeeweb.xyz
  #   service: tcp://localhost:10000

status:
  accountid: xxx
  tunnelid: yyy-zzz
  hostnames:
    - example1.zeeweb.xyz
    - kd.zeeweb.xyz
  conditions:
  - lastTransitionTime: "2022-01-28T16:09:46Z"
    message: Cloudflare tunnel created successfully with ID yyy-zzz
    reason: CreationSucceeded
    status: "True"
    type: Created
```

The operator creates a secret (by default named after the `Tunnel` resource) containing the necessary files to execute `cloudflared run`: `credentials.json` and `config.yaml`

With `run: true`, the operator will start a deployment executing `cloudflared tunnel run`, providing ingress access to the cluster. The deployment being created can be fully customizable by specifying a `deploymentSpec` field. When not specified, the default deploymentSpec will be added to the `Tunnel` resource so the user can customize it easily.

The default deployment will optionally mount a configmap named `openshift-ca` into `/openshift-ca`. This allows to get access to the internal CA and validate automatically generated certs.

## TODO
* implement a `TunnelAccess` custom resource which can be used to run a `cloudflared access` deployment in order to access a remote tunnel TCP endpoint.
