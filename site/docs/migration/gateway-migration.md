# Migration Guide: External Processor Secret Reconciliation Updates

## Overview

This migration guide covers changes to the external processor secret reconciliation behavior introduced in version 0.4.0. The changes improve reliability by preventing deleted AIGatewayRoutes from being processed during secret updates.

## What Changed

### Background

Previously, when an AIGatewayRoute was being deleted (marked with a DeletionTimestamp), the controller would still attempt to process it during external processor secret reconciliation. This could lead to:

- Stale or incorrect configuration in the external processor secret
- Potential conflicts with the deletion process
- Unnecessary processing of routes that are no longer active

### New Behavior

The controller now includes a deletion timestamp check before processing each AIGatewayRoute during secret reconciliation:

```go
if !aiGatewayRoutes[i].GetDeletionTimestamp().IsZero() {
    c.logger.Info("AIGatewayRoute is being deleted, skipping extproc secret update", 
                  "namespace", aiGatewayRoutes[i].Namespace, "name", aiGatewayRoutes[i].Name)
    continue
}
```

## Impact on Your Setup

### No Breaking Changes

This change is fully backward compatible. No configuration updates are required.

### Behavior Changes

- **Active routes**: Continue to be processed normally ✅
- **Routes being deleted**: Now skipped with informative log messages ✅
- **External processor secrets**: Will only contain data from active routes

### Log Message Changes

You may notice new informational log messages when AIGatewayRoutes are being deleted:

```
AIGatewayRoute is being deleted, skipping extproc secret update namespace=default name=my-route
```

This is expected behavior and indicates the system is working correctly.

## Benefits

### Improved Reliability

- **Prevents stale configuration**: Deleted routes no longer contaminate active secret data
- **More robust deletion handling**: Better behavior during route deletion scenarios
- **Consistent state**: External processor configuration remains accurate

### Better Observability

- Clear logging when routes are skipped during deletion
- Easier troubleshooting of route lifecycle issues

## Verification

To verify the new behavior is working correctly:

1. **Monitor logs** when deleting AIGatewayRoutes for the skip messages
2. **Check external processor secrets** to ensure they only contain active route data
3. **Verify route deletion** completes without configuration conflicts

## No Action Required

This improvement is automatically applied when upgrading to version 0.4.0 or later. No configuration changes or manual intervention are required.

The change enhances the system's robustness while maintaining full backward compatibility with existing setups.