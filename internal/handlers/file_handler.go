package handlers

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/services"
	"qvarkk/kvault/internal/tasks"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type FileService interface {
	CreateNew(context.Context, services.CreateFileMetaInput) (*domain.FileMeta, error)
	ValidatePdfFile(context.Context, *multipart.FileHeader) error
	UploadPdfFileToS3(context.Context, *multipart.FileHeader) (string, error)
	EnqueuePdfProcessTask(context.Context, tasks.PdfProcessPayload) (*asynq.TaskInfo, error)
}

type FileHandler struct {
	fileService FileService
	itemService ItemService
}

func NewFileHandler(fileService FileService, itemService ItemService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		itemService: itemService,
	}
}

type uploadFileForm struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func (f *FileHandler) UploadFile(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	var form uploadFileForm
	if err := ctx.ShouldBind(&form); err != nil {
		abortOnBindError(ctx, err)
		return
	}

	err := f.fileService.ValidatePdfFile(ctx, form.File)
	if errors.Is(err, services.ErrPdfFileFormat) {
		abortWithPublicError(ctx, err, "Failed to validate PDF file.")
		return
	} else if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	s3Key, err := f.fileService.UploadPdfFileToS3(ctx, form.File)
	if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	fileMetaInput := services.CreateFileMetaInput{
		S3Key:    s3Key,
		Size:     form.File.Size,
		MimeType: form.File.Header.Get("Content-Type"),
		Status:   string(domain.FileStatusUploading),
	}

	fileMeta, err := f.fileService.CreateNew(ctx.Request.Context(), fileMetaInput)
	if errors.Is(err, services.ErrItemNotCreated) || fileMeta == nil {
		abortOnDbError(ctx, err)
		return
	} else if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	itemInput := services.CreateItemInput{
		UserID:     userID,
		Type:       string(domain.ItemTypeFile),
		Title:      form.File.Filename,
		FileMetaID: fileMeta.ID,
	}

	item, err := f.itemService.CreateNew(ctx.Request.Context(), itemInput)
	if errors.Is(err, services.ErrItemNotCreated) || item == nil {
		abortOnDbError(ctx, err)
		return
	} else if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	payload := tasks.PdfProcessPayload{
		UserID:     userID,
		FileMetaID: fileMeta.ID,
		ItemID:     item.ID,
	}

	_, err = f.fileService.EnqueuePdfProcessTask(ctx, payload)
	if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, toTaskResponse(item, fileMeta))
}
