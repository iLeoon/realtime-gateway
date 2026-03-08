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
	Content         string `json:"content"`
	ConversationID  string `json:"conversationID"`
	RecipientUserID string `json:"recipientUserID"`
}

// UpdateMessagePayload is the JSON structure used by the browser to send a
// chat message that needs to be updated. It is extracted from the ClientPayload's raw JSON payload
// when the opcode indicates a send-message operation.
type UpdateMessagePayload struct {
	MessageID      string `json:"messageID"`
	Content        string `json:"content"`
	ConversationID string `json:"conversationID"`
}

// DeleteMessagePayload the JSON structure used by the browser to send a
// chat message that needs to be deleted. It is extracted from the ClientPayload's raw JSON payload
// when the opcode indicates a send-message operation.
type DeleteMessagePayload struct {
	MessageID      string `json:"messageID"`
	ConversationID string `json:"conversationID"`
}

// TypingPayload is the JSON structure used by the browser to signal a
// typing status change inside a conversation.
type TypingPayload struct {
	ConversationID string `json:"conversationID"`
	IsTyping       bool   `json:"isTyping"`
}
