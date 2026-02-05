package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// PlusOneInfo for guests bringing additional people
type PlusOneInfo struct {
	FirstName string `bson:"first_name" json:"first_name"`
	LastName  string `bson:"last_name" json:"last_name"`
	Dietary   string `bson:"dietary,omitempty" json:"dietary,omitempty"`
}

// CustomAnswer stores responses to custom questions
type CustomAnswer struct {
	QuestionID string      `bson:"question_id" json:"question_id"`
	Question   string      `bson:"question" json:"question"`
	Answer     interface{} `bson:"answer" json:"answer"` // Can be string, []string, bool, etc.
}

type RSVP struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	WeddingID primitive.ObjectID  `bson:"wedding_id" json:"wedding_id"`
	GuestID   *primitive.ObjectID `bson:"guest_id,omitempty" json:"guest_id,omitempty"` // Link to pre-registered guest

	// Guest Information (if not linked to pre-registered guest)
	FirstName string `bson:"first_name" json:"first_name" validate:"required,max=50"`
	LastName  string `bson:"last_name" json:"last_name" validate:"required,max=50"`
	Email     string `bson:"email,omitempty" json:"email,omitempty" validate:"omitempty,email,max=100"`
	Phone     string `bson:"phone,omitempty" json:"phone,omitempty"`

	// RSVP Response
	Status          string `bson:"status" json:"status" validate:"oneof=attending not-attending maybe"`
	AttendanceCount int    `bson:"attendance_count" json:"attendance_count" validate:"min=1"`

	// Plus Ones
	PlusOnes     []PlusOneInfo `bson:"plus_ones,omitempty" json:"plus_ones,omitempty"`
	PlusOneCount int           `bson:"plus_one_count" json:"plus_one_count" validate:"min=0,max=5"`

	// Dietary & Preferences
	DietaryRestrictions string   `bson:"dietary_restrictions,omitempty" json:"dietary_restrictions,omitempty"`
	DietarySelected     []string `bson:"dietary_selected,omitempty" json:"dietary_selected,omitempty"`
	AdditionalNotes     string   `bson:"additional_notes,omitempty" json:"additional_notes,omitempty" validate:"omitempty,max=500"`

	// Custom Questions Answers
	CustomAnswers []CustomAnswer `bson:"custom_answers,omitempty" json:"custom_answers,omitempty"`

	// Metadata
	SubmittedAt time.Time  `bson:"submitted_at" json:"submitted_at"`
	UpdatedAt   *time.Time `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	IPAddress   string     `bson:"ip_address,omitempty" json:"-"` // For spam prevention
	UserAgent   string     `bson:"user_agent,omitempty" json:"-"` // For analytics

	// Confirmation
	ConfirmationSent   bool       `bson:"confirmation_sent" json:"confirmation_sent"`
	ConfirmationSentAt *time.Time `bson:"confirmation_sent_at,omitempty" json:"confirmation_sent_at,omitempty"`

	// Internal tracking
	Source string `bson:"source" json:"source" validate:"oneof=web direct_link qr_code manual"`
	Notes  string `bson:"notes,omitempty" json:"notes,omitempty"` // Admin notes
}

// RSVPStatus represents possible response statuses
type RSVPStatus string

const (
	RSVPAttending    RSVPStatus = "attending"
	RSVPNotAttending RSVPStatus = "not-attending"
	RSVPMaybe        RSVPStatus = "maybe"
)

// RSVPStatistics for dashboard display
type RSVPStatistics struct {
	TotalResponses  int            `json:"total_responses"`
	Attending       int            `json:"attending"`
	NotAttending    int            `json:"not_attending"`
	Maybe           int            `json:"maybe"`
	TotalGuests     int            `json:"total_guests"` // Including plus ones
	PlusOnesCount   int            `json:"plus_ones_count"`
	DietaryCounts   map[string]int `json:"dietary_counts"`
	SubmissionTrend []DailyCount   `json:"submission_trend"`
}

type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// RSVPSource represents where the RSVP came from
type RSVPSource string

const (
	RSVPSourceWeb        RSVPSource = "web"
	RSVPSourceDirectLink RSVPSource = "direct_link"
	RSVPSourceQRCode     RSVPSource = "qr_code"
	RSVPSourceManual     RSVPSource = "manual"
)

// Helper methods for RSVP
func (r *RSVP) GetFullName() string {
	return r.FirstName + " " + r.LastName
}

func (r *RSVP) GetTotalGuests() int {
	return r.AttendanceCount + r.PlusOneCount
}

func (r *RSVP) IsConfirmed() bool {
	return r.ConfirmationSent && r.ConfirmationSentAt != nil
}

func (r *RSVP) CanBeModified() bool {
	// Allow modification within 24 hours of submission
	return time.Since(r.SubmittedAt) <= 24*time.Hour
}
