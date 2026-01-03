package web

// APIEvent represents a single event in the API response.
// This is a flattened structure for JSON serialization.
type APIEvent struct {
	Number          int    `json:"number"`
	Type            string `json:"type"`
	TimestampMillis int64  `json:"timestamp_millis"`
	Participant     string `json:"participant,omitempty"`
	Content         string `json:"content,omitempty"`
	Next            string `json:"next,omitempty"`
	ID              string `json:"id,omitempty"`
}

// StatusResponse is the response for GET /api/status
type StatusResponse struct {
	SessionID    string     `json:"session_id"`
	Participants []string   `json:"participants"`
	EventCount   int        `json:"event_count"`
	Events       []APIEvent `json:"events"`
}

// PostRequest is the request body for POST /api/post
type PostRequest struct {
	Session string  `json:"session"`
	Content string  `json:"content"`
	After   int     `json:"after"`
	Next    *string `json:"next,omitempty"`
}

// PostResponse is the response for POST /api/post
type PostResponse struct {
	EventNumber int `json:"event_number"`
}

// ParticipantsResponse is the response for GET /api/participants
type ParticipantsResponse struct {
	Participants []string `json:"participants"`
}

// ErrorResponse for API errors
type ErrorResponse struct {
	Error string `json:"error"`
}
