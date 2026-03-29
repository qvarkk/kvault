package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeFileUpload = "file:upload"
)

func NewFileUploadTask(payload FileUploadPayload) (*asynq.Task, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeFileUpload, jsonPayload), nil
}
