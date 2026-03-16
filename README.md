# Real-Time Messaging Gateway

This repository contains the implementation of a high-performance real-time communication gateway built on top of WebSockets and a custom TCP engine. The architecture is designed for low-latency delivery, high throughput, and a clean separation between connection management and business logic.

### ⚡ TL;DR
*   **Dual-Layer Design:** WebSocket Gateway (Proxy) + TCP Engine (Brain).
*   **Custom Protocol:** binary framing for all internal communication.
*   **High Performance:** fan-out and asynchronous database persistence via worker pools.
*   **Feature Rich:** Real-time presence, typing indicators, and full message lifecycle (Edit/Delete).

---

# Overview

<details>
<summary><b>Architecture Diagram</b></summary>

The following diagram illustrates the data flow from external clients through the Gateway and into the core TCP Engine.

![Backend Architecture](./docs/Architecture%20review.svg)

</details>


### WebSocket Gateway (The Front Office)
The Gateway acts as the entry point for all browser and external clients. It manages the full WebSocket lifecycle and performs real-time translation between browser-based JSON and our internal binary framing protocol. This ensures the core engine only deals with structured, pre-validated binary packets, reducing CPU overhead and memory fragmentation.

### TCP Engine (The Brain)
The Engine is the central logic layer of the system, operating over raw TCP. It maintains the global state of active user connections and conversations in-memory for maximum speed.
*   **Intelligent Fan-out:** When a message is received, the engine identifies all online participants across group or private chats and broadcasts the payload to their specific TCP streams simultaneously.
*   **Async Persistence:** To maintain sub-millisecond response times, the Engine offloads all database operations to a background worker pool, ensuring that slow disk I/O never blocks the real-time pipeline.

These two components communicate using a dedicated 👉 **[`custom binary protocol`](./internal/protocol/PROTOCOL.md)**.

---

## Feature Set (Protocol Layer)

The system utilizes a custom binary protocol with a dedicated packet factory. The following features are implemented natively at the protocol level:

*   **Message Fan-out:** Native support for broadcasting messages across Group and Private chats with high concurrency.
*   **Real-time Presence:** Automatic tracking and global notification of user online/offline status.
*   **Typing Indicators:** Low-overhead packets to signal active typing status within specific conversations.
*   **Message Lifecycle:** Full support for real-time message updates (editing) and deletions across all authorized participant clients.
*   **Conversation Management:** Real-time notifications when users are added to or removed from conversations.
*   **Connection Heartbeats:** Built-in Ping/Pong packets to maintain connection health and aggressively prune stale sessions.

---

## API & Documentation
*   **Standardized Design:** The API follows the Microsoft REST API Guidelines for consistent resource naming, error classification, and response structures.
*   **OpenAPI Specification:** The entire API is fully documented using OpenAPI 3.0. Detailed definitions and path descriptions are available in the `/api` directory.

