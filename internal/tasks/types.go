package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeFileUpload = "file:upload"
)

func NewFileUploadTask(id int) (*asynq.Task, error) {
	payload, err := json.Marshal(FileUploadPayload{UserID: id})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeFileUpload, payload), nil
}
