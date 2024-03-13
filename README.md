# logtracr is a log demultiplexer backend

This package implements a log demultiplexer backend compatible with the [logr](https://github.com/go-logr/logr) interface.
Please check [the design document](DESIGN.md) for more details

The package is used mostly in conjunction Kubernetes' [klog](https://github.com/kubernetes/klog) in the kubernetes ecosystem,
but is meant to be generic and working with any logr-based solution.

## Maturity

This is a ALPHA grade implementation.

## LICENSE

Apache v2
