# grpc-xds-sample

In this repo:
* Sample Go gRPC application, with code to use xDS features:
  * [Client-side configuration](https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md)
  * [Server-side configuration](https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md)
  * [Other gRFCs](https://github.com/grpc/proposal)
* A sample Kubernetes deployment compatible with [Istio](https://github.com/istio/isttio).

This is intended mostly as a guide for how to make your own gRPC applications compatible with xDS.

## As a guide for writing your own applications

### Clients

First, your application _must_ import the gRPC xDS package.

https://github.com/stevenctl/grpc-xds-sample/blob/092b8406e45586d61140ead68d139f0a19ce2516/main.go#L14-L16 

If you don't actually use the import, still include it as a side-effect import like this:

```go
// install xDS resolvers
_ "google.golang.org/grpc/xds"
```

The next important bit is to make sure your `Dial` or `DialContext` calls use the `xds:///` scheme.
In this app, we build the URL dynamically, but prefix it with `xds:///`

https://github.com/stevenctl/grpc-xds-sample/blob/092b8406e45586d61140ead68d139f0a19ce2516/main.go#L112-L117

To enable client-side security, we pass a `TransportCredentials` option. To allow the control plane to send
an empty security configuraiton, we also include a fallback of `insecure`.

https://github.com/stevenctl/grpc-xds-sample/blob/092b8406e45586d61140ead68d139f0a19ce2516/main.go#L44-L59

### Servers

For servers, the main step is to create the server with a special constructor:

https://github.com/stevenctl/grpc-xds-sample/blob/092b8406e45586d61140ead68d139f0a19ce2516/main.go#L76

Note that this returns an `xds.GRPCServer`. If your protobuf generated Go code used an older version of
protoc-gen-go-grpc, it may need to be regenerated so that the `RegisterServiceNameServer` function accepts
this new implementation:

https://github.com/stevenctl/grpc-xds-sample/blob/092b8406e45586d61140ead68d139f0a19ce2516/greeter/foo_grpc.pb.go#L65

Finally, similarly to clients we use special credentials and a fallback to enable server side security:

https://github.com/stevenctl/grpc-xds-sample/blob/092b8406e45586d61140ead68d139f0a19ce2516/main.go#L68-L76

### Environment

The code changes alone aren't enough to enable xDS in gRPC. We still need to connect to a control plane.
Istio makes this easy. We add an annotation that tells Istio's sidecar injector to do a few things for us:

* Run `pilot-agent` in a special mode that does _not_ run Envoy proxy. Instead it:
  * Fetches certificates and places them on a volume that the gRPC library can access.
  * Proxies connections from the gRPC library to the `istiod` control plane and handles authentication.
  * Generates a bootstrap file to tell gRPC how to reach the control plane and where to find data plane certs.
* Set a couple of environment variables to configure the gRPC library:
  * `GRPC_XDS_BOOTSTRAP` is the path to the bootstrap file the agent generates.
  * `GRPC_XDS_EXPERIMENTAL_SECURITY_SUPPORT` allows actually utilizing security config for mTLS configured by `istiod`.

Your application should have the following annotations to work easily with gRPC+xDS:

https://github.com/stevenctl/grpc-xds-sample/blob/092b8406e45586d61140ead68d139f0a19ce2516/deployment.yaml#L27-L29

The first one sets up the agent and environment variables. The second allows the agent to get everything
ready before we try to reach the control plane. If your application is robust to client failures, or to `server.Serve`
failures, this isn't necessary.

## Deploying the sample

First, create an injection-enabled namespace:

```bash
kubectl create ns xdssample
kubectl label ns xdssample istio-injection=enabled
```

Then simply deploy the example Service and Deployment:

```bash
kubectl -n xdssample apply -f deployment.yaml
```

The application makes a request to the Service every 5 seconds. We can check the logs to make sure it's working:

```bash
kubectl -n xdssample logs $(kubectl -n xdssample get po -ojsonpath='{.items[0].metadata.name}')  
```

They should look something like this:

```text
2021/08/11 22:57:04 Hello, xDS client! From: greeter-748cc9cbff-zlcfm.
2021/08/11 22:57:09 Hello, xDS client! From: greeter-748cc9cbff-8mgmg.
2021/08/11 22:57:14 Hello, xDS client! From: greeter-748cc9cbff-zlcfm.
2021/08/11 22:57:19 Hello, xDS client! From: greeter-748cc9cbff-8mgmg.
2021/08/11 22:57:24 Hello, xDS client! From: greeter-748cc9cbff-zlcfm.
2021/08/11 22:57:29 Hello, xDS client! From: greeter-748cc9cbff-8mgmg.
2021/08/11 22:57:34 Hello, xDS client! From: greeter-748cc9cbff-zlcfm.
```