---
slug: envoy-in-the-era-of-ai-traffic
title: Envoy in the Era of AI - Building the Future of Traffic Handling Together
authors: [missberg, yanavlasov]
tags: [news]
---

# Envoy in the Era of AI: Building the Future of Traffic Handling Together
**How software systems communicate is evolving faster than ever.** Agents talk to agents, agents leverage tools, and applications interact dynamically with AI-driven services. These emerging patterns are redefining network architectures. If you thought the transition from monolithic websites to microservices reshaped networking, this new AI-driven evolution takes it to an entirely new level.

This article provides a broader Envoy Ecosystem view, and how Envoy AI Gateway fits into it. **Learn how Envoy Proxy, a foundational CNCF project, and its community are building the future of traffic handling.**

---
## How Envoy fits perfectly into the era of AI
Envoy Proxy is designed to handle dynamic and complex traffic patterns with high performance, and can handle high volumes of concurrent requests while remaining memory efficient. Its extensible architecture makes it future-proof and equipped to adapt quickly to the evolving demands of AI-driven interactions. This adaptability ensures that Envoy remains at the forefront of technological advancements.

**Envoy isn't just a proxy; it's a platform.** With its robust plugin ecosystem, Envoy enables engineers and developers to create solutions tailored to their traffic-handling needs. This extensibility is critical in an era where emerging standards and protocols, such as Google's A2A (Agent-to-Agent) and the MCP (Managed Client Protocol), require rapid innovation and community-driven evolution.

---

## A community built on collaboration
**The real strength of Envoy lies in its vibrant, collaborative builder community.** With contributors from industry leaders such as Google, Tetrate, Lyft, AWS, and others, the Envoy community attracts some of the brightest minds in software engineering, networking, and infrastructure management. The diversity of perspectives within our community ensures that Envoy is built to solve today's problems and proactively tackle tomorrow's challenges.

**Cross-industry collaboration is at the heart of Envoy's success.** The community's commitment to openness and transparency makes it easy for new members to get involved, share ideas, and collaborate on solutions. As traffic patterns evolve, the Envoy community continues to push forward, refining existing capabilities and developing new features to meet changing needs.

---
## Tackling AI traffic handling in Envoy
AI-driven workloads introduce unique challenges for traffic handling, including dynamic protocol negotiation, stateful conversational flows, and secure and efficient data exchange between diverse AI systems. Protocols like MCP (a lightweight JSON-RPC-based protocol for dynamically discovering and invoking tools) and A2A (Google's Agent-to-Agent protocol) represent just the beginning of this evolution.

Envoy must evolve to support emerging protocols and new ways of making intelligent routing decisions, security enforcement, and enhanced observability tailored to AI-driven interactions, addressing the emerging needs for AI traffic patterns.

The Envoy community recognizes these challenges and actively develops strategies and extensions to ensure the proxy remains robust and future-proof.

### The Data Plane
**From a data plane perspective,** the Envoy Proxy is evolving to support new protocols better, for example, through recent work and proposals for MCP support [Proposal: Envoy Support for Model Context Protocol (MCP) · Issue \#39174](https://github.com/envoyproxy/envoy/issues/39174). Other examples include rate-limiting based on dynamic metadata to support token-based rate-limiting when acting as an LLM gateway.

### The Control Plane
**On the control plane side,** the Envoy AI Gateway project within the Envoy ecosystem aims to simplify the use of Envoy Proxy as an AI Gateway. This gateway provides a uniform API for accessing diverse large language models (LLMs) services, supporting self-hosted and external services. It incorporates essential features such as authentication, failover between LLM services or models, and rate limiting. Additionally, it offers users the necessary configuration patterns to quickly set up and integrate Envoy Proxy with their AI servers and workloads. This approach accelerates the adoption and integration of GenAI features in their applications.

---
## Building the future together
We are excited to see everyone coming together within the community, including builders, developers, and architects, collaborating and innovating within the Envoy project.

We welcome anyone who wants to join. Whether your expertise lies in networking, AI, security, or observability, your insights are valuable as we collectively shape the next generation of proxy technology.

Let's collaborate openly, innovate swiftly, and tackle the evolving challenges head-on.
Join us in the Envoy community, and let's build the future of traffic handling together.

### Community Links

- [Join us on Slack](https://communityinviter.com/apps/envoyproxy/envoy%20)
- [Join us on GitHub](https://github.com/envoyproxy%20)
