package tasks

type FileUploadPayload struct {
	UserID       string
	FileMetaID   string
	ItemID       string
	TempFilepath string
}
