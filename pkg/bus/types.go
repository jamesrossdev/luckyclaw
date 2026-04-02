package bus

import "context"

type InboundMessage struct {
	Channel            string            `json:"channel"`
	SenderID           string            `json:"sender_id"`
	ChatID             string            `json:"chat_id"`
	Content            string            `json:"content"`
	Media              []string          `json:"media,omitempty"`
	SessionKey         string            `json:"session_key"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	StanzaID           string            `json:"stanza_id,omitempty"`            // WhatsApp: message ID for reply-to
	ReplyToParticipant string            `json:"reply_to_participant,omitempty"` // WhatsApp: sender JID for reply-to
}

type OutboundMessage struct {
	Channel            string `json:"channel"`
	ChatID             string `json:"chat_id"`
	Content            string `json:"content"`
	FilePath           string `json:"file_path,omitempty"`            // For file attachments
	ReplyToID          string `json:"reply_to_id,omitempty"`          // Discord: reply to this message ID
	ReplyToStanzaID    string `json:"reply_to_stanza_id,omitempty"`   // WhatsApp: reply to this stanza ID
	ReplyToParticipant string `json:"reply_to_participant,omitempty"` // WhatsApp: sender JID for reply-to

	// Event scheduling fields (WhatsApp EventMessage)
	EventName            string  `json:"event_name,omitempty"`        // Event title
	EventDescription     string  `json:"event_description,omitempty"` // Event description
	EventStartTime       int64   `json:"event_start_time,omitempty"`  // Unix timestamp in UTC
	EventEndTime         int64   `json:"event_end_time,omitempty"`    // Unix timestamp in UTC
	EventLocationName    string  `json:"event_location_name,omitempty"`
	EventLocationAddress string  `json:"event_location_address,omitempty"`
	EventLatitude        float64 `json:"event_latitude,omitempty"`
	EventLongitude       float64 `json:"event_longitude,omitempty"`
	EventJoinLink        string  `json:"event_join_link,omitempty"`   // Video call link
	EventIsCall          bool    `json:"event_is_call,omitempty"`     // Video/voice call event
	EventIsCanceled      bool    `json:"event_is_canceled,omitempty"` // Cancel an existing event
}

type MessageHandler func(InboundMessage) error

// OutboundHandler sends a message synchronously, returning any error.
// Used by SendDirect for error-aware sends (e.g. the message tool).
type OutboundHandler func(ctx context.Context, msg OutboundMessage) error
