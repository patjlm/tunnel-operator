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
  # - hostname: example12.zeeweb.xyz
  #   service: tcp://localhost:10000

status:
  accountid: xxx
  tunnelid: yyy-zzz
  hostnames:
    - example1.zeeweb.xyz
  conditions:
  - lastTransitionTime: "2022-01-28T16:09:46Z"
    message: Cloudflare tunnel created successfully with ID yyy-zzz
    reason: CreationSucceeded
    status: "True"
    type: Created
```
The operator creates a secret (by default named after the `Tunnel` resource) containing the necessary files to execute `cloudflared run`: `credentials.json` and `config.yaml`

With `run: true`, the operator will start a 1-replica deployment executing `cloudflared run`, providing ingress access to the cluster

## TODO
* improve the `run: bool` feature:
  * run custom images, avoid downloading cloudflared at runtime, ...
* support more `config.yaml` features (i think i saw one to trust self-signed certs on the backend service, useful for internal kubernetes services)
* implement a `TunnelAccess` custom resource which can be used to run a `cloudflared access` deployment in order to access a remote tunnel TCP endpoint.
