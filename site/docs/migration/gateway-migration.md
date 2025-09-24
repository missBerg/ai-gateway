# Migration Guide: AIGatewayRoute Deletion Handling

## Overview

Starting with commit `2bddd76f`, the AI Gateway controller now includes improved handling for AIGatewayRoutes that are being deleted during external processor secret reconciliation. This change enhances the reliability and consistency of the external processor configuration by preventing deleted routes from contaminating active secret data.

## What Changed

The `reconcileFilterConfigSecret` function in the gateway controller now includes a deletion timestamp check to skip AIGatewayRoutes that are in the process of being deleted.

### Before

Previously, when an AIGatewayRoute was being deleted (had a `DeletionTimestamp`), the controller would still attempt to process it during secret reconciliation, potentially leading to:

- Stale or incorrect configuration in the external processor secret
- Potential conflicts with the deletion process  
- Unnecessary processing of routes that are no longer active

### After

The controller now checks for deletion timestamps before processing each AIGatewayRoute:

```go
if !aiGatewayRoutes[i].GetDeletionTimestamp().IsZero() {
    c.logger.Info("AIGatewayRoute is being deleted, skipping extproc secret update", 
                  "namespace", aiGatewayRoutes[i].Namespace, "name", aiGatewayRoutes[i].Name)
    continue
}
```

## Impact on Your Deployment

This change is **fully backward compatible** and requires no action from users. The improvement provides:

- **Enhanced Reliability**: More robust handling during route deletion scenarios
- **Cleaner Configuration**: Prevents stale configuration from deleted routes
- **Better Observability**: Clear logging when routes are skipped during deletion
- **Consistent State**: External processor configuration remains accurate even during route deletions

## Behavior Changes

| Route State | Previous Behavior | New Behavior |
|-------------|------------------|--------------|
| Active routes (no deletion timestamp) | Processed normally | Processed normally ✅ |
| Deleting routes (has deletion timestamp) | Still processed | Skipped with log message ✅ |

## Log Messages

When a route is skipped due to deletion, you will see log messages like:

```
AIGatewayRoute is being deleted, skipping extproc secret update namespace=default name=my-route
```

These informational messages help with debugging and monitoring the deletion process.

## Migration Steps

**No migration steps are required.** This change is automatically applied when you upgrade to a version containing this fix.

## Verification

To verify the fix is working correctly:

1. Create an AIGatewayRoute
2. Monitor the external processor secret to confirm it includes the route configuration
3. Delete the AIGatewayRoute
4. Check the logs for skip messages during the deletion process
5. Verify the external processor secret no longer contains configuration for the deleted route

## Related Issues

This change resolves issue #1193 regarding stale configuration during route deletion operations.