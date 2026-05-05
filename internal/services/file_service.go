package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/logger"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type FileRepo interface {
	CreateNew(context.Context, *domain.File) error
	List(context.Context, domain.ListFileFilter) ([]domain.File, int, error)
	GetByID(context.Context, string) (*domain.File, error)
	GetActiveByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.File, error)
	GetDeletedByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.File, error)
	SoftDeleteByIDTx(context.Context, *sqlx.Tx, string) error
	RestoreByIDTx(context.Context, *sqlx.Tx, string) error
}

type FileService struct {
	fileRepo     FileRepo
	transactor   Transactor
	tasker       TaskEnqueuer
	storage      FileStorage
	cache        CacheStore
	fileCacheTtl time.Duration
}

type CreateFileInput struct {
	UserID       string
	OriginalName string
	S3Key        string
	Size         int64
	MimeType     string
	Status       string
}

type cachedFileList struct {
	Files []domain.File
	Count int
}

func NewFileService(
	fileRepo FileRepo,
	transactor Transactor,
	tasker TaskEnqueuer,
	storage FileStorage,
	cache CacheStore,
	cacheTtl time.Duration,
) *FileService {
	return &FileService{
		fileRepo:     fileRepo,
		transactor:   transactor,
		tasker:       tasker,
		storage:      storage,
		cache:        cache,
		fileCacheTtl: cacheTtl,
	}
}

func (s *FileService) Upload(
	ctx context.Context,
	userID string,
	fileHeader *multipart.FileHeader,
) (*domain.File, error) {
	body, err := s.validatePdfAndOpenFile(fileHeader)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	s3Key := uuid.New().String() + ".pdf"
	err = s.storage.Upload(ctx, s3Key, body)
	if err != nil {
		return nil, err
	}

	fileInput := CreateFileInput{
		UserID:       userID,
		OriginalName: fileHeader.Filename,
		S3Key:        s3Key,
		Size:         fileHeader.Size,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		Status:       string(domain.FileStatusUploading),
	}

	file, err := s.createNew(ctx, fileInput)
	if err != nil {
		_ = s.storage.Delete(ctx, s3Key)
		return nil, err
	}

	err = s.tasker.EnqueuePdfProcess(ctx, userID, file.ID)
	if err != nil {
		_ = s.storage.Delete(ctx, s3Key)
		return nil, err
	}

	s.invalidateFileListCache(ctx, userID)

	return file, nil
}

func (s *FileService) validatePdfAndOpenFile(fileHeader *multipart.FileHeader) (io.ReadCloser, error) {
	ext := filepath.Ext(fileHeader.Filename)
	if ext != ".pdf" {
		return nil, NewServiceError(ErrPdfFileFormat, "invalid file extension", nil)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to open uploaded file", err)
	}

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		file.Close()
		return nil, NewServiceError(ErrInternal, "failed to read uploaded file", err)
	}

	contentType := http.DetectContentType(buffer[:n])
	if contentType != "application/pdf" {
		file.Close()
		return nil, NewServiceError(ErrPdfFileFormat, "invalid file content type", nil)
	}

	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	return file, nil
}

func (s *FileService) createNew(ctx context.Context, input CreateFileInput) (*domain.File, error) {
	file := &domain.File{
		UserID:       input.UserID,
		OriginalName: input.OriginalName,
		S3Key:        input.S3Key,
		Size:         input.Size,
		MimeType:     input.MimeType,
		Status:       domain.FileStatus(input.Status),
	}

	err := s.fileRepo.CreateNew(ctx, file)
	if err != nil {
		return nil, NewServiceError(ErrFileNotCreated, "database error", err)
	}

	return file, nil
}

func (s *FileService) List(ctx context.Context, f domain.ListFileFilter) ([]domain.File, int, error) {
	version, err := s.fileListVersion(ctx, f.UserID)
	if err != nil {
		logger.Logger.Warn("failed to get file list cache version",
			zap.String("userID", f.UserID),
			zap.Error(err),
		)
	}

	cacheKey := fileListKey(version, f)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		var result cachedFileList
		if err := json.Unmarshal(cached, &result); err == nil {
			return result.Files, result.Count, nil
		}
	}

	files, count, err := s.fileRepo.List(ctx, f)
	if err != nil {
		return nil, 0, NewServiceError(ErrInternal, "list files internal error", err)
	}

	if payload, err := json.Marshal(cachedFileList{Files: files, Count: count}); err == nil {
		if err := s.cache.Set(ctx, cacheKey, payload, s.fileCacheTtl); err != nil {
			logger.Logger.Warn("failed to set file list cache",
				zap.String("key", cacheKey),
				zap.Error(err),
			)
		}
	}

	return files, count, err
}

func (s *FileService) GetFilePresignedUrl(ctx context.Context, fileID, userID string) (*domain.PresignedURL, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, NewServiceError(ErrFileNotFound, "not found", err)
	}

	if file.UserID != userID {
		return nil, NewServiceError(ErrFileNotFound, "forbidden", nil)
	}

	url, expiresAt, err := s.storage.GeneratePresignUrl(ctx, file.S3Key, file.OriginalName)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "generate presign url error", err)
	}

	return &domain.PresignedURL{
		URL:       url,
		Filename:  file.OriginalName,
		MimeType:  file.MimeType,
		Size:      file.Size,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *FileService) DeleteByID(ctx context.Context, fileID, userID string) error {
	return s.authorizeAndMutateTx(
		ctx, fileID, userID,
		s.fileRepo.GetActiveByIDForUpdate,
		s.fileRepo.SoftDeleteByIDTx,
	)
}

func (s *FileService) RestoreByID(ctx context.Context, fileID, userID string) error {
	return s.authorizeAndMutateTx(
		ctx, fileID, userID,
		s.fileRepo.GetDeletedByIDForUpdate,
		s.fileRepo.RestoreByIDTx,
	)
}

func (s *FileService) authorizeAndMutateTx(
	ctx context.Context, fileID, userID string,
	getFn func(context.Context, *sqlx.Tx, string) (*domain.File, error),
	mutateFn func(context.Context, *sqlx.Tx, string) error,
) error {
	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		file, err := getFn(ctx, tx, fileID)
		if err != nil {
			return NewServiceError(ErrFileNotFound, "not found", err)
		}

		if file.UserID != userID {
			return NewServiceError(ErrFileNotFound, "forbidden", nil)
		}

		err = mutateFn(ctx, tx, fileID)
		if err != nil {
			return NewServiceError(ErrInternal, "mutate file internal error", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.invalidateFileListCache(ctx, userID)

	return nil
}

func fileListVersionKey(userID string) string {
	return fmt.Sprintf("files:version:user:%s", userID)
}

func fileListKey(version int64, f domain.ListFileFilter) string {
	return fmt.Sprintf(
		"files:list:v%d:user:%s:mime:%s:page:%d:size:%d:dis:%s:col:%s:q:%s",
		version, f.UserID, f.MimeType, f.Page, f.PageSize, f.Direction, f.Column, f.Query,
	)
}

func (s *FileService) fileListVersion(ctx context.Context, userID string) (int64, error) {
	key := fileListVersionKey(userID)
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if val == nil {
		return 0, err
	}
	return strconv.ParseInt(string(val), 10, 64)
}

func (s *FileService) invalidateFileListCache(ctx context.Context, userID string) {
	_, err := s.cache.Incr(ctx, fileListVersionKey(userID))
	if err != nil {
		logger.Logger.Warn("failed to invalidate file list cache",
			zap.String("userID", userID),
			zap.Error(err),
		)
	}
}
