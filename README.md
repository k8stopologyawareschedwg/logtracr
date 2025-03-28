# Logtracr: a go-logr flight recorder backend

This package implements a log [flight recorder](https://en.wikipedia.org/wiki/JDK_Flight_Recorder) backend compatible with the [logr](https://github.com/go-logr/logr) interface.
Please check [the design document](DESIGN.md) for more details

The package is used mostly in conjunction Kubernetes' [klog](https://github.com/kubernetes/klog) in the kubernetes ecosystem,
but is meant to be generic and working with any logr-based solution.

For more context, please see the [presentation](docs/a-go-logr-de-entanglr.pdf).

## What is a log flight recorder

Let's borrow a great concise description from the [go blog](https://go.dev/blog/execution-traces-2024):
```
Suppose you work on a web service and an RPC took a very long time.
You couldn’t start tracing at the point you knew the RPC was already taking a while, because the root cause of the slow request already happened and wasn’t recorded.
[...]
The insight with flight recording is to have tracing on continuously and always keep the most recent trace data around, just in case.
Then, once something interesting happens, the program can just write out whatever it has!
```

This package wants to offer the same functionality but for logs, to enable (better) review, auditing, support of the program decisions.

## Maturity

This is a ALPHA grade implementation.

## LICENSE

Apache v2
