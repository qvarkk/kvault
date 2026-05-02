package domain

import (
	"database/sql"
	"time"
)

type (
	FileStatus     string
	ItemType       string
	TagSource      string
	StopwordSource string
)

const (
	FileStatusUploading  FileStatus = "uploading"
	FileStatusProcessing FileStatus = "processing"
	FileStatusReady      FileStatus = "ready"
	FileStatusError      FileStatus = "error"
)

const (
	ItemTypeText ItemType = "text"
	ItemTypeUrl  ItemType = "url"
)

const (
	TagSourceAuto   TagSource = "auto"
	TagSourceManual TagSource = "manual"
)

const (
	StopwordSourceDefault StopwordSource = "default"
	StopwordSourceUser    StopwordSource = "user"
)

type User struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	APIKey    string    `db:"api_key"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Item struct {
	ID           string         `db:"id"`
	UserID       string         `db:"user_id"`
	Type         ItemType       `db:"type"`
	Title        string         `db:"title"`
	Content      sql.NullString `db:"content"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	SearchVector string         `db:"search_vector"`
	DeletedAt    sql.NullTime   `db:"deleted_at"`
}

type File struct {
	ID           string         `db:"id"`
	UserID       string         `db:"user_id"`
	OriginalName string         `db:"original_name"`
	TextContent  sql.NullString `db:"text_content"`
	S3Key        string         `db:"s3_key"`
	Size         int64          `db:"size"`
	MimeType     string         `db:"mime_type"`
	Status       FileStatus     `db:"status"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	SearchVector string         `db:"search_vector"`
	DeletedAt    sql.NullTime   `db:"deleted_at"`
}

type Tag struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Stopword struct {
	Word      string         `db:"word"`
	UserID    string         `db:"user_id"`
	Source    StopwordSource `db:"source"`
	IsEnabled bool           `db:"is_enabled"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type ItemTag struct {
	ItemID string    `db:"item_id"`
	TagID  string    `db:"tag_id"`
	Source TagSource `db:"source"`
}
