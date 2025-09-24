# Migration Guide: Deleted Route Handling in External Processor Secret Reconciliation

## Overview

Starting with commit `2bddd76f`, the AI Gateway controller now includes enhanced handling for deleted AIGatewayRoutes during external processor secret reconciliation. This change improves system reliability by preventing stale configuration data from deleted routes.

## What Changed

### Before
- The controller would process all AIGatewayRoutes during external processor secret reconciliation, including routes that were being deleted
- This could lead to stale or incorrect configuration in the external processor secret
- Potential conflicts could occur during the deletion process

### After
- The controller now checks for deletion timestamps on AIGatewayRoutes
- Routes with deletion timestamps are automatically skipped during reconciliation
- Only active routes are processed and included in the external processor secret configuration

## Technical Details

### New Deletion Check Logic

The controller now includes this safety check in the `reconcileFilterConfigSecret` function:

```go
if !aiGatewayRoutes[i].GetDeletionTimestamp().IsZero() {
    c.logger.Info("AIGatewayRoute is being deleted, skipping extproc secret update", 
                  "namespace", aiGatewayRoutes[i].Namespace, "name", aiGatewayRoutes[i].Name)
    continue
}
```

### Behavior Changes

| Route State | Previous Behavior | New Behavior |
|-------------|-------------------|--------------|
| Active routes (no deletion timestamp) | Processed normally | Processed normally ✅ |
| Deleted routes (with deletion timestamp) | Processed normally | Skipped with log message ✅ |

## Migration Impact

### For Existing Deployments

✅ **No Breaking Changes**: This is a fully backward-compatible enhancement

✅ **Automatic Benefits**: Existing deployments will automatically benefit from more robust route deletion handling

✅ **No Configuration Changes Required**: No updates needed to existing AIGatewayRoute or configuration files

### Expected Improvements

1. **Cleaner Configuration**: External processor secrets will no longer contain stale data from deleted routes
2. **Better Reliability**: Reduced potential for conflicts during route deletion scenarios  
3. **Enhanced Observability**: Clear log messages when routes are skipped during deletion

## Monitoring and Logging

### New Log Messages

When a route with a deletion timestamp is encountered, you'll see log entries like:

```
INFO AIGatewayRoute is being deleted, skipping extproc secret update
  namespace=default name=my-deleted-route
```

### What to Monitor

- Check logs for skipped route messages during route deletion operations
- Verify that external processor secret contents only include active routes
- Monitor for any unexpected behavior during route lifecycle operations

## Troubleshooting

### Common Questions

**Q: Will this affect my existing routes?**  
A: No, only routes that are actively being deleted (have a deletion timestamp) are affected.

**Q: Do I need to update my AIGatewayRoute configurations?**  
A: No, this change is entirely internal to the controller logic.

**Q: What if I see routes being skipped unexpectedly?**  
A: Check if the route has a deletion timestamp set. This typically means a deletion operation is in progress.

### Verification Steps

To verify the new behavior is working correctly:

1. Create an AIGatewayRoute
2. Check that it appears in the external processor secret
3. Delete the route
4. Verify the route is removed from the secret and skip messages appear in logs

## Related Information

- **Issue**: [#1193](https://github.com/envoyproxy/ai-gateway/issues/1193)
- **Pull Request**: [#1196](https://github.com/envoyproxy/ai-gateway/pull/1196)
- **Commit**: `2bddd76fd7ebe0fb973164b8d015faef5157adec`

This enhancement is part of ongoing improvements to make the AI Gateway more robust and reliable during resource lifecycle operations.