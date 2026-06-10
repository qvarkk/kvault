package worker

import (
	"context"
	"encoding/json"
	"qvarkk/kvault/internal/tasks"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type UrlTaskService interface {
	FetchAndExtract(ctx context.Context, userID, itemID string) error
}

type UrlFetchHandler struct {
	urlTaskService UrlTaskService
}

func NewUrlFetchHandler(urlTaskService UrlTaskService) *UrlFetchHandler {
	return &UrlFetchHandler{urlTaskService: urlTaskService}
}

func (h *UrlFetchHandler) HandleUrlFetchTask(ctx context.Context, t *asynq.Task) error {
	var p tasks.UrlFetchPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		zap.L().Error("Failed to parse url fetch payload", zap.Error(err))
		return err
	}

	zap.L().Info("Starting URL fetch",
		zap.String("item_id", p.ItemID),
		zap.String("user_id", p.UserID),
	)

	if err := h.urlTaskService.FetchAndExtract(ctx, p.UserID, p.ItemID); err != nil {
		zap.L().Error("Failed to fetch URL",
			zap.Error(err),
			zap.String("item_id", p.ItemID),
		)
		return err
	}

	zap.L().Info("URL fetch complete", zap.String("item_id", p.ItemID))
	return nil
}
