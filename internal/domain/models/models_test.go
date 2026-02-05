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
