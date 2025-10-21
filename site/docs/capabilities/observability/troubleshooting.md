---
id: troubleshooting
title: Troubleshooting and Debugging
sidebar_position: 9
---

Envoy AI Gateway provides built-in tools to help troubleshoot performance issues and debug problems. This guide covers the available debugging capabilities and how to use them effectively.

## Profiling with pprof

Envoy AI Gateway's extproc component has a built-in profiling server that uses Go's standard `pprof` package. This allows you to collect CPU profiles, memory profiles, and other debugging information that can help diagnose performance issues.

### How pprof Works

The pprof server exposes several HTTP endpoints that allow you to:

- Collect CPU profiles to identify CPU-intensive functions
- Examine memory allocations and detect memory leaks
- View running goroutines and their stack traces
- Generate execution traces for detailed analysis

### Using pprof

The pprof server is **enabled by default** on port 6060, but can be disabled using the `DISABLE_PPROF` environment variable.

To access the pprof server:

1. Port-forward the extproc pod to your local machine:

```bash
# Find the extproc pod name
kubectl get pods -n envoy-ai-gateway-system -l app.kubernetes.io/component=ai-gateway-extproc

# Port-forward the pod (replace POD_NAME with actual pod name)
kubectl port-forward -n envoy-ai-gateway-system POD_NAME 6060:6060
```

2. Access the pprof endpoints:

- Web UI: http://localhost:6060/debug/pprof/
- CPU Profile: http://localhost:6060/debug/pprof/profile
- Heap Profile: http://localhost:6060/debug/pprof/heap
- Goroutines: http://localhost:6060/debug/pprof/goroutine
- Execution Trace: http://localhost:6060/debug/pprof/trace

### Disabling pprof

For security-sensitive environments, you may want to disable the pprof server. This can be done by setting the `DISABLE_PPROF` environment variable to any value when deploying the AI Gateway:

```yaml
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --reuse-values \
  --set "extProc.extraEnvVars[0].name=DISABLE_PPROF" \
  --set "extProc.extraEnvVars[0].value=true"
```

Or by adding the following to your Helm values file:

```yaml
extProc:
  extraEnvVars:
    - name: DISABLE_PPROF
      value: "true"
```

### Using pprof Tools

For more advanced analysis, you can use the Go pprof tools:

```bash
# Install Go tools if needed
go install github.com/google/pprof@latest

# Collect and analyze a 30-second CPU profile
curl -s http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
pprof -http=:8080 cpu.prof

# Analyze memory allocations
curl -s http://localhost:6060/debug/pprof/heap > heap.prof
pprof -http=:8080 heap.prof
```

## When to Use Profiling

Consider using pprof in the following scenarios:

- When the AI Gateway is experiencing high latency
- When memory usage is growing unexpectedly
- During load testing to identify bottlenecks
- To troubleshoot performance regressions

## Best Practices

- **Security**: The pprof endpoints expose internal details of the application. In production environments, ensure these endpoints are not publicly accessible.
- **Performance Impact**: While the impact is minimal when not in use, collecting profiles (especially traces) can have some overhead. Avoid running long profiling sessions in production during peak times.
- **Collaboration**: When reporting performance issues to the AI Gateway team, including pprof profiles can be extremely helpful for diagnosing problems.

## Additional Troubleshooting Resources

For more information about troubleshooting AI Gateway, refer to:

- [AI/LLM Metrics](./metrics.md) - Monitor system performance through metrics
- [GenAI Distributed Tracing](./tracing.md) - Track requests through the system
- [Access Logs](./accesslogs.md) - Analyze request and response patterns

## See Also

- [Go pprof Documentation](https://golang.org/pkg/net/http/pprof/)
- [Profiling Go Applications](https://go.dev/blog/pprof)