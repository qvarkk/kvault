package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypePdfProcess = "pdf:process"
	TypeUrlFetch   = "url:fetch"
)

func NewPdfProcessTask(payload PdfProcessPayload) (*asynq.Task, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypePdfProcess, jsonPayload), nil
}

func NewUrlFetchTask(payload UrlFetchPayload) (*asynq.Task, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeUrlFetch, jsonPayload), nil
}
