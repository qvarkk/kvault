package handlers

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/services"

	"github.com/gin-gonic/gin"
)

type ItemService interface {
	CreateNew(context.Context, services.CreateItemInput) (*domain.Item, error)
	CreateFileMeta(context.Context, services.CreateFileMetaInput) (*domain.FileMeta, error)
	ValidatePdfFile(context.Context, *multipart.FileHeader) error
	GeneratePdfFileDestination() string
}

type ItemHandler struct {
	itemService ItemService
}

func NewItemHandler(itemService ItemService) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

type createItemRequest struct {
	Type       string `json:"type" binding:"required,oneof=text file url"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	FileMetaID string `json:"file_meta_id" binding:"omitempty,uuid4"`
}

type uploadFileForm struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func (h *ItemHandler) Create(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	var req createItemRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		abortOnBindError(ctx, err)
		return
	}

	itemInput := services.CreateItemInput{
		UserID:     userID,
		Type:       req.Type,
		Title:      req.Title,
		Content:    req.Content,
		FileMetaID: req.FileMetaID,
	}

	item, err := h.itemService.CreateNew(ctx.Request.Context(), itemInput)
	if errors.Is(err, services.ErrItemNotCreated) || item == nil {
		abortOnDbError(ctx, err)
		return
	} else if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, toItemResponse(item))
}

func (h *ItemHandler) UploadFile(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	var form uploadFileForm
	if err := ctx.ShouldBind(&form); err != nil {
		abortOnBindError(ctx, err)
		return
	}

	err := h.itemService.ValidatePdfFile(ctx, form.File)
	if errors.Is(err, services.ErrPdfFileFormat) {
		abortWithPublicError(ctx, err, "Failed to validate PDF file.")
		return
	} else if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	dst := h.itemService.GeneratePdfFileDestination()
	if err := ctx.SaveUploadedFile(form.File, dst); err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	fileMetaInput := services.CreateFileMetaInput{
		Path:     dst,
		Size:     form.File.Size,
		MimeType: form.File.Header.Get("Content-Type"),
		Status:   string(domain.FileStatusUploaded),
	}

	fileMeta, err := h.itemService.CreateFileMeta(ctx.Request.Context(), fileMetaInput)
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

	item, err := h.itemService.CreateNew(ctx.Request.Context(), itemInput)
	if errors.Is(err, services.ErrItemNotCreated) || item == nil {
		abortOnDbError(ctx, err)
		return
	} else if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, toFileMetaResponse(fileMeta))
}
