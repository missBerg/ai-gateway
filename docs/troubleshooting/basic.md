# Troubleshooting Envoy AI Gateway on Kubernetes

This troubleshooting guide is designed to help you identify and fix common issues with your Envoy AI Gateway deployment on Kubernetes. Follow these steps to ensure each component of your setup is operating smoothly.

## Overview

In this section, you'll find troubleshooting tips covering deployment verification, pod status checks, connectivity issues, resource allocation, and log inspections based on the latest configurations and capabilities described in the documentation.

## Basic Checks

### Deployment Verification
- **Are Components Deployed Correctly?**
  - Verify the presence of deployments:
    ```bash
    kubectl get deployments -n envoy-gateway-system
    ```
  - Check deployments for availability:
    ```bash
    kubectl rollout status deployment/envoy-gateway -n envoy-gateway-system
    ```
  - Ensure that CRDs like `AIGatewayRoute`, `AIServiceBackend`, and `BackendSecurityPolicy` are correctly applied.

### Pod Status
- **Are Pods Running and Healthy?**
  - Check pod status in the specific namespace:
    ```bash
    kubectl get pods -n envoy-gateway-system
    ```
  - For detailed pod logs to diagnose issues, including security-related configurations and authentication issues:
    ```bash
    kubectl logs -n envoy-ai-gateway-system deployment/ai-gateway-controller
    ```

### Connectivity
- **Is Service Connectivity Proper?**
  - List services and check if an external IP is assigned:
    ```bash
    kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway-basic
    ```
  - Verify the application of `BackendTrafficPolicy` for rate limiting and security settings.

## Logs Inspection

### Envoy Proxy Logs
- **Inspect Logs for Errors:**
  - Access Envoy proxy logs to check for connectivity issues, especially for token-based rate limiting:
    ```bash
    kubectl logs <envoy-proxy-pod-name> -c envoy
    ```

### AI Gateway Sidecar Logs
- **Check Sidecar Logs:**
  - View logs for in-depth diagnostics, focusing on the latest API configurations:
    ```bash
    kubectl logs <ai-gateway-sidecar-pod-name>
    ```

### Control Plane Logs
- **Review Gateway Control Plane Logs:**
  - Check logs for any discrepancies in the configuration:
    ```bash
    kubectl logs -n envoy-ai-gateway-system deployment/ai-gateway-controller
    ```

## Additional Resources

- For an in-depth understanding and specific troubleshooting scenarios, refer to the latest [Envoy AI Gateway Proposal](../proposals/001-ai-gateway-proposal/proposal.md).
- For more, visit [envoyproxy.io](https://gateway.envoyproxy.io) _(opens in a new tab)_.

By following these guidelines, you should be able to troubleshoot common issues effectively. For advanced support, consider community forums or deeper log insights.
