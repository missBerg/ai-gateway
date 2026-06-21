# Security Policy

The Envoy AI Gateway maintainers and community take security seriously. We
appreciate your efforts to responsibly disclose your findings, and will make
every effort to acknowledge your contributions.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues,
discussions, pull requests, or any other public channel.**

Instead, please report vulnerabilities privately using one of the following
channels:

1. **Preferred:** Open a
   [GitHub Security Advisory](https://github.com/envoyproxy/ai-gateway/security/advisories/new).
   This allows the maintainers to triage the report and collaborate with you
   privately, directly on GitHub.

2. **Alternative:** Email the Envoy AI Gateway maintainers at
   [envoy-ai-gateway-security@googlegroups.com](mailto:envoy-ai-gateway-security@googlegroups.com).

If the vulnerability relates to Envoy Proxy or Envoy Gateway (the components
that Envoy AI Gateway is built on top of) rather than Envoy AI Gateway itself,
please report it to the relevant upstream project:

- Envoy Proxy: https://github.com/envoyproxy/envoy/security/advisories/new
- Envoy Gateway: https://github.com/envoyproxy/gateway/security/advisories/new

### What to Include

To help us triage and resolve the issue as quickly as possible, please include
as much of the following information as you can:

- A description of the vulnerability and its potential impact.
- The affected version(s) of Envoy AI Gateway.
- Steps to reproduce the issue, including any sample configuration, requests, or
  proof-of-concept code.
- Any known workarounds or suggested remediation.

> **Note**: If your report includes data that has privacy concerns, please
> sanitize the data prior to sharing it.

## Disclosure Process

After a vulnerability is reported, the maintainers will follow this process:

1. **Acknowledgement** — We aim to acknowledge receipt of your report within
   **3 business days**.
2. **Triage** — We will investigate to validate the report, determine the
   affected versions, and assess the severity and impact.
3. **Fix and coordination** — We will work on a fix and coordinate a release
   timeline with you. We aim to resolve or publicly disclose privately reported
   issues within **90 days** of the initial report.
4. **Release and disclosure** — Once a fix is available, we will publish the
   patched release and an accompanying security advisory.

We will keep you informed of the progress throughout the process. If you would
like to be credited for the discovery, please let us know, and we will include
an acknowledgement in the advisory (with your consent).

Please keep the details of any reported vulnerability confidential until a fix
has been released and the embargo has been lifted, so that users have a chance
to upgrade.

## Supported Versions

Envoy AI Gateway is pre-`v1.0.0` and follows the support policy described in
[RELEASES.md](./RELEASES.md). In short, the end of life for a given release is
**two releases after** it was published (for example, `v0.1.0` reaches end of
life when `v0.3.0` is released).

Security fixes are applied to the `main` branch and backported to the latest
supported release line. We strongly recommend that all users run a supported,
up-to-date version. Reports against unsupported versions may be addressed by
asking you to reproduce the issue on a supported version.

## Security Updates

Security patches and advisories are announced through:

- The [GitHub Releases page](https://github.com/envoyproxy/ai-gateway/releases)
- The [GitHub Security Advisories page](https://github.com/envoyproxy/ai-gateway/security/advisories)

We recommend watching the repository and subscribing to these channels to stay
informed about security updates.

## Best Practices for Secure Usage

To minimize security risks when running Envoy AI Gateway:

- Run the latest supported version and apply patches promptly.
- Follow the principle of least privilege when configuring credentials and
  `BackendSecurityPolicy` resources, and store secrets using Kubernetes Secrets
  (or an external secret manager) rather than in plaintext configuration.
- Restrict network access to the control plane and management interfaces.
- Regularly review the security-related documentation under the project
  [docs](./site/docs/capabilities/security).

## Contact

If you have any questions about this security policy, please reach out to the
maintainers listed in [MAINTAINERS.md](./MAINTAINERS.md) or email
[envoy-ai-gateway-security@googlegroups.com](mailto:envoy-ai-gateway-security@googlegroups.com).

Thank you for helping to keep Envoy AI Gateway and its users secure.
