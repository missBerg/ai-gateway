# Migration Guide: External Processor Secret Reconciliation

## Overview

This migration guide covers changes to the external processor secret reconciliation behavior introduced in commit `2bddd76f`. The change adds safety checks to prevent deleted AIGatewayRoutes from being processed during secret updates.

## What Changed

The `reconcileFilterConfigSecret` function in the gateway controller now includes a deletion timestamp check to skip AIGatewayRoutes that are in the process of being deleted.

### Before

Previously, the controller would process all AIGatewayRoutes during secret reconciliation, including those marked for deletion, which could lead to:
- Stale configuration in external processor secrets
- Potential conflicts during the deletion process
- Unnecessary processing of inactive routes

### After

The controller now checks for deletion timestamps and skips routes that are being deleted:

```go
if !aiGatewayRoutes[i].GetDeletionTimestamp().IsZero() {
    c.logger.Info("AIGatewayRoute is being deleted, skipping extproc secret update", 
                  "namespace", aiGatewayRoutes[i].Namespace, "name", aiGatewayRoutes[i].Name)
    continue
}
```

## Impact on Users

### No Breaking Changes

This is a **backward-compatible** change that improves the reliability of external processor secret reconciliation without requiring any modifications to existing configurations or workflows.

### Behavioral Changes

1. **Active Routes**: Continue to be processed normally during secret reconciliation
2. **Deleted Routes**: Are now automatically skipped with informative log messages
3. **Secret Content**: External processor secrets will only contain configuration data from active routes

## What You Need to Do

**Nothing!** This change is fully automatic and requires no action from users. The improvement is transparent and enhances the system's robustness.

## Monitoring and Observability

### Log Messages

When routes are skipped due to deletion, you'll see log entries like:

```
AIGatewayRoute is being deleted, skipping extproc secret update namespace=default name=my-route
```

These messages are informational and indicate the system is working correctly.

### Expected Behavior

- External processor secrets will be more accurate and contain only active route configurations
- Reduced potential for configuration conflicts during route deletion
- Improved system reliability during route lifecycle management

## Benefits

✅ **Prevents Stale Configuration**: Deleted routes won't contaminate active secret data  
✅ **Improves Reliability**: More robust handling during route deletion scenarios  
✅ **Better Observability**: Clear logging when routes are skipped  
✅ **Zero Breaking Changes**: Fully backward compatible  

## Related Resources

- [Issue #1193](https://github.com/envoyproxy/ai-gateway/issues/1193)
- [Pull Request #1196](https://github.com/envoyproxy/ai-gateway/pull/1196)
- [AIGatewayRoute Documentation](/docs/concepts/resources#aigatewayroute)