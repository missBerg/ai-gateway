# AIGatewayRoute Deletion Handling Migration Guide

## Overview

This migration guide covers changes to how the AI Gateway controller handles AIGatewayRoute deletion during external processor secret reconciliation, introduced in commit `2bddd76f`.

## What Changed

The controller now includes a safety check to skip AIGatewayRoutes that are being deleted (have a `DeletionTimestamp`) during external processor secret updates. This prevents stale configuration and improves reliability during route deletion scenarios.

## Background

Previously, when an AIGatewayRoute was being deleted, the controller would still attempt to process it during secret reconciliation, which could lead to:

- Stale or incorrect configuration in the external processor secret
- Potential conflicts with the deletion process
- Unnecessary processing of routes that are no longer active

## Implementation Details

### New Deletion Check

The controller now performs a deletion timestamp check before processing each AIGatewayRoute:

```go
if !aiGatewayRoutes[i].GetDeletionTimestamp().IsZero() {
    c.logger.Info("AIGatewayRoute is being deleted, skipping extproc secret update", 
                  "namespace", aiGatewayRoutes[i].Namespace, "name", aiGatewayRoutes[i].Name)
    continue
}
```

### Behavior Changes

| Route State | Previous Behavior | New Behavior |
|-------------|------------------|--------------|
| Active routes (no deletion timestamp) | Processed normally | ✅ Processed normally |
| Deleted routes (has deletion timestamp) | Processed, potentially causing issues | ✅ Skipped with informative log message |

## Impact on Users

### Positive Changes

- **Prevents stale configuration**: Deleted routes won't contaminate active secret data
- **Improves reliability**: More robust handling during route deletion scenarios
- **Better observability**: Clear logging when routes are skipped
- **Zero breaking changes**: Fully backward compatible

### Log Messages

When a route with a deletion timestamp is encountered, you'll see log messages like:

```
AIGatewayRoute is being deleted, skipping extproc secret update namespace=default name=my-route
```

## Migration Steps

**No migration steps required** - this change is fully backward compatible and automatically improves the system's behavior without requiring any user action.

## Verification

To verify the fix is working correctly:

1. Create an AIGatewayRoute
2. Monitor the external processor secret to ensure it contains the route's configuration
3. Delete the AIGatewayRoute
4. Check logs for the skip message
5. Verify the external processor secret no longer contains the deleted route's configuration

## Troubleshooting

If you notice issues with external processor configuration after upgrading:

1. Check controller logs for deletion skip messages
2. Verify that only active routes appear in the external processor secret
3. Ensure deleted routes are properly removed from configuration

This change ensures the external processor configuration remains consistent and accurate, even during route deletion operations.