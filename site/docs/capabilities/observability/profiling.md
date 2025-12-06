---
title: Profiling with pprof
description: Using Go's pprof to profile the AI Gateway
---

# Profiling with pprof

AI Gateway includes built-in support for profiling using Go's [pprof](https://golang.org/pkg/net/http/pprof/) package. This enables operators to diagnose performance issues in production environments with minimal overhead.

## How it works

The extproc component automatically starts a pprof HTTP server on port 6060. This server exposes standard pprof endpoints that can be used to collect CPU profiles, memory profiles, goroutine dumps, and other debugging information.

The pprof server is enabled by default because:

1. It has negligible performance impact when endpoints aren't being accessed
2. It provides critical diagnostic capabilities when needed for troubleshooting
3. It's not exposed outside the pod by default, requiring port-forwarding to access

## Accessing pprof endpoints

To access the pprof endpoints, you'll need to set up port-forwarding to the extproc pod:

```bash
# Get the name of the extproc pod
kubectl get pods -n your-namespace

# Set up port forwarding
kubectl port-forward pods/your-extproc-pod-name 6060:6060 -n your-namespace
```

Once port forwarding is established, you can access the pprof web interface at:
```
http://localhost:6060/debug/pprof/
```

## Available profiling endpoints

The following pprof endpoints are available:

- `/debug/pprof/` - Index page listing all available profiles
- `/debug/pprof/cmdline` - Command-line arguments used to start the process
- `/debug/pprof/profile` - CPU profile (30-second duration by default)
- `/debug/pprof/symbol` - Symbol lookup for program counters
- `/debug/pprof/trace` - Execution trace (1-second duration by default)
- `/debug/pprof/heap` - Memory allocation profile
- `/debug/pprof/goroutine` - Stack traces of all current goroutines
- `/debug/pprof/threadcreate` - Stack traces that led to creation of new OS threads
- `/debug/pprof/block` - Stack traces that led to blocking on synchronization primitives
- `/debug/pprof/mutex` - Stack traces of holders of contended mutexes

## Using the pprof tool

The `go tool pprof` command can be used to analyze profiles collected from these endpoints:

```bash
# Collect and analyze a 30-second CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# Collect and analyze a heap profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

## Disabling pprof

In security-sensitive environments, you may want to disable the pprof server. This can be done by setting the `DISABLE_PPROF` environment variable.

When using Helm, you can disable pprof by adding the following to your values file:

```yaml
extProc:
  extraEnvVars:
    - name: DISABLE_PPROF
      value: "true"
```

## Performance impact

The pprof server has negligible impact on performance when the endpoints are not being accessed. When profiles are being collected, there may be a small performance overhead depending on the type of profile.

For production environments, it's generally safe to leave pprof enabled as it provides valuable diagnostic capabilities when needed and has minimal impact otherwise.