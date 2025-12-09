
# Real-Time Messaging Gateway

This repository contains the implementation of a real-time communication
gateway built on top of WebSockets and a custom TCP engine. The system is
designed for high-performance, low-latency message delivery between clients,
with a clear separation between transport, protocol handling, and business
logic.



# Overview

![System Architecture Diagram](./docs/System%20Arhc%202.svg)

The gateway architecture consists of two major components:

- **WebSocket Server** – Handles inbound connections from browsers and external
  clients. It receives JSON messages, validates them, and maps them to
  internal packet types.

- **TCP Engine** – A dedicated backend engine responsible for processing
  packets, routing messages, managing client sessions, and performing the core
  application logic.

These two components communicate using a custom binary protocol documented in
[`PROTOCOL.md`](./pkg/protocol/PROTOCOL.md).

---

# Why WebSocket + TCP?

While WebSockets provide a convenient and widely supported real-time
transport layer for browsers, they are not ideal for internal backend logic
or scalable message routing.

A dedicated **TCP engine** allows us to:

- Keep real-time logic isolated and modular  
- Manage thousands of WebSocket clients efficiently  
- Execute business logic without being tied to HTTP/WebSocket constraints  
- Utilize custom binary framing instead of JSON overhead  
- Scale horizontally by running multiple gateways connected to one engine  

This design creates a clean separation between:

- **Frontend protocol** (WebSocket JSON)
- **Internal protocol** (binary TCP frames)
- **Application logic** (TCP engine)

---

# Custom Protocol

The gateway and engine communicate using a compact binary frame format
(Magic + Opcode + Length + Payload).  
This protocol is documented thoroughly in:

👉 **[`PROTOCOL.md`](./pkg/protocol/PROTOCOL.md)**

The README gives only the high-level view — all encoding, decoding, and byte-level
details remain in the protocol document.

---




This repository is actively expanding. Planned additions include:


