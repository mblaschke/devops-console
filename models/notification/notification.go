package notification

type (
	NotificationMessage struct {
		Channel  string `json:"channel"`
		Username string `json:"username"`
		Text     string `json:"text"`
		Blocks   []NotificationMessageBlockContext `json:"blocks"`
	}

	NotificationMessageBlockContext struct {
		Type string  `json:"type"`
		Text *NotificationMessageBlockText `json:"text,omitempty"`
		Elements []*NotificationMessageBlockText `json:"elements,omitempty"`
	}

	NotificationMessageBlockText struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
)
