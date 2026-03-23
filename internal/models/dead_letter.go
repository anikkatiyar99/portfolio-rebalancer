package models

type DeadLetterMessage struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Stage     string `json:"stage"`
	Reason    string `json:"reason"`
	Payload   string `json:"payload,omitempty"`
	CreatedAt string `json:"created_at"`
}
