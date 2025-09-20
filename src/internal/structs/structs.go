package structs

import (
	"time"
)



type BaseResponse struct {
	Error            string      `json:"error,omitempty"`
	Message          string      `json:"message,omitempty"`
	ErrorDetails     string      `json:"error_details,omitempty"`
}

type Slot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
}




type AvailabilityResponse struct {
	BaseResponse
	Slots          []AvailabilitySlot    `json:"slots"`
	TotalAvailable int       `json:"total_available"`
}

type BlockSlotRequest struct {
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
}

type BlockSlotResponse struct {
	Success        bool      `json:"success"`
	CoachId        string    `json:"coach_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	BlockedAt time.Time  `json:"blocked_at"`
	BlockId      string   `json:"block_id"`
}

type ReleaseSlotRequest struct {
	BlockId      string `json:"block_id"`
}

type ReleaseSlotResponse struct {
	Success        bool      `json:"success"`
	CoachId        string    `json:"coach_id"`
	ReleasedAt time.Time  `json:"released_at"`
	BlockId      string   `json:"block_id"`
}

type WorkingHours struct {
    Start    string `json:"start"`
    End      string `json:"end"`
    Timezone string `json:"timezone"`
}

type AvailabilityRules struct {
    MinNoticeHours int `json:"min_notice_hours"`
    MaxAdvanceDays int `json:"max_advance_days"`
    BufferMinutes  int `json:"buffer_minutes"`
}

type CoachSettingsResponse struct {
    CoachID          string            `json:"coach_id"`
    WorkingHours     WorkingHours      `json:"working_hours"`
    AvailabilityRules AvailabilityRules `json:"availability_rules"`
    BlockedDates     []string          `json:"blocked_dates"`
}


type GetAvailabilityResponse struct {
	CoachID        string `json:"coach_id"`
	Slots          []Slot `json:"slots"`
	TotalAvailable int    `json:"total_available"`
}


type AvailabilitySlot struct {
	CoachID   string    `json:"coach_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}


type WebHookRequest struct {
	EventType     string `json:"event_type"`
	AppointmentID string `json:"appointment_id"`
}

type WebHookResponse struct {
	BaseResponse
	Received      bool  `json:"received"`
	EventID       string  `json:"event_id"`
}

type CoachDistribution struct {
	CoachID           string  `json:"coach_id"`
	Name              string  `json:"name"`
	Email             string  `json:email`
	Score             float64 `json:score`
	AppointmentsCount int     `json:"appointments_count"`
	Utilization       float64 `json:"utilization"`
}

type CoachDistributionResponse struct {
	Distribution  []CoachDistribution `json:"distribution"`
	FairnessScore float64             `json:"fairness_score"`
}

type BookAppointmentRequest struct {
	CalendarID   string    `json:"calendar_id"`
	ContactEmail string    `json:"contact_email"`
	ContactName  string    `json:"contact_name"`
	StartTime    time.Time `json:"start_time"`
	TimeZone     string    `json:"timezone"`
	Notes        string    `json:"notes"`
}

type BookAppointmentResponse struct {
	BaseResponse
	AppointmentID string    `json:"appointment_id"`
	CoachID       string    `json:"coach_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:end_time`
	Status        string    `json:"status"`
}

type Coach struct {
	ID                   string
	Name                 string
	Email                string
	Score                float64
	MaxDailyAppointments int
	WorkingHoursStart    time.Time
	WorkingHoursEnd      time.Time
	Timezone             string
}



type ValidateRequest struct {
	APIKey string `json:"api_key"`
}

type ValidateResponse struct {
	Valid       bool   `json:"valid"`
	Error       string `json:"error,omitempty"`
	KeyType     string `json:"key_type,omitempty"`
	RateLimit   *RateLimit `json:"rate_limit,omitempty"`
	Permissions *Permissions `json:"permissions,omitempty"`
}

type RateLimit struct {
	Limit     int    `json:"limit"`
	Remaining int    `json:"remaining"`
	Reset     string `json:"reset"`
	RetryAfter int   `json:"retry_after,omitempty"`
}

type Permissions struct {
	Read   bool `json:"read"`
	Write  bool `json:"write"`
	Delete bool `json:"delete"`
}

type KeyInfo struct {
	KeyID     string `json:"key_id"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	LastUsed  string `json:"last_used"`
	RateLimit int    `json:"rate_limit"`
	Active    bool   `json:"active"`
}

type RotateKeyResponse struct {
	OldKey    string `json:"old_key"`
	NewKey    string `json:"new_key"`
	RotatedAt string `json:"rotated_at"`
	ExpiresIn int    `json:"expires_in"`
}


type AppointmentCreatedRequest struct {
	AppointmentID string `json:"appointment_id"`
	CoachID       string `json:"coach_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	ClientID      string `json:"client_id"`
}

type AppointmentCreatedResponse struct {
	Success     bool   `json:"success"`
	CrmID       string `json:"crm_id"`
	ProcessedAt string `json:"processed_at"`
	Message     string `json:"message,omitempty"`
}

type AppointmentUpdatedRequest struct {
	AppointmentID string `json:"appointment_id"`
	Status        string `json:"status"`
}

type AppointmentUpdatedResponse struct {
	Success     bool   `json:"success"`
	ProcessedAt string `json:"processed_at"`
	Message     string `json:"message,omitempty"`
}

type AppointmentCancelledRequest struct {
	AppointmentID string `json:"appointment_id"`
	Reason        string `json:"reason"`
}

type AppointmentCancelledResponse struct {
	Success     bool   `json:"success"`
	ProcessedAt string `json:"processed_at"`
}

type Contact struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Tags      []string  `json:"tags"`
	Score     float64   `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}
