package tcp

import "encoding/json"

// ClientPayload represents the standardized JSON format sent by the
// browser or mobile/desktop clients. Every inbound message contains a
// `type` field indicating the opcode, and a `payload` field containing the
// message-specific data encoded as raw JSON.
//
// The `payload` field is intentionally typed as json.RawMessage to allow
// flexible decoding into different data structures depending on the opcode.
//
// Typical client message format:
//
//	{
//	  "type": "send_message",
//	  "payload": {
//	    "content": "Hi"
//	  }
//	}
type ClientPayload struct {
	Opcode  string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SendMessagePayload is the JSON structure used by the browser to send a
// chat message. It is extracted from the ClientPayload's raw JSON payload
// when the opcode indicates a send-message operation.
type SendMessagePayload struct {
	Content string `json:"content"`
}
