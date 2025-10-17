---
id: high-availability
title: High Availability and Scaling
sidebar_position: 1
---

# High Availability and Scaling

This guide explains how to configure Envoy AI Gateway for high availability and horizontal scaling in production environments. Understanding the controller's architecture is essential for proper scaling configuration.

## AI Gateway Controller Architecture

The AI Gateway controller has two main responsibilities that operate independently:

### Do CRD Reconciliation (Kubernetes Controller)

This portion performs mutable operations on the Kubernetes API server:

- Watches and reconciles AI Gateway Custom Resources (AIGatewayRoute, AIServiceBackend, etc.)
- Creates and manages HTTPRoute, HTTPRouteFilter, and Secret resources
- Manages backend security policies and credential rotation
- **Requires leader election** to prevent split-brain scenarios

### Being an Envoy Gateway Extension Server (gRPC Server)

This portion serves the [Envoy Gateway Extension Server protocol](https://gateway.envoyproxy.io/docs/tasks/extensibility/extension-server/):

- Receives xDS configuration from Envoy Gateway
- Fine-tunes the Envoy configuration for AI-specific features
- Handles requests from all Envoy Gateway controller instances
- **Read-only operations** (no writes to Kubernetes API server)
- **Horizontally scalable** without leader election

## Why Multiple Replicas Matter

The Envoy Gateway controller scales horizontally with multiple replicas to distribute the load of its xDS server component, which communicates with all Envoy proxy instances. Since the AI Gateway controller's extension server is called by Envoy Gateway's xDS translation process, **you should configure multiple replicas of the AI Gateway controller to distribute this load alongside the Envoy Gateway control plane scaling.**.

### Benefits of Multiple Replicas

1. **Load Distribution**: Multiple extension server instances handle requests from Envoy Gateway
2. **High Availability**: If one replica fails, others continue serving
3. **Reduced Latency**: Distributes gRPC connection load across replicas
4. **Resilience**: Better fault tolerance during updates or node failures

### How Leader Election Works

The leader election mechanism applies **only** to the CRD reconciliation portion of the controller:

- Leader election starts **after** the extension server is running. This ensures the extension server is available to handle requests immediately, even before a leader is elected.
- One replica is elected as the leader to perform Kubernetes writes
- Non-leader replicas continue serving the extension server gRPC endpoint
- If the leader fails, another replica is automatically elected

This design ensures that the extension server functionality scales horizontally while maintaining safe, coordinated access to the Kubernetes API server.

## Configuring Multiple Replicas

### Using Helm

Set the replica count in your Helm values:

```yaml
controller:
  replicaCount: 3  # Recommended: 2-3 replicas minimum for production

  # Leader election is enabled by default
  leaderElection:
    enabled: true
```

Install or upgrade with the custom values:

```bash
helm upgrade -i aieg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set controller.replicaCount=3
```

### Verifying Replica Status

Check that all replicas are running:

```bash
kubectl get pods -n envoy-ai-gateway-system -l app.kubernetes.io/name=ai-gateway-controller
```

Expected output:

```
NAME                                     READY   STATUS    RESTARTS   AGE
ai-gateway-controller-5d7c8f9b4c-7xqwm   1/1     Running   0          5m
ai-gateway-controller-5d7c8f9b4c-kj8qs   1/1     Running   0          5m
ai-gateway-controller-5d7c8f9b4c-tn2wp   1/1     Running   0          5m
```

Check the leader election status in the logs:

```bash
kubectl logs -n envoy-ai-gateway-system deployment/ai-gateway-controller | grep leader
```

You should see one replica reporting as the leader.

Note: This command shows logs from one replica. To check all replicas, use:

```bash
kubectl logs -n envoy-ai-gateway-system -l app.kubernetes.io/name=ai-gateway-controller | grep leader
```

## Resource Configuration

Properly sizing your controller resources ensures stable performance under load.

### Basic Resource Configuration

```yaml
controller:
  replicaCount: 3
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi
```

### Production Resource Configuration

For larger deployments with many routes and backends:

```yaml
controller:
  replicaCount: 3
  resources:
    requests:
      cpu: 200m
      memory: 512Mi
    limits:
      cpu: 1000m
      memory: 1Gi
```

### Sizing Guidelines

Consider these factors when sizing resources:

- **Number of AIGatewayRoutes**: More routes require more memory for caching
- **Number of Envoy proxy instances**: More proxies mean more extension server requests
- **Frequency of configuration changes**: More reconciliation activity requires more CPU
- **Number of backends with credential rotation**: Active rotation increases CPU usage

## Horizontal Pod Autoscaling (HPA)

Enable automatic scaling based on CPU or memory utilization:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ai-gateway-controller
  namespace: envoy-ai-gateway-system
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ai-gateway-controller
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

Apply the HPA:

```bash
kubectl apply -f ai-gateway-controller-hpa.yaml
```

## Envoy Gateway Controller Scaling

The Envoy Gateway controller should also be scaled appropriately, as it distributes load across AI Gateway controller replicas. Envoy Gateway uses standard Kubernetes Service load balancing to distribute extension server requests across AI Gateway controller replicas.

Recommended Envoy Gateway scaling:

```bash
kubectl scale deployment envoy-gateway -n envoy-gateway-system --replicas=3
```

Or configure it in your Envoy Gateway Helm values:

```yaml
deployment:
  replicas: 3
```

## High Availability Best Practices

### 1. Pod Disruption Budgets

Ensure availability during node maintenance:

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: ai-gateway-controller
  namespace: envoy-ai-gateway-system
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: ai-gateway-controller
```

### 2. Pod Anti-Affinity

Distribute replicas across different nodes:

```yaml
controller:
  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchLabels:
              app.kubernetes.io/name: ai-gateway-controller
          topologyKey: kubernetes.io/hostname
```

### 3. Readiness and Liveness Probes

The Helm chart includes default probes. Verify they're configured:

```bash
kubectl describe pod -n envoy-ai-gateway-system -l app.kubernetes.io/name=ai-gateway-controller | grep -A 5 Liveness
```

## Monitoring and Metrics

Monitor these key metrics for your controller replicas:

- **CPU and Memory Usage**: Track resource consumption
- **Extension Server Request Rate**: Monitor gRPC requests from Envoy Gateway
- **Reconciliation Duration**: Track how long it takes to process changes
- **Leader Election Events**: Monitor leader transitions

Access controller metrics:

```bash
kubectl port-forward -n envoy-ai-gateway-system deployment/ai-gateway-controller 9090:9090
curl http://localhost:9090/metrics
```

See the [Observability documentation](../capabilities/observability/) for detailed monitoring setup.

## Troubleshooting

### Replicas Not Starting

Check pod status and logs:

```bash
kubectl describe pod -n envoy-ai-gateway-system <pod-name>
kubectl logs -n envoy-ai-gateway-system <pod-name>
```

### Multiple Leaders Elected

This should not happen if leader election is enabled. Check the configuration:

```bash
kubectl get deployment ai-gateway-controller -n envoy-ai-gateway-system -o yaml | grep -A 5 "enable-leader-election"
```

### High CPU Usage

If CPU usage is consistently high:

1. Check the number of AIGatewayRoutes and backends
2. Increase resource limits
3. Add more replicas
4. Review reconciliation frequency in logs

## Summary

For production deployments:

- ✅ **Configure at least 2-3 controller replicas**
- ✅ **Keep leader election enabled** (default: true)
- ✅ **Set appropriate resource requests and limits**
- ✅ **Consider HPA for dynamic scaling**
- ✅ **Scale Envoy Gateway controller similarly**
- ✅ **Configure Pod Disruption Budgets**
- ✅ **Use pod anti-affinity rules**

The AI Gateway controller's extension server scales horizontally to handle load from Envoy Gateway, while leader election ensures safe coordination of Kubernetes API operations.

## Next Steps

- Explore [Deployment Architectures](./deployment-architectures.md) for multi-tier patterns
- Review [Production Best Practices](./production-best-practices.md) for comprehensive operational guidance
- See [Control Plane Architecture](../concepts/architecture/control-plane.md) for deeper architectural understanding
