package models

import (
	"time"
	"net"
	"gorm.io/gorm"
)

// UpdateStat represents an update statistic record in the database
type UpdateStat struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Version       string         `json:"version" gorm:"not null;size:50;index" validate:"required"`
	ClientID      string         `json:"client_id" gorm:"size:128;index"`
	ClientVersion string         `json:"client_version" gorm:"size:50"`
	Region        string         `json:"region" gorm:"size:10"`
	IPAddress     net.IP         `json:"ip_address" gorm:"type:inet"`
	UserAgent     string         `json:"user_agent" gorm:"type:text"`
	Action        string         `json:"action" gorm:"not null;size:20;index" validate:"required,oneof=check download install success failed"`
	ErrorMessage  string         `json:"error_message" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for UpdateStat model
func (UpdateStat) TableName() string {
	return "update_stats"
}

// UpdateStatRequest represents the request payload for creating update statistics
type UpdateStatRequest struct {
	Version       string `json:"version" validate:"required"`
	ClientID      string `json:"client_id"`
	ClientVersion string `json:"client_version"`
	Region        string `json:"region"`
	UserAgent     string `json:"user_agent"`
	Action        string `json:"action" validate:"required,oneof=check download install success failed"`
	ErrorMessage  string `json:"error_message"`
}

// UpdateStatResponse represents the response payload for update stat queries
type UpdateStatResponse struct {
	ID            uint      `json:"id"`
	Version       string    `json:"version"`
	ClientID      string    `json:"client_id"`
	ClientVersion string    `json:"client_version"`
	Region        string    `json:"region"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	Action        string    `json:"action"`
	ErrorMessage  string    `json:"error_message"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToResponse converts UpdateStat model to UpdateStatResponse
func (us *UpdateStat) ToResponse() *UpdateStatResponse {
	return &UpdateStatResponse{
		ID:            us.ID,
		Version:       us.Version,
		ClientID:      us.ClientID,
		ClientVersion: us.ClientVersion,
		Region:        us.Region,
		IPAddress:     us.IPAddress.String(),
		UserAgent:     us.UserAgent,
		Action:        us.Action,
		ErrorMessage:  us.ErrorMessage,
		CreatedAt:     us.CreatedAt,
	}
}
