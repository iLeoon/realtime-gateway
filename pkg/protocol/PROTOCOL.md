
  # 1. Introduction

  

This document defines the communication protocol used between the WebSocket

gateway and the TCP engine. The system converts browser JSON messages into

typed packets, encodes them into binary frames, and sends them over a TCP

connection. Responses from the engine travel back through the same path in

reverse.

  

  

# 2. High-Level Architecture

  

Browser JSON

↓

WebSocket Gateway

↓

Interface (Session)

↓

TCP Client (EncodeFrame → TCP)

↓

TCP Engine (DecodeFrame → Packet → Logic)

↓

Router → WebSocket Gateway

↓

Browser Client

  
  

The gateway handles WebSocket I/O while the engine performs the core logic.

  

  



# 3. Binary Frame Structure (TCP Wire Format)

All messages exchanged between the WebSocket gateway and the TCP engine use  
a binary frame containing a **6-byte header** followed by a payload.


```lua
+--------+--------+--------------+-----------------------+
| Magic  | Opcode | Length       | Payload [...]         |
+--------+--------+--------------+-----------------------+
 1 byte   1 byte     4 bytes          variable(N bytes)
```
## Encoding Rules

-   All multi-byte numbers use **big-endian**
    
-   Total frame size = `6 + payloadLength`
    
-   Magic byte must match protocol constant
    
-   Payload structure depends on the packet type
## Decoding Rules

When receiving raw bytes from the TCP stream, the decoder must reconstruct
the frame by reading fields in the exact order defined by the header
structure:

1. **Read Magic (1 byte)**  
   - Validate it matches the protocol’s magic constant  
   - If it does not match, the frame MUST be rejected and the connection
     closed

2. **Read Opcode (1 byte)**  
   - Determines which concrete packet type to construct  
   - Unknown opcodes MUST result in closing the session

3. **Read Length (4 bytes, big-endian)**  
   - Indicates the number of bytes to read for the payload  
   - If a length is larger than the allowed maximum, the frame MUST be
     discarded

4. **Read Payload (Length bytes)**  
   - These bytes are passed to the packet’s `Decode()` method  
   - Payload content depends on the packet type

5. **Construct Packet**
   - Use the Opcode to determine the correct `BuildPayload` implementation  
   - Call its `Decode(payload)` method to populate fields

6. **Return Fully Decoded Frame**
   - A complete Frame contains:
     - Magic
     - Opcode
     - Length
     - Payload (decoded into a packet struct)

If any step fails (invalid magic, invalid opcode, short read, mismatched
length, or decode error), the connection should be safely terminated.

## Example Frame Encoding

This example demonstrates the full transformation from a browser message
(JSON) into the final encoded raw bytes sent over the TCP connection.

### 1. Browser Sends JSON

```json
{
  "type": "send_message",
  "payload": {
    "content": "Hello"
  }
}
```
 JSON is converted into Raw Bytes
### 2. Final Encoded Frame (Raw Bytes)
```python
                    |------- Payload (9 bytes) -------| 
                    |                                 |
                    |                                 |
                    |                                 |

[137  3   0 0 0 9   163 62 41 230   72 101 108 108 111]
 |    |   |          |               |
 |    |   |          |               └─ Content ("Hello")
 |    |   |          └─ ConnectionID = 4 bytes
 |    |   └─ Length = 9 bytes (0 0 0 9)
 |    └─ Opcode = 3 (SendMessage)
 └─ Magic = 137
```


### Explanation

- **137** → Magic byte identifying the protocol  
- **3** → Opcode for `SendMessage`  
- **0 0 0 9** → Payload length = 9 bytes (The length changes based on the payload)
- **163 62 41 230** →  Example metadata (e.g., ConnectionID = `0xA33E29E6`)  
- **72 101 108 108 111** → `"Hello"` encoded in UTF-8  

---

## Example Decoding (Step-by-Step)

Below is the same byte sequence from the encoding example:

```text
[137  3   0 0 0 9   163 62 41 230   72 101 108 108 111]
```


### 1. Decoding Flow

**Step 1 — Read Magic (1 byte)**

-   Value: `137`
    
-   OK → matches protocol constant
    

**Step 2 — Read Opcode (1 byte)**

-   Value: `3`
    
-   Mapped to: `SendMessagePacket`
    

**Step 3 — Read Length (4 bytes, big-endian)**

-   Bytes: `0 0 0 9`
    
-   PayloadLength = **9**
    

**Step 4 — Read Payload (9 bytes):**
```text
163 62 41 230   72 101 108 108 111
```
Breakdown:

-   `163 62 41 230` → ConnectionID (`0xA33E29E6`)
    
-   `72 101 108 108 111` → `"Hello"`
    
**Step 5 — Construct Packet**

Based on Opcode `3`:

`SendMessagePacket` 

**Step 6 — Decode Payload Inside Packet**
After calling `Decode(payload)`:
```vbnet
ToConnectionID: A33E29E6 Message:  "Hello"
```
### 2. Final Decoded Packet
```go
Frame{
    Magic:  137,
    Opcode: 3,
    Length: 9,
    Payload: []byte{
        // ConnectionID (4 bytes)
        0xA3, 0x3E, 0x29, 0xE6,

        // Content "Hello" (UTF-8)
        'H', 'e', 'l', 'l', 'o',
    },
}
```
