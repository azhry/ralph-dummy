package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserStatus_Constants(t *testing.T) {
	assert.Equal(t, "active", string(UserStatusActive))
	assert.Equal(t, "inactive", string(UserStatusInactive))
	assert.Equal(t, "suspended", string(UserStatusSuspended))
}

func TestWedding_IsRSVPOpen(t *testing.T) {
	tests := []struct {
		name    string
		wedding *Wedding
		want    bool
	}{
		{
			name: "RSVP enabled, no deadline",
			wedding: &Wedding{
				RSVP: RSVPSettings{
					Enabled: true,
				},
			},
			want: true,
		},
		{
			name: "RSVP disabled",
			wedding: &Wedding{
				RSVP: RSVPSettings{
					Enabled: false,
				},
			},
			want: false,
		},
		{
			name: "RSVP enabled, deadline passed",
			wedding: &Wedding{
				RSVP: RSVPSettings{
					Enabled:  true,
					Deadline: &time.Time{},
				},
			},
			want: false, // Assuming deadline is in the past
		},
		{
			name: "RSVP enabled, future deadline",
			wedding: &Wedding{
				RSVP: RSVPSettings{
					Enabled: true,
					Deadline: func() *time.Time {
						future := time.Now().Add(24 * time.Hour)
						return &future
					}(),
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.wedding.IsRSVPOpen())
		})
	}
}

func TestWedding_IsAccessible(t *testing.T) {
	tests := []struct {
		name    string
		wedding *Wedding
		want    bool
	}{
		{
			name: "public published wedding",
			wedding: &Wedding{
				Status:   string(WeddingStatusPublished),
				IsPublic: true,
			},
			want: true,
		},
		{
			name: "public draft wedding",
			wedding: &Wedding{
				Status:   string(WeddingStatusDraft),
				IsPublic: true,
			},
			want: false,
		},
		{
			name: "private published wedding",
			wedding: &Wedding{
				Status:   string(WeddingStatusPublished),
				IsPublic: false,
			},
			want: true,
		},
		{
			name: "private draft wedding",
			wedding: &Wedding{
				Status:   string(WeddingStatusDraft),
				IsPublic: false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.wedding.IsAccessible())
		})
	}
}

func TestWedding_IsExpired(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name    string
		wedding *Wedding
		want    bool
	}{
		{
			name: "no expiry date",
			wedding: &Wedding{
				ExpiresAt: nil,
			},
			want: false,
		},
		{
			name: "expired wedding",
			wedding: &Wedding{
				ExpiresAt: &past,
			},
			want: true,
		},
		{
			name: "not expired wedding",
			wedding: &Wedding{
				ExpiresAt: &future,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.wedding.IsExpired())
		})
	}
}

func TestRSVP_GetFullName(t *testing.T) {
	rsvp := &RSVP{
		FirstName: "John",
		LastName:  "Doe",
	}
	assert.Equal(t, "John Doe", rsvp.GetFullName())
}

func TestRSVP_GetTotalGuests(t *testing.T) {
	tests := []struct {
		name string
		rsvp *RSVP
		want int
	}{
		{
			name: "single guest",
			rsvp: &RSVP{
				AttendanceCount: 1,
				PlusOneCount:    0,
			},
			want: 1,
		},
		{
			name: "guest with plus ones",
			rsvp: &RSVP{
				AttendanceCount: 2,
				PlusOneCount:    2,
			},
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.rsvp.GetTotalGuests())
		})
	}
}

func TestRSVP_CanBeModified(t *testing.T) {
	now := time.Now()
	old := now.Add(-25 * time.Hour)
	recent := now.Add(-23 * time.Hour)

	tests := []struct {
		name string
		rsvp *RSVP
		want bool
	}{
		{
			name: "old RSVP",
			rsvp: &RSVP{
				SubmittedAt: old,
			},
			want: false,
		},
		{
			name: "recent RSVP",
			rsvp: &RSVP{
				SubmittedAt: recent,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.rsvp.CanBeModified())
		})
	}
}

func TestRSVPStatus_Constants(t *testing.T) {
	assert.Equal(t, "attending", string(RSVPAttending))
	assert.Equal(t, "not-attending", string(RSVPNotAttending))
	assert.Equal(t, "maybe", string(RSVPMaybe))
}

func TestWeddingStatus_Constants(t *testing.T) {
	assert.Equal(t, "draft", string(WeddingStatusDraft))
	assert.Equal(t, "published", string(WeddingStatusPublished))
	assert.Equal(t, "expired", string(WeddingStatusExpired))
	assert.Equal(t, "archived", string(WeddingStatusArchived))
}

func TestGuestModel_Fields(t *testing.T) {
	guestID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()
	createdBy := primitive.NewObjectID()

	guest := &Guest{
		ID:        guestID,
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		CreatedBy: createdBy,
	}

	assert.Equal(t, guestID, guest.ID)
	assert.Equal(t, weddingID, guest.WeddingID)
	assert.Equal(t, "John", guest.FirstName)
	assert.Equal(t, "Doe", guest.LastName)
	assert.Equal(t, "john@example.com", guest.Email)
	assert.Equal(t, createdBy, guest.CreatedBy)
}

func TestMediaModel_IsImage(t *testing.T) {
	tests := []struct {
		name     string
		media    *Media
		expected bool
	}{
		{
			name: "JPEG image",
			media: &Media{
				MimeType: "image/jpeg",
			},
			expected: true,
		},
		{
			name: "PNG image",
			media: &Media{
				MimeType: "image/png",
			},
			expected: true,
		},
		{
			name: "WebP image",
			media: &Media{
				MimeType: "image/webp",
			},
			expected: true,
		},
		{
			name: "non-image file",
			media: &Media{
				MimeType: "application/pdf",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.media.IsImage())
		})
	}
}

func TestMediaModel_HasThumbnails(t *testing.T) {
	tests := []struct {
		name     string
		media    *Media
		expected bool
	}{
		{
			name: "no thumbnails",
			media: &Media{
				Thumbnails: nil,
			},
			expected: false,
		},
		{
			name: "empty thumbnails",
			media: &Media{
				Thumbnails: map[string]string{},
			},
			expected: false,
		},
		{
			name: "with thumbnails",
			media: &Media{
				Thumbnails: map[string]string{
					"small": "http://example.com/small.jpg",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.media.HasThumbnails())
		})
	}
}

func TestMediaModel_GetThumbnailURL(t *testing.T) {
	media := &Media{
		Thumbnails: map[string]string{
			"small":  "http://example.com/small.jpg",
			"medium": "http://example.com/medium.jpg",
		},
	}

	// Test existing thumbnail
	url, exists := media.GetThumbnailURL("small")
	assert.True(t, exists)
	assert.Equal(t, "http://example.com/small.jpg", url)

	// Test non-existing thumbnail
	url, exists = media.GetThumbnailURL("large")
	assert.False(t, exists)
	assert.Empty(t, url)

	// Test with nil thumbnails
	media.Thumbnails = nil
	url, exists = media.GetThumbnailURL("small")
	assert.False(t, exists)
	assert.Empty(t, url)
}

func TestMediaModel_AddThumbnail(t *testing.T) {
	media := &Media{}

	// Add first thumbnail
	media.AddThumbnail("small", "http://example.com/small.jpg")
	assert.Len(t, media.Thumbnails, 1)
	assert.Equal(t, "http://example.com/small.jpg", media.Thumbnails["small"])

	// Add second thumbnail
	media.AddThumbnail("medium", "http://example.com/medium.jpg")
	assert.Len(t, media.Thumbnails, 2)
	assert.Equal(t, "http://example.com/medium.jpg", media.Thumbnails["medium"])

	// Overwrite existing thumbnail
	media.AddThumbnail("small", "http://example.com/new-small.jpg")
	assert.Len(t, media.Thumbnails, 2)
	assert.Equal(t, "http://example.com/new-small.jpg", media.Thumbnails["small"])
}

func TestMediaModel_IsDeleted(t *testing.T) {
	tests := []struct {
		name     string
		media    *Media
		expected bool
	}{
		{
			name: "not deleted",
			media: &Media{
				DeletedAt: nil,
			},
			expected: false,
		},
		{
			name: "deleted",
			media: &Media{
				DeletedAt: &time.Time{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.media.IsDeleted())
		})
	}
}

func TestMediaModel_SoftDelete(t *testing.T) {
	media := &Media{
		DeletedAt: nil,
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	beforeSoftDelete := time.Now()
	media.SoftDelete()

	assert.NotNil(t, media.DeletedAt)
	assert.True(t, media.DeletedAt.After(beforeSoftDelete))
	assert.True(t, media.UpdatedAt.After(beforeSoftDelete))
}

func TestMediaModel_GetExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "simple extension",
			filename: "photo.jpg",
			expected: "jpg",
		},
		{
			name:     "multiple dots",
			filename: "wedding.photo.final.jpg",
			expected: "jpg",
		},
		{
			name:     "no extension",
			filename: "photo",
			expected: "",
		},
		{
			name:     "empty filename",
			filename: "",
			expected: "",
		},
		{
			name:     "dot at end",
			filename: "photo.",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			media := &Media{Filename: tt.filename}
			assert.Equal(t, tt.expected, media.GetExtension())
		})
	}
}

func TestMediaModel_GetPublicURL(t *testing.T) {
	expectedURL := "http://example.com/photo.jpg"
	media := &Media{OriginalURL: expectedURL}
	assert.Equal(t, expectedURL, media.GetPublicURL())
}

func TestMediaModel_BeforeCreate(t *testing.T) {
	media := &Media{}
	beforeCreate := time.Now().Add(-1 * time.Second)

	media.BeforeCreate()

	assert.True(t, media.CreatedAt.After(beforeCreate))
	assert.True(t, media.UpdatedAt.After(beforeCreate))
}

func TestMediaModel_BeforeUpdate(t *testing.T) {
	originalUpdatedAt := time.Now().Add(-1 * time.Hour)
	media := &Media{
		UpdatedAt: originalUpdatedAt,
	}
	
	beforeUpdate := time.Now().Add(-1 * time.Second)
	
	media.BeforeUpdate()
	
	assert.True(t, media.UpdatedAt.After(beforeUpdate))
	assert.True(t, media.UpdatedAt.After(originalUpdatedAt))
}
