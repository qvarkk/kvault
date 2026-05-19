package web

import (
	"context"
	"mime/multipart"
	"net/http"
	"qvarkk/kvault/internal/domain"

	"github.com/gin-gonic/gin"
)

type FileService interface {
	Upload(context.Context, string, *multipart.FileHeader) (*domain.File, error)
	List(context.Context, domain.ListFileFilter) ([]domain.File, int, error)
	ListDeleted(context.Context, domain.ListFileFilter) ([]domain.File, int, error)
	GetFilePresignedUrl(ctx context.Context, fileID, userID string) (*domain.PresignedURL, error)
	GetFilePresignedViewUrl(ctx context.Context, fileID, userID string) (*domain.PresignedURL, error)
	DeleteByID(ctx context.Context, fileID, userID string) error
	RestoreByID(ctx context.Context, fileID, userID string) error
	ClearTrash(ctx context.Context, userID string) error
	DeleteAllByUserID(ctx context.Context, userID string) error
}

type FileHandler struct {
	fileService FileService
}

func NewFileHandler(fileService FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

type uploadFileForm struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type listFileRequest struct {
	Query    string `form:"q"`
	MimeType string `form:"mime_type" binding:"omitempty,oneof=application/pdf"`
	PaginationParams
	FileSortingParams
}

type fileIDUri struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// @Summary      Upload a PDF file to your vault
// @Description  Validates and uploads given file to S3 container,
// @Description  enqueues redis task to process the file
// @Tags         Files
// @Security     ApiKeyAuth
// @Accept       mpfd
// @Produce      json
// @Param        file formData file true "PDF file"
// @Success      201   {object}  FileResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /files/upload [post]
func (h *FileHandler) UploadFile(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var form uploadFileForm
	if err := ctx.ShouldBind(&form); err != nil {
		return err
	}

	file, err := h.fileService.Upload(ctx, userID, form.File)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusCreated, toFileResponse(file))
	return nil
}

// @Summary      Get all files
// @Description  Returns a list of files owned by the User
// @Tags         Files
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param				 params query listFileRequest false "Query parameters"
// @Success      200   {object}  PaginatedResponse[FileResponse]
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /files [get]
func (h *FileHandler) List(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req listFileRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return err
	}

	params := domain.ListFileFilter{
		UserID:   userID,
		MimeType: req.MimeType,
		QueryFilter: domain.QueryFilter{
			Query: req.Query,
		},
		PaginationFilter: domain.PaginationFilter{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		SortFilter: domain.SortFilter{
			Direction: req.Direction,
			Column:    req.Column,
		},
	}

	files, total, err := h.fileService.List(ctx, params)
	if err != nil {
		return err
	}

	fileResponses := make([]FileResponse, len(files))
	for i, file := range files {
		fileResponses[i] = toFileResponse(&file)
	}

	ctx.JSON(http.StatusOK, toPaginatedResponse(fileResponses, total, params.Page, params.PageSize))
	return nil
}

// @Summary      Get a file from user's vault
// @Description  Gets a URL to download the file with given ID
// @Tags         Files
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "File ID"
// @Success      200   {object}  AwsUrlResponse      "Presigned URL Data"
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /files/{id} [get]
func (h *FileHandler) Download(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var uri fileIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	url, err := h.fileService.GetFilePresignedUrl(ctx.Request.Context(), uri.ID, userID)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toAwsUrlResponse(url))
	return nil
}

// @Summary      Soft delete a file
// @Description  Marks a file with given ID as deleted if it's owned by the User
// @Tags         Files
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "File ID"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /files/{id} [delete]
func (h *FileHandler) Delete(ctx *gin.Context) error {
	return h.withOwnedFileAction(ctx, h.fileService.DeleteByID)
}

// @Summary      Restore a soft deleted file
// @Description  Unmarks a file with given ID as deleted if it's owned by the User
// @Tags         Files
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "File ID"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /files/{id}/restore [post]
func (h *FileHandler) Restore(ctx *gin.Context) error {
	return h.withOwnedFileAction(ctx, h.fileService.RestoreByID)
}

func (h *FileHandler) withOwnedFileAction(
	ctx *gin.Context,
	fn func(context.Context, string, string) error,
) error {
	userID := ctx.MustGet("userID").(string)

	var uri itemIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	err := fn(ctx.Request.Context(), uri.ID, userID)
	if err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// @Summary      Get inline view URL for a file
// @Description  Returns a long-lived presigned URL with inline content disposition for browser viewing
// @Tags         Files
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "File ID"
// @Success      200   {object}  AwsUrlResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /files/{id}/view [get]
func (h *FileHandler) GetViewURL(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var uri fileIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	url, err := h.fileService.GetFilePresignedViewUrl(ctx.Request.Context(), uri.ID, userID)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toAwsUrlResponse(url))
	return nil
}

// @Summary      List deleted files
// @Description  Returns soft-deleted files owned by the user
// @Tags         Files
// @Security     ApiKeyAuth
// @Produce      json
// @Param        params query listFileRequest false "Query parameters"
// @Success      200   {object}  PaginatedResponse[FileResponse]
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /files/deleted [get]
func (h *FileHandler) ListDeleted(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req listFileRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return err
	}

	params := domain.ListFileFilter{
		UserID: userID,
		PaginationFilter: domain.PaginationFilter{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		SortFilter: domain.SortFilter{
			Direction: req.Direction,
			Column:    req.Column,
		},
	}

	files, total, err := h.fileService.ListDeleted(ctx, params)
	if err != nil {
		return err
	}

	fileResponses := make([]FileResponse, len(files))
	for i, file := range files {
		fileResponses[i] = toFileResponse(&file)
	}

	ctx.JSON(http.StatusOK, toPaginatedResponse(fileResponses, total, params.Page, params.PageSize))
	return nil
}

// @Summary      Clear file trash
// @Description  Permanently deletes all soft-deleted files owned by the user and removes them from S3
// @Tags         Files
// @Security     ApiKeyAuth
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /files/deleted [delete]
func (h *FileHandler) ClearTrash(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	if err := h.fileService.ClearTrash(ctx.Request.Context(), userID); err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}
