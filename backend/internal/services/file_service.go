package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"qvarkk/kvault/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type FileRepo interface {
	CreateNew(context.Context, *domain.File) error
	List(context.Context, *domain.ListFileFilter) ([]domain.File, int, error)
	ListDeleted(context.Context, *domain.ListFileFilter) ([]domain.File, int, error)
	GetAllDeleted(ctx context.Context, userID string) ([]domain.File, error)
	GetAllByUserID(ctx context.Context, userID string) ([]domain.File, error)
	GetByID(context.Context, string) (*domain.File, error)
	GetActiveByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.File, error)
	GetDeletedByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.File, error)
	SoftDeleteByIDTx(context.Context, *sqlx.Tx, string) error
	RestoreByIDTx(context.Context, *sqlx.Tx, string) error
	PermanentlyDeleteAllDeleted(ctx context.Context, userID string) error
	PermanentlyDeleteByIDTx(context.Context, *sqlx.Tx, string) error
	HardDeleteByID(ctx context.Context, fileID string) error
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
	defer func() {
		if err := body.Close(); err != nil {
			zap.L().Error("failed to close PDF file body", zap.Error(err))
		}
	}()

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

	file, err := s.createNew(ctx, &fileInput)
	if err != nil {
		_ = s.storage.Delete(ctx, s3Key)
		return nil, err
	}

	err = s.tasker.EnqueuePdfProcess(ctx, userID, file.ID)
	if err != nil {
		if delErr := s.storage.Delete(ctx, s3Key); delErr != nil {
			zap.L().Warn("failed to delete s3 object after enqueue failure",
				zap.String("s3_key", s3Key), zap.Error(delErr))
		}
		if delErr := s.fileRepo.HardDeleteByID(ctx, file.ID); delErr != nil {
			zap.L().Warn("failed to delete file row after enqueue failure",
				zap.String("file_id", file.ID), zap.Error(delErr))
		}
		return nil, err
	}

	invalidateListCache(ctx, s.cache, fileListVersionKey(userID))

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
		if err := file.Close(); err != nil {
			return nil, NewServiceError(ErrInternal, "failed to close file", err)
		}
		return nil, NewServiceError(ErrInternal, "failed to read uploaded file", err)
	}

	contentType := http.DetectContentType(buffer[:n])
	if contentType != "application/pdf" {
		if err := file.Close(); err != nil {
			return nil, NewServiceError(ErrInternal, "failed to close file", err)
		}
		return nil, NewServiceError(ErrPdfFileFormat, "invalid file content type", nil)
	}

	if seeker, ok := file.(io.Seeker); ok {
		if _, err = seeker.Seek(0, io.SeekStart); err != nil {
			return nil, NewServiceError(ErrInternal, "failed to reset file pointer", err)
		}
	}

	return file, nil
}

func (s *FileService) createNew(ctx context.Context, input *CreateFileInput) (*domain.File, error) {
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

func (s *FileService) List(ctx context.Context, f *domain.ListFileFilter) ([]domain.File, int, error) {
	version := listVersion(ctx, s.cache, fileListVersionKey(f.UserID))
	cacheKey := fileListKey(version, f)

	if cached := getFromCache[cachedList[domain.File]](ctx, s.cache, cacheKey); cached != nil {
		return cached.Entities, cached.Count, nil
	}

	files, count, err := s.fileRepo.List(ctx, f)
	if err != nil {
		return nil, 0, NewServiceError(ErrInternal, "list files internal error", err)
	}

	setToCache(ctx, s.cache, cacheKey, cachedList[domain.File]{Entities: files, Count: count}, s.fileCacheTtl)

	return files, count, err
}

func (s *FileService) GetByID(ctx context.Context, fileID, userID string) (*domain.File, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, NewServiceError(ErrFileNotFound, "not found", err)
	}
	if file.UserID != userID {
		return nil, NewServiceError(ErrFileNotFound, "forbidden", nil)
	}
	return file, nil
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

	invalidateListCache(ctx, s.cache, fileListVersionKey(userID))

	return nil
}

func (s *FileService) GetFilePresignedViewUrl(ctx context.Context, fileID, userID string) (*domain.PresignedURL, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, NewServiceError(ErrFileNotFound, "not found", err)
	}

	if file.UserID != userID {
		return nil, NewServiceError(ErrFileNotFound, "forbidden", nil)
	}

	url, expiresAt, err := s.storage.GeneratePresignViewUrl(ctx, file.S3Key)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "generate presign view url error", err)
	}

	return &domain.PresignedURL{
		URL:       url,
		Filename:  file.OriginalName,
		MimeType:  file.MimeType,
		Size:      file.Size,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *FileService) ListDeleted(ctx context.Context, f *domain.ListFileFilter) ([]domain.File, int, error) {
	files, count, err := s.fileRepo.ListDeleted(ctx, f)
	if err != nil {
		return nil, 0, NewServiceError(ErrInternal, "list deleted files error", err)
	}
	return files, count, nil
}

func (s *FileService) ClearTrash(ctx context.Context, userID string) error {
	files, err := s.fileRepo.GetAllDeleted(ctx, userID)
	if err != nil {
		return NewServiceError(ErrInternal, "get deleted files error", err)
	}

	for i := range files {
		if err := s.storage.Delete(ctx, files[i].S3Key); err != nil {
			// best effort — log but continue so DB records are cleaned up
			zap.L().Warn("failed to delete s3 object while clearing trash",
				zap.String("s3_key", files[i].S3Key), zap.String("file_id", files[i].ID), zap.Error(err))
		}
	}

	if err := s.fileRepo.PermanentlyDeleteAllDeleted(ctx, userID); err != nil {
		return NewServiceError(ErrInternal, "permanently delete files error", err)
	}

	invalidateListCache(ctx, s.cache, fileListVersionKey(userID))
	return nil
}

func (s *FileService) PermanentlyDeleteByID(ctx context.Context, fileID, userID string) error {
	var s3Key string
	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		file, err := s.fileRepo.GetDeletedByIDForUpdate(ctx, tx, fileID)
		if err != nil {
			return NewServiceError(ErrFileNotFound, "not found", err)
		}
		if file.UserID != userID {
			return NewServiceError(ErrFileNotFound, "forbidden", nil)
		}
		s3Key = file.S3Key
		return s.fileRepo.PermanentlyDeleteByIDTx(ctx, tx, fileID)
	})
	if err != nil {
		return err
	}
	_ = s.storage.Delete(ctx, s3Key)
	invalidateListCache(ctx, s.cache, fileListVersionKey(userID))
	return nil
}

func (s *FileService) DeleteAllByUserID(ctx context.Context, userID string) error {
	files, err := s.fileRepo.GetAllByUserID(ctx, userID)
	if err != nil {
		return NewServiceError(ErrInternal, "get user files error", err)
	}

	for i := range files {
		if err := s.storage.Delete(ctx, files[i].S3Key); err != nil {
			zap.L().Warn("failed to delete s3 object while deleting user files",
				zap.String("s3_key", files[i].S3Key), zap.String("file_id", files[i].ID), zap.Error(err))
		}
	}

	return nil
}

func fileListVersionKey(userID string) string {
	return fmt.Sprintf("files:version:user:%s", userID)
}

func fileListKey(version int64, f *domain.ListFileFilter) string {
	return fmt.Sprintf(
		"files:list:v%d:user:%s:mime:%s:page:%d:size:%d:dir:%s:col:%s:q:%s",
		version, f.UserID, f.MimeType, f.Page, f.PageSize, f.Direction, f.Column, f.Query,
	)
}
