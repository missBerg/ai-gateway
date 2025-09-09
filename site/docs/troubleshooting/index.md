---
id: troubleshooting
title: Troubleshooting
sidebar_position: 6
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Troubleshooting Envoy AI Gateway

This guide helps you find and fix common issues with Envoy AI Gateway on Kubernetes. Follow the steps to diagnose problems one by one.

## Quick Diagnostic Checklist

Before you start, review this quick checklist:

1. ✅ **Installation complete**: Envoy Gateway and Envoy AI Gateway are installed and running.
2. ✅ **Resources applied**: All CRDs (AIGatewayRoute, AIServiceBackend, BackendSecurityPolicy) are applied.
3. ✅ **Pods running**: All pods are Running and Ready.
4. ✅ **Services available**: Gateway services have external IPs.
5. ✅ **Configuration valid**: YAML manifests validate and apply without errors.

## Common Issues

### Installation Problems

#### Pods Not Starting

**Symptoms**: Pods are stuck in `Pending`, `CrashLoopBackOff`, or `ImagePullBackOff`.

<Tabs groupId="pod-issues">
<TabItem value="pending" label="Pending Pods">

Check resources:
```bash
kubectl describe pods -n envoy-ai-gateway-system
kubectl describe pods -n envoy-gateway-system
```

Common causes:
- Not enough cluster resources (CPU or memory).
- Node selector constraints.
- Missing persistent volume claims (PVCs).

</TabItem>
<TabItem value="crashloop" label="CrashLoopBackOff">

Check pod logs:
```bash
kubectl logs -n envoy-ai-gateway-system deployment/ai-gateway-controller
kubectl logs -n envoy-gateway-system deployment/envoy-gateway
```

Common causes:
- Configuration errors.
- Missing dependencies (for example, Redis for rate limiting).
- Permission issues.

</TabItem>
<TabItem value="imagepull" label="ImagePullBackOff">

Verify the image:
```bash
kubectl describe pod <pod-name> -n <namespace>
```

Common causes:
- Network issues.
- Private registry auth problems.
- Wrong image tags.

</TabItem>
</Tabs>

#### CRDs Not Applied

**Symptoms**: Errors about unknown resource types when applying configs.

**Solution**:
```bash
kubectl get crd | grep ai-gateway

helm upgrade -i aieg-crd oci://docker.io/envoyproxy/ai-gateway-crds-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --create-namespace
```

### Connectivity Issues

#### Gateway Not Reachable

**Symptoms**: Requests to the gateway time out or are refused.

**Diagnosis**:
```bash
kubectl get svc -n envoy-gateway-system
kubectl get gateway -n default
kubectl get gatewayclass
kubectl get httproute -A
kubectl get aigatewayroute -A
```

**Common Solutions**:
- Make sure the LoadBalancer service has an external IP.
- Check your cloud provider LoadBalancer settings.
- Review firewall rules and security groups.
- Confirm DNS resolves correctly if you use hostnames.

#### Backend Connection Failures

**Symptoms**: The gateway returns 5xx errors or "upstream connect error".

**Diagnosis**:
```bash
kubectl get aiservicebackend -A
kubectl get endpoints -n <backend-namespace>
kubectl run test-pod --image=curlimages/curl -i --tty --rm -- /bin/sh
```

### Configuration Issues

#### Invalid YAML Manifests

**Symptoms**: `kubectl apply` fails with validation errors.

**Common issues**:
- Wrong API versions.
- Missing required fields.
- Invalid values.
- Indentation errors.

**Solution**:
```bash
kubectl apply --dry-run=client -f your-manifest.yaml
kubectl explain aigatewayroute.spec
kubectl explain aiservicebackend.spec
```

#### Rate Limiting Not Working

**Symptoms**: Rate limits are not enforced.

**Check these first**:
1. Redis is deployed and reachable.
2. BackendTrafficPolicy is configured.
3. Metadata keys match between AIGatewayRoute and BackendTrafficPolicy.

**Diagnosis**:
```bash
kubectl get pods -n envoy-gateway-system | grep redis
kubectl get backendtrafficpolicy -A
kubectl logs -n envoy-gateway-system deployment/envoy-gateway | grep ratelimit
```

### Authentication and Authorization Issues

#### Upstream Authentication Failures

**Symptoms**: 401 or 403 errors from AI providers.

**Common causes**:
- Wrong API keys or credentials.
- Missing auth configuration.
- Expired tokens.

**Diagnosis**:
```bash
kubectl get backendsecuritypolicy -A
kubectl get secret <auth-secret-name> -o yaml
kubectl describe backendsecuritypolicy <policy-name>
```

#### Missing Required Headers

**Symptoms**: Requests fail with "missing required header" errors.

**Solution**: Ensure requests include the required headers:
```bash
curl -H "Content-Type: application/json" \
     -H "x-user-id: user123" \
     -d '{
       "model": "gpt-3.5-turbo",
       "messages": [{"role": "user", "content": "Hello"}]
     }' \
     $GATEWAY_URL/v1/chat/completions
```

## Detailed Diagnostics

### Component Health Checks

#### Envoy Gateway

<Tabs groupId="health-checks">
<TabItem value="status" label="Deployment Status">

```bash
kubectl get deployment -n envoy-gateway-system
kubectl rollout status deployment/envoy-gateway -n envoy-gateway-system
kubectl get pods -n envoy-gateway-system
```

</TabItem>
<TabItem value="logs" label="Logs Analysis">

```bash
kubectl logs -n envoy-gateway-system deployment/envoy-gateway
kubectl logs -n envoy-gateway-system deployment/envoy-gateway | grep -i error
kubectl logs -n envoy-gateway-system deployment/envoy-gateway | grep -i warn
```

</TabItem>
<TabItem value="config" label="Configuration">

```bash
kubectl get envoygateway -o yaml
kubectl get gatewayclass envoy-gateway-class -o yaml
```

</TabItem>
</Tabs>

#### AI Gateway Controller

<Tabs groupId="ai-gateway-health">
<TabItem value="controller" label="Controller Status">

```bash
kubectl get deployment -n envoy-ai-gateway-system
kubectl rollout status deployment/ai-gateway-controller -n envoy-ai-gateway-system
kubectl logs -n envoy-ai-gateway-system deployment/ai-gateway-controller
```

</TabItem>
<TabItem value="resources" label="Resource Status">

```bash
kubectl get aigatewayroute -A
kubectl get aiservicebackend -A
kubectl get backendsecuritypolicy -A
kubectl describe aigatewayroute <route-name>
```

</TabItem>
</Tabs>

### Log Analysis

#### Key Log Locations

| Component | Namespace | Deployment | Purpose |
|-----------|-----------|------------|---------|
| Envoy Gateway | `envoy-gateway-system` | `envoy-gateway` | Control plane operations |
| AI Gateway Controller | `envoy-ai-gateway-system` | `ai-gateway-controller` | AI-specific resource management |
| Envoy Proxy | `envoy-gateway-system` | `envoy-<gateway-name>-*` | Data plane traffic processing |
| Redis (if used) | `envoy-gateway-system` | `redis` | Rate limiting state |

#### Common Log Patterns

**Configuration issues**:
```bash
kubectl logs -n envoy-ai-gateway-system deployment/ai-gateway-controller | grep -i "validation\|invalid\|error"
```

**Connectivity issues**:
```bash
kubectl logs -n envoy-gateway-system <envoy-proxy-pod> | grep -i "upstream\|connection\|timeout"
```

**Authentication issues**:
```bash
kubectl logs -n envoy-gateway-system <envoy-proxy-pod> | grep -i "auth\|401\|403\|unauthorized"
```

### Network Diagnostics

#### Service Mesh Connectivity

```bash
kubectl run debug-pod --image=nicolaka/netshoot -i --tty --rm -- /bin/bash
nslookup envoy-gateway.envoy-gateway-system.svc.cluster.local
curl -v http://envoy-gateway.envoy-gateway-system.svc.cluster.local:18000/ready
```

#### External Connectivity

```bash
kubectl run curl-test --image=curlimages/curl -i --tty --rm -- /bin/sh
curl -v https://api.openai.com/v1/models
curl -v https://bedrock-runtime.us-east-1.amazonaws.com
```

## Performance Issues

### High Latency

**Symptoms**: Requests take longer than expected.

**Investigation**:
```bash
kubectl top pods -n envoy-gateway-system
kubectl top pods -n envoy-ai-gateway-system
kubectl describe pod <pod-name> -n <namespace>
kubectl port-forward -n envoy-gateway-system svc/envoy-gateway-metrics-service 19001:19001
curl http://localhost:19001/stats
```

### Resource Exhaustion

**Symptoms**: Pods are killed due to resource limits.

**Solutions**:
```bash
kubectl describe pod <pod-name> -n <namespace>
kubectl patch deployment <deployment-name> -n <namespace> -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container-name>","resources":{"limits":{"memory":"1Gi","cpu":"500m"},"requests":{"memory":"512Mi","cpu":"200m"}}}]}}}}'
```

## Getting Help

If issues continue after using this guide:

### Community Support

- **Slack**: Join the [#envoy-ai-gateway](https://envoyproxy.slack.com/archives/C07Q4N24VAA) channel.
- **GitHub Discussions**: Post questions in [GitHub Discussions](https://github.com/envoyproxy/ai-gateway/discussions).
- **Weekly Meetings**: Attend [community meetings](https://docs.google.com/document/d/10e1sfsF-3G3Du5nBHGmLjXw5GVMqqCvFDqp_O65B0_w) every Thursday.

### Bug Reports

For bugs and feature requests, please [create an issue](https://github.com/envoyproxy/ai-gateway/issues/new). Include:
- A clear problem description.
- Steps to reproduce.
- Environment details.
- Relevant logs and configs.

### Diagnostic Information

When asking for help, include:

```bash
kubectl version
kubectl get nodes
kubectl get pods -n envoy-ai-gateway-system
kubectl get pods -n envoy-gateway-system
kubectl get crd | grep ai-gateway
kubectl get aigatewayroute -A -o yaml
kubectl get aiservicebackend -A -o yaml
```

## Related Documentation

- [Getting Started](../getting-started/) - Setup and basics
- [Installation Guide](../getting-started/installation.md) - Install instructions
- [Capabilities](../capabilities/) - Feature configuration
- [Security](../capabilities/security/) - Auth and security
- [Traffic Management](../capabilities/traffic/) - Rate limiting and traffic policies
