# grpc-xds-sample

In this repo:
* Sample Go gRPC application, with code to use xDS features:
  * [Client-side configuration](https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md)
  * [Server-side configuration](https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md)
  * [Other gRFCs](https://github.com/grpc/proposal)
* A sample Kubernetes deployment compatible with [Istio](https://github.com/istio/isttio).

This is intended mostly as a guide for how to make your own gRPC applications compatible with xDS.

## As a guide for writing your own applications

// TODO fill in with permalinks

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