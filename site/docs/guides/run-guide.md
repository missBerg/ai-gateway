# Monitoring Envoy Readiness in Standalone Mode

When running the AI Gateway in standalone mode, the Envoy process takes some time to initialize and become ready to accept connections. This guide explains how to monitor the Envoy startup process and understand when your gateway is ready to handle traffic.

## Understanding Envoy Readiness States

The AI Gateway CLI provides real-time feedback about Envoy's initialization status through structured logging. During startup, you'll see periodic status updates that indicate the current state of the Envoy process.

### Status States

- **PRE_INITIALIZING**: Envoy is starting up but not yet ready to accept connections
- **LIVE**: Envoy has completed initialization and is ready to handle requests

## Running with Debug Output

To see detailed Envoy readiness information, run the AI Gateway with the `--debug` flag:

```bash
./aigw run --debug
```

## Example Startup Flow

Here's what you can expect to see during a typical startup sequence:

```bash
$ ./aigw run --debug
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

## Key Points

1. **Status Updates**: The CLI checks Envoy's readiness every 2 seconds and logs the current status
2. **Ready State**: Wait for the "Envoy is ready!" message with `status=LIVE` before sending requests
3. **Typical Startup Time**: Envoy initialization usually takes 10-20 seconds depending on your system and configuration
4. **Debug Mode Required**: Readiness monitoring is only available when running with the `--debug` flag

## Troubleshooting

If Envoy remains in `PRE_INITIALIZING` state for an extended period:

1. Check system resources (CPU, memory)
2. Verify your configuration files are valid
3. Look for error messages in the Envoy logs
4. Ensure required ports are available

## Integration with Scripts

You can parse the structured log output in scripts to wait for readiness:

```bash
./aigw run --debug 2>&1 | grep -q "Envoy is ready!"
echo "Gateway is now ready to accept requests"
```

This monitoring capability helps ensure reliable deployments and provides clear feedback during the startup process in standalone mode.