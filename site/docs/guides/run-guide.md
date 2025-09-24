# CLI Run Mode: Understanding Envoy Readiness Status

This guide explains how the AI Gateway CLI provides feedback on Envoy's startup process when running in standalone mode.

## Overview

When using the `aigw run` command in standalone mode, the Envoy proxy process needs time to initialize before it can accept connections. The CLI now provides real-time status updates to help you understand when Envoy is ready to handle requests.

## Status Messages

The CLI monitors Envoy's readiness and displays status messages at regular intervals:

```bash
$ ./aigw run --debug
looking up the latest patch for Envoy version 1.35
1.35.3 is already downloaded
starting: /tmp/envoy-gateway/versions/1.35.3/bin/envoy in run directory /tmp/envoy-gateway/runs/1758020241848418000
time=2025-09-16T12:57:22.781+02:00 level=INFO msg="Waiting for Envoy to be ready..." status=PRE_INITIALIZING
time=2025-09-16T12:57:24.781+02:00 level=INFO msg="Waiting for Envoy to be ready..." status=PRE_INITIALIZING
time=2025-09-16T12:57:38.782+02:00 level=INFO msg="Envoy is ready!" status=LIVE
```

## Status Types

### PRE_INITIALIZING
- **Meaning**: Envoy is starting up but not yet ready to accept connections
- **Action**: Wait for the status to change to LIVE
- **Typical Duration**: 15-30 seconds depending on configuration complexity

### LIVE
- **Meaning**: Envoy has completed initialization and is ready to accept requests
- **Action**: You can now send requests to the AI Gateway
- **What happens next**: The gateway is fully operational

## Using the Status Information

### During Development
Monitor the status messages to:
- Understand when your configuration changes have been applied
- Know when it's safe to send test requests
- Debug startup issues by observing how long initialization takes

### In Scripts
The status messages can help you create robust deployment scripts:

```bash
#!/bin/bash
# Start AI Gateway in background
./aigw run --config my-config.yaml &

# Wait for "Envoy is ready!" message
while ! grep -q "Envoy is ready!" <(tail -f aigw.log); do
    echo "Waiting for AI Gateway to be ready..."
    sleep 1
done

echo "AI Gateway is ready - running tests"
# Your test commands here
```

### Troubleshooting

If Envoy remains in `PRE_INITIALIZING` status for an extended period:

1. **Check Configuration**: Verify your YAML configuration files are valid
2. **Review Logs**: Look for error messages in the debug output
3. **Resource Constraints**: Ensure sufficient system resources (CPU, memory)
4. **Network Issues**: Verify that required ports are available

## Best Practices

1. **Enable Debug Mode**: Use `--debug` flag to see detailed status information
2. **Monitor Logs**: Keep an eye on status messages during development
3. **Automated Deployment**: Use status messages to coordinate deployment scripts
4. **Patience During Startup**: Allow adequate time for Envoy initialization, especially on resource-constrained systems

## Related Commands

- `aigw run --help`: View all available run options
- `aigw translate`: Convert configurations before running
- `aigw run --debug`: Enable detailed logging including readiness status