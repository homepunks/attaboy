package bot

type Update struct {
	UpdateID int64   `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageID int64   `json:"message_id"`
	Chat      Chat    `json:"chat"`
	Text      string  `json:"text"`
	Photo     []Photo `json:"photo"`
}

type Chat struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type Photo struct {
	FileID   string `json:"file_id"`
	FileSize int64  `json:"file_size"`
	Width    int32  `json:"width"`
	Height   int32  `json:"height"`
}
