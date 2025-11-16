---
id: profiling
title: Profiling
---

# Profiling

Envoy AI Gateway includes built-in profiling capabilities to help diagnose performance issues in production environments. The gateway uses Go's standard [pprof](https://golang.org/pkg/net/http/pprof/) package to expose runtime profiling data.

## Enabled by Default

The pprof HTTP server is **enabled by default** on port `6060` in the extproc component. This design choice was made because:

- The performance impact is negligible when the endpoints are not being accessed
- It ensures profiling is available when needed for troubleshooting without redeployment
- The endpoints are only accessible from within the cluster or via port forwarding

## Accessing the Profiler

To access the profiling data:

1. Set up port forwarding to the extproc pod:

```bash
kubectl port-forward deployment/envoy-ai-gateway-extproc 6060:6060 -n your-namespace
```

2. Access the profiler using one of these methods:

   - Web UI: Open http://localhost:6060/debug/pprof/ in your browser
   - Command line: Use `go tool pprof` to analyze specific profiles:

```bash
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Available Profiles

The following profiles are available:

- `/debug/pprof/profile`: CPU profile
- `/debug/pprof/heap`: Memory allocation profile
- `/debug/pprof/block`: Blocking profile showing where goroutines block waiting
- `/debug/pprof/goroutine`: Stack traces of all current goroutines
- `/debug/pprof/threadcreate`: Stack traces that led to the creation of new OS threads
- `/debug/pprof/mutex`: Stack traces of holders of contended mutexes

## Disabling the Profiler

In production environments where you prefer not to have the profiler enabled, you can disable it by setting the `DISABLE_PPROF` environment variable.

When using the Helm chart, add the following to your values file:

```yaml
extProc:
  extraEnvVars:
    - name: DISABLE_PPROF
      value: "true"
```

## Security Considerations

The pprof endpoints expose internal runtime details of your application that could be useful to attackers. Consider the following security practices:

- In production environments, ensure these endpoints are not publicly exposed
- Use network policies to restrict access to the profiling port
- When profiling is not needed, disable it via the environment variable
- Consider enabling authentication if exposing these endpoints in sensitive environments