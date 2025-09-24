# Understanding Envoy Readiness in Standalone Mode

When running the AI Gateway in standalone mode, the Envoy proxy process needs time to initialize before it can accept connections. This guide explains the readiness feedback feature that helps you understand when your AI Gateway is ready to handle requests.

## Overview

In standalone mode, the AI Gateway CLI provides real-time feedback about the Envoy startup process. This eliminates guesswork about when your gateway is ready to receive traffic and helps with debugging startup issues.

## Envoy Startup States

The AI Gateway monitors and reports the following Envoy states during startup:

- **PRE_INITIALIZING**: Envoy is starting up and not yet ready to accept connections
- **LIVE**: Envoy has completed initialization and is ready to handle requests

## Example Output

When you start the AI Gateway in standalone mode with debug logging, you'll see output like this:

```bash
$ ./aigw run --debug
(...)
looking up the latest patch for Envoy version 1.35
1.35.3 is already downloaded
starting: /tmp/envoy-gateway/versions/1.35.3/bin/envoy in run directory /tmp/envoy-gateway/runs/1758020241848418000
[2025-09-16 12:57:22.144][18234916][warning][config] [source/server/options_impl_platform_default.cc:9] CPU number provided by HW thread count (instead of cpuset).
time=2025-09-16T12:57:22.781+02:00 level=INFO msg="Waiting for Envoy to be ready..." status=PRE_INITIALIZING
time=2025-09-16T12:57:24.781+02:00 level=INFO msg="Waiting for Envoy to be ready..." status=PRE_INITIALIZING
time=2025-09-16T12:57:26.781+02:00 level=INFO msg="Waiting for Envoy to be ready..." status=PRE_INITIALIZING
time=2025-09-16T12:57:28.781+02:00 level=INFO msg="Waiting for Envoy to be ready..." status=PRE_INITIALIZING
time=2025-09-16T12:57:30.781+02:00 level=INFO msg="Waiting for Envoy to be ready..." status=PRE_INITIALIZING
time=2025-09-16T12:57:32.781+02:00 level=INFO msg="Waiting for Envoy to be ready..." status=PRE_INITIALIZING
time=2025-09-16T12:57:34.781+02:00 level=INFO msg="Waiting for Envoy to be ready..." status=PRE_INITIALIZING
time=2025-09-16T12:57:36.780+02:00 level=INFO msg="Waiting for Envoy to be ready..." status=PRE_INITIALIZING
time=2025-09-16T12:57:38.782+02:00 level=INFO msg="Envoy is ready!" status=LIVE
```

## Key Features

### Periodic Status Updates

The AI Gateway checks Envoy's readiness every 2 seconds and logs the current status. This provides continuous feedback during the startup process.

### Clear Ready Signal

When Envoy transitions to the `LIVE` state, you'll see the message "Envoy is ready!" indicating that your AI Gateway is now ready to accept and process requests.

### Debug Visibility

The readiness monitoring is particularly useful when running with the `--debug` flag, as it provides detailed insight into the startup process.

## Best Practices

### Wait for Ready State

Always wait for the "Envoy is ready!" message before sending requests to your AI Gateway. Sending requests before Envoy is ready will result in connection errors.

### Monitor Startup Time

If you notice consistently long startup times, this may indicate:
- Resource constraints on your system
- Network issues affecting Envoy initialization
- Configuration problems

### Troubleshooting Startup Issues

If Envoy remains in `PRE_INITIALIZING` state for an extended period:

1. Check system resources (CPU, memory)
2. Verify network connectivity
3. Review Envoy logs for specific error messages
4. Ensure proper configuration files are present

## Integration with Automation

When integrating the AI Gateway into automated deployments or scripts, you can:

1. Parse the log output to detect the "Envoy is ready!" message
2. Use the status information to implement proper health checks
3. Build startup timeouts based on expected initialization times

This readiness feedback makes the AI Gateway more suitable for production deployments where reliable startup detection is crucial.