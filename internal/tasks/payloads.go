package tasks

type PdfProcessPayload struct {
	UserID string
	FileID string
}

type UrlFetchPayload struct {
	UserID string
	ItemID string
}
