package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Media represents a stored media file with metadata
type Media struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Filename    string                 `bson:"filename" json:"filename"`
	OriginalURL string                 `bson:"originalUrl" json:"originalUrl"`
	Thumbnails  map[string]string      `bson:"thumbnails,omitempty" json:"thumbnails,omitempty"`
	Size        int64                  `bson:"size" json:"size"`
	MimeType    string                 `bson:"mimeType" json:"mimeType"`
	Width       int                    `bson:"width,omitempty" json:"width,omitempty"`
	Height      int                    `bson:"height,omitempty" json:"height,omitempty"`
	Format      string                 `bson:"format,omitempty" json:"format,omitempty"`
	EXIF        map[string]interface{} `bson:"exif,omitempty" json:"exif,omitempty"`
	StorageKey  string                 `bson:"storageKey" json:"-"`
	CreatedAt   time.Time              `bson:"createdAt" json:"createdAt"`
	CreatedBy   primitive.ObjectID     `bson:"createdBy" json:"createdBy"`
	UpdatedAt   time.Time              `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
	DeletedAt   *time.Time             `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
}

// IsImage checks if the media file is an image
func (m *Media) IsImage() bool {
	return m.MimeType == "image/jpeg" || m.MimeType == "image/png" || m.MimeType == "image/webp"
}

// HasThumbnails checks if the media has thumbnails generated
func (m *Media) HasThumbnails() bool {
	return len(m.Thumbnails) > 0
}

// GetThumbnailURL returns the URL for a specific thumbnail size
func (m *Media) GetThumbnailURL(size string) (string, bool) {
	if m.Thumbnails == nil {
		return "", false
	}
	url, exists := m.Thumbnails[size]
	return url, exists
}

// AddThumbnail adds a thumbnail URL for a specific size
func (m *Media) AddThumbnail(size, url string) {
	if m.Thumbnails == nil {
		m.Thumbnails = make(map[string]string)
	}
	m.Thumbnails[size] = url
}

// IsDeleted checks if the media is soft-deleted
func (m *Media) IsDeleted() bool {
	return m.DeletedAt != nil
}

// SoftDelete marks the media as deleted
func (m *Media) SoftDelete() {
	now := time.Now()
	m.DeletedAt = &now
	m.UpdatedAt = now
}

// GetExtension returns the file extension from the filename
func (m *Media) GetExtension() string {
	if len(m.Filename) == 0 {
		return ""
	}
	
	// Find last dot
	for i := len(m.Filename) - 1; i >= 0; i-- {
		if m.Filename[i] == '.' {
			return m.Filename[i+1:]
		}
	}
	return ""
}

// GetPublicURL returns a URL suitable for public access (without storage key)
func (m *Media) GetPublicURL() string {
	return m.OriginalURL
}

// BeforeCreate sets timestamps before creating the record
func (m *Media) BeforeCreate() {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
}

// BeforeUpdate updates the timestamp before updating the record
func (m *Media) BeforeUpdate() {
	m.UpdatedAt = time.Now()
}