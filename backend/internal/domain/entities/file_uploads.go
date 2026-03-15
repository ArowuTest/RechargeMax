package entities

import (
	"time"

)

// FileUpload represents the file_uploads table
type FileUpload struct {
	FileName    string     `json:"file_name" gorm:"column:file_name;not null" validate:"required"`
	FilePath    string     `json:"file_path" gorm:"column:file_path;not null" validate:"required"`
	FileSize    int        `json:"file_size" gorm:"column:file_size;not null" validate:"required"`
	MimeType    string     `json:"mime_type" gorm:"column:mime_type;not null" validate:"required"`
	BucketName  string     `json:"bucket_name" gorm:"column:bucket_name;not null" validate:"required"`
	FileType    string     `json:"file_type" gorm:"column:file_type;not null" validate:"required"`
	Description string     `json:"description" gorm:"column:description"`
	IsActive    *bool      `json:"is_active" gorm:"column:is_active"`
	IsVerified  *bool      `json:"is_verified" gorm:"column:is_verified"`
	UploadIp    string     `json:"upload_ip" gorm:"column:upload_ip"`
	UserAgent   string     `json:"user_agent" gorm:"column:user_agent"`
	CreatedAt   *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   *time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for FileUpload
func (FileUpload) TableName() string {
	return "file_uploads"
}
