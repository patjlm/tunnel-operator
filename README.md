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

## TODO
* implement a `run: bool` field in the `Tunnel`. When this field is set, run a deployment which executed `cloudflared run` for this tunnel, allowing ingress traffic in Kubernetes.
* implement a `TunnelAccess` custom resource which can be used to run a `cloudflared access` deployment in order to access a remote tunnel TCP endpoint.
