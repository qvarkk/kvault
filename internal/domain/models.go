package domain

import (
	"database/sql"
	"time"
)

type (
	FileStatus     string
	ItemType       string
	UrlStatus      string
	TagSource      string
	StopwordSource string
)

const (
	UrlStatusPending UrlStatus = "pending"
	UrlStatusReady   UrlStatus = "ready"
	UrlStatusError   UrlStatus = "error"
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
	ID         string    `db:"id"`
	Username   string    `db:"username"`
	Password   string    `db:"password"`
	APIKeyHash string    `db:"api_key_hash"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
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

	SourceURL        sql.NullString `db:"source_url"`
	UrlMetadata      sql.NullString `db:"url_metadata"`
	ExtractedContent sql.NullString `db:"extracted_content"`
	UrlStatus        sql.NullString `db:"url_status"`

	Tags []Tag
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
	Source    TagSource `db:"source"`
	ItemCount int       `db:"item_count"`
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
