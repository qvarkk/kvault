package domain

import "time"

type FileStatus string
type ItemType string

const (
	FileStatusUploaded   FileStatus = "uploaded"
	FileStatusProcessing FileStatus = "processing"
	FileStatusReady      FileStatus = "ready"
	FileStatusError      FileStatus = "error"
)

const (
	ItemTypeText ItemType = "text"
	ItemTypeFile ItemType = "file"
	ItemTypeUrl  ItemType = "url"
)

type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	APIKey    string    `db:"api_key" json:"api_key"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Item struct {
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	Type       ItemType  `db:"type"`
	Title      string    `db:"title"`
	Content    string    `db:"content"`
	FileMetaID *string   `db:"file_meta_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type FileMeta struct {
	ID        string     `db:"id"`
	Path      string     `db:"path"`
	Size      int64      `db:"size"`
	MimeType  string     `db:"mime_type"`
	Hash      *string    `db:"hash"`
	Status    FileStatus `db:"status"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

type Tag struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ItemTag struct {
	ItemID string `db:"item_id"`
	TagID  string `db:"tag_id"`
}
