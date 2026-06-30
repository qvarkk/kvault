package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"qvarkk/kvault/config"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var noExpiry = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

type LocalStorage struct {
	dir             string
	maxUploadSizeMB int64
}

func NewLocalStorage(cfg *config.StorageConfig) (*LocalStorage, error) {
	safePath := filepath.Clean(cfg.UploadDir)
	if !filepath.IsAbs(safePath) {
		return nil, fmt.Errorf("storage directory must be an absolute path: %q", safePath)
	}

	err := os.MkdirAll(safePath, 0o700)
	if err != nil {
		return nil, err
	}

	return &LocalStorage{
		dir:             safePath,
		maxUploadSizeMB: cfg.MaxUploadSizeMB,
	}, nil
}

func (s *LocalStorage) Upload(ctx context.Context, body io.Reader) (filename string, err error) {
	filename = uuid.New().String()
	path := filepath.Join(s.dir, filename)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return filename, fmt.Errorf("create file failed in upload: %w", err)
	}
	defer func() {
		cerr := file.Close()
		if cerr != nil && err == nil {
			err = fmt.Errorf("close file failed in upload:%w", cerr)
		}
		if err != nil {
			if rerr := os.Remove(path); rerr != nil {
				zap.L().Error("failed to remove partial file", zap.Error(rerr))
			}
		}
	}()

	limit := s.maxUploadSizeMB << 20
	n, err := io.Copy(file, &ctxReader{ctx: ctx, r: io.LimitReader(body, limit+1)})
	if err == nil && n > limit {
		err = fmt.Errorf("upload exceeds max size %d MB", s.maxUploadSizeMB)
	}
	if err == nil {
		err = file.Sync()
	}
	return filename, err
}

func (s *LocalStorage) Get(ctx context.Context, filename string) (io.ReadCloser, error) {
	if _, err := uuid.Parse(filename); err != nil {
		return nil, fmt.Errorf("filename should be an UUID: %w", err)
	}

	path := filepath.Join(s.dir, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file from storage: %w", err)
	}
	return &ctxFile{ctx: ctx, File: file}, nil
}

func (s *LocalStorage) Delete(ctx context.Context, filename string) error {
	if _, err := uuid.Parse(filename); err != nil {
		return fmt.Errorf("filename should be an UUID: %w", err)
	}

	path := filepath.Join(s.dir, filename)
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to delete the file from storage: %w", err)
	}
	return nil
}

func (s *LocalStorage) GenerateDownloadUrl(ctx context.Context, filename, originalName string) (
	url string,
	expiresAt time.Time,
	err error,
) {
	return rawUrl(filename, originalName, "attachment"), noExpiry, nil
}

func (s *LocalStorage) GenerateViewUrl(ctx context.Context, filename string) (
	url string,
	expiresAt time.Time,
	err error,
) {
	return rawUrl(filename, "", "inline"), noExpiry, nil
}

type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *ctxReader) Read(p []byte) (int, error) {
	if err := cr.ctx.Err(); err != nil {
		return 0, err
	}
	return cr.r.Read(p)
}

type ctxFile struct {
	ctx context.Context
	*os.File
}

func (cf *ctxFile) Read(p []byte) (int, error) {
	if err := cf.ctx.Err(); err != nil {
		return 0, err
	}
	return cf.File.Read(p)
}

func rawUrl(filename, originalName, disposition string) string {
	u := url.Values{}
	u.Set("name", originalName)
	u.Set("disposition", disposition)
	return "/files/raw/" + filename + "?" + u.Encode()
}
