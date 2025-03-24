# A flight recorder for logs
Owner: fromani@redhat.com

## Summary
Logging is not a solved problem in a complex system, especially in a complex distributed system.
Focusing on the kubernetes ecosystem, the most common experienced pain points are excessive or
insufficient verbosiness, which in turn creates the need to change the verbosiness level during
the component lifetime.
This is because keeping the verbosiness high will create a large amount of logs, while keeping
it low will make it way harder to troubleshoot an issue without increase the verbosiness before,
restarting the affected components and re-create the issue, which can take time and effort.

The component logs are affected by all these issues. Keeping the log level high is, as it stands
today (March 2024), still discouraged and impractical. The matter is further complicated by the
some compontents may require post-fact review or audit of their behaviour, regardless of how
comprenshive and solid their testsuite it is.

This is because some component decision may depend on the request and on the cluster state,
which can be impossible to predict and/or mock.
In the best case scenarios, the cluster state is mockable only after it is known, which
severely limits the investigation scope. So it is especially important to have enough
data to troubleshoot issue, which once again calls for high verbosiness.

We would like to improve the current flow, which is basically keep verbosiness=2, and in case
of incidents (but note: always after the fact), bump the verbosiness to 4 or more,
reproduce again, send logs.

## Motivation: flight recording for logs

A great and concise definition of "flight recording" comes from the [go blog](https://go.dev/blog/execution-traces-2024):
```
Suppose you work on a web service and an RPC took a very long time.
You couldn’t start tracing at the point you knew the RPC was already taking a while, because the root cause of the slow request already happened and wasn’t recorded.
[...]
The insight with flight recording is to have tracing on continuously and always keep the most recent trace data around, just in case.
Then, once something interesting happens, the program can just write out whatever it has!
```

Some platforms enable "flight recording" for events and traces, with focus often times for performance
analysis. Notable examples are the aforementioned [golang](https://go.dev/blog/execution-traces-2024) and the [JDK](https://en.wikipedia.org/wiki/JDK_Flight_Recorder).

The flight recording pattern is a very powerful tool for events and traces; it is mostly used for performance
analysis. Oftentimes, though something similar to analyze the business logic behavior would be very useful.
Is worth highlighting that flight recording is not meant by any means as replacement for tests, but rather as complement.
For any software, a good testsuite is just a prerequisite.

This package aims to provide a flight recorder (-alike) functionality for logs.

## Goals
- Make it possible/easier to correlate all the logs pertaining to a code flow
- Improve the signal-to-noise ratio of the logs, filtering out all the content not relevant the specific use case
- Avoid excessive storage consumption
- Minimize the slowdown caused by logging
- Be compatible with the go-logr API for easy integration in a codebase

## Non-Goals
- Change the logging system (e.g. migrate away from go-logr/klog)
- Introduce a replacement logger (e.g. module with klog-like API)
- Break source code compatibility (e.g. no changes to the hosting code base)
- Move to traces (independent effort not mutually exclusive)
- Make the verbosiness tunable at runtime (no consensus about how to do securely and safely,
  will require a new logging package)

## Proposal
- Introduce and use extensively a logID key/value pair to enable correlation of all the log entries
  pertaining to a specific code flow
- Introduce a new logging backend plugging into the logr framework, which klog >= 2.100 fully supports
- Let the new logging backend handle the logging demultiplexing
  - Aggregate the logs per-object

## Risks and Mitigations

- injecting a new logging mechanism which overrides the main setting can introduce bugs which can lead
  to lost logs or process crashing. The mitigation is to make sure that all the new code is completely isolated
  and can be completely and easily switched off and disabled.
- overriding the verbosity setting and storing extra logs on disk can cause storage exhaustion.
  The mitigation is to set strong upper bounds on the extra storage and implementing a builtin log rotation mechanism.
- injecting a new logging means doing all the logging twice, which has a performance cost.
  The mitigation is optimizing the code path, keeping the extra log storage in volatile memory for speed reasons
  (tmpfs) but ultimately the proposed approach WILL still have a performance impact.

## Design details

### mapping concepts to OpenTelemetry

Where feasible, we try to mimic as close as possible, or plainly to adopt, the same concepts and terminology
as [OpenTelemetry](https://opentelemetry.io/docs/concepts/signals/traces/) (otel).

### custom logging backend

First and foremost, we need to filter the logs using a different criteria. Not just the global verbosity anymore,
but using a combination of global verbosity and flow matching. In other words we want to introduce two log sinks
1. standard log sink: all the logs from all the flows are filtered by verbosity
2. NUMA-specific log sink: all the logs from all the logs are filtered by verbosity and some flowID, and grouped by the same flowID.
A `flowID` would be roughly equivalent to a otel `trace_id`.

A "flow ID" is any token which is unique per-flow and per-object. For example, every code flow should
have its own uinque flowID. The value of flowID is not relevant.

The log packages accepts a Logger implementation that will be used as backing implementation of the traditional klog log calls
using the `stdr` interface. See  `klog.SetLogger` for details. Thus, we can keep the usual klog interface without changes,
but inject out custom logic cleanly.

Note however that klog is doing the verbosity check, so we need first and foremost to bump the default klog verbosity
so klog will pass all the messages to our backend, and then do the real filtering in our backend.
Unfortunately, client-go has some klog calls backed in, so the klog verbosity is better kept lower than 6.
In order to convey all the settings to our logging backend, a separate config option set is needed.
We initially use environment variables, but configuration files or new command line flag is also an option.
Environment variables or configuration files are the most backward compatible ways, because the rest of the process
can just safely ignore them.

```
+------------------+
| Application code |
+------------------+
         |
         | calls
         V
+-----------------+
| klog (frontend) |<-- -v=MAX (overridden) gather all logs, passingh through the verbosiness check
+-----------------+
         |
         | calls
         V
+-----------------+
|logtracr(backend)|<-- verbosiness value copied from -v - real verbosiness filter
+-----------------+
         |               +------------------+
         +-------------->| groups by flowID |<- sends to different storage
         V               +------------------+
  /---------------/         |             |
 /    main log   /          |             |
/---------------/           V             V
                   /--------------/  /--------------/
                  / flowID=A log /  / flowID=X log /
                 /--------------/  /--------------/
```

### grouping log entries by flowID

Once all the logs reach our backend, we can filter them using the above criteria (real user-provided verbosiness,
and flowID). Note the configuration agent (numaresources-operator for example) is in charge of:
1. forcing the klog verbosiness (-`v`) to the maximum allowed (typically 5)
2. move the real user-specified verbosiness setting into the backend-specific config setting.

In order to determine `flowID` the logging backend needs some cooperation from the application layer.
The application layer is already using a `logID`/`xxx` key/value pair consistently across all the logs
for the hosting codebase. This was an independent decision predating the introduction of the custom logging.
Having a consistent key enables to easily grep the general log to fetch all the messages pertaining a flow,
so it's a good addition (and completely transparent to the application logic) anyway.

The custom logging backend can leverage `logID` as `flowID`. We hook in the `InfoS` and `ErrorS` implementation
(which a `stdr`-conformant logging backend must implement anyway) and if one of the key/value pairs is `logID`,
the backend marks the log entry as belonging to the flow matching the `logID` value.
For example, the line
```
I0312 10:36:40.113054       1 filter.go:155] "final verdict" logID="numalign-pod-sdmxg/numalign" node="node01.lab.eng.k8s.lan" suitable=true
```
would belong to the flowID `numalign-pod-sdmxg/numalign`.
Entries extracted this way and sent to further processing would be the rough equivalent of otel's `span`s.

### multiple storage sinks

Once we can bind log entries to opaque `flowID` we can group them and send to multiple sinks.
The logging backed bifurcates: log entries are sent to the general logging according to the user-specified verbosiness rules,
and they are _also_ sent to the per-flowID storage.

The per-flowID storage is a collection of memory buffers to which log-lines are appended to.
The memory buffers are monitored for both their number and individual sizes to prevent excessive memory usage.
A last-updated timestamp is recorded for each memory buffer. If a buffer grows too much, if there are too many active
buffers or if a buffer is not updated over a configurable threshold, the buffers are sent to durable storage and the
memory is freed.

This aggregation of entries (otel `span`s) are aggregated in the rough equivalent of an otel `trace`.

### per-flow logging storage

In order to minimize the slowdown caused by the extra work and the extra writes the custom backend is performing,
the recommended set for the durable storage is actually not durable, but rather a tmpfs.
Another key benefit of the tmpfs as storage backend is that we can trivially set a hard cap for the storage size,
preventing storage space exhaustion. This is especially important to minimize the side effects and contain the
blast radius of the logging backend failures.

Since the per-flow storage space is bound, the logging backend must tolerate (and ignore) write failures,
because this is an expected, albeit unlikely, occurrence.
Is worth stressing out that the per-flow storage augments, not replaces, the main log, so augmentation failures
are not critical

### future extensions

#### builtin log rotation
We can add a simple log rotation mechanism in the logging backend. Log entries older than a configurable threshold
can be deleted to make room for the new entries. It's often preferrables to have newer entries rather than old ones.

## Discarded alternatives

### use an off-the-shelf 3rd party logging package

At moment of writing, we are not aware of a logging package which easily enables flight recording and that fits
in the go-logr ecosystem. The search is not done though. If a comparable alternative is found, we will run an evaluation
and a case study about gaps and differences with this solution.

### just replace the core logging package

This is among the non-goals; besides, this is specifically a non-goal unless the replacement is 1:1 source code API compatible,
and per previous research there is not such option.

