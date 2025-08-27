package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
)

// RolloutConfig represents the rollout configuration
type RolloutConfig struct {
	Percentage      int      `json:"percentage"`
	Regions         []string `json:"regions"`
	VersionRange    string   `json:"version_range"`
	StartTime       string   `json:"start_time"`
	EndTime         string   `json:"end_time"`
	MaxConcurrent   int      `json:"max_concurrent"`
}

// Value implements the driver.Valuer interface for JSONB storage
func (rc RolloutConfig) Value() (driver.Value, error) {
	return json.Marshal(rc)
}

// Scan implements the sql.Scanner interface for JSONB storage
func (rc *RolloutConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(b, rc)
}

// UpdateRule represents an update rule in the database
type UpdateRule struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	Name              string         `json:"name" gorm:"not null;size:100" validate:"required"`
	TargetRegion      string         `json:"target_region" gorm:"size:50"`
	TargetVersionRange string        `json:"target_version_range" gorm:"size:100"`
	Enabled           bool           `json:"enabled" gorm:"default:true"`
	Priority          int            `json:"priority" gorm:"default:0"`
	RolloutConfig     RolloutConfig  `json:"rollout_config" gorm:"type:jsonb"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for UpdateRule model
func (UpdateRule) TableName() string {
	return "update_rules"
}

// UpdateRuleRequest represents the request payload for creating/updating update rules
type UpdateRuleRequest struct {
	Name               string        `json:"name" validate:"required"`
	TargetRegion       string        `json:"target_region"`
	TargetVersionRange string        `json:"target_version_range"`
	Enabled            bool          `json:"enabled"`
	Priority           int           `json:"priority"`
	RolloutConfig      RolloutConfig `json:"rollout_config"`
}

// UpdateRuleResponse represents the response payload for update rule queries
type UpdateRuleResponse struct {
	ID                 uint          `json:"id"`
	Name               string        `json:"name"`
	TargetRegion       string        `json:"target_region"`
	TargetVersionRange string        `json:"target_version_range"`
	Enabled            bool          `json:"enabled"`
	Priority           int           `json:"priority"`
	RolloutConfig      RolloutConfig `json:"rollout_config"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// ToResponse converts UpdateRule model to UpdateRuleResponse
func (ur *UpdateRule) ToResponse() *UpdateRuleResponse {
	return &UpdateRuleResponse{
		ID:                 ur.ID,
		Name:               ur.Name,
		TargetRegion:       ur.TargetRegion,
		TargetVersionRange: ur.TargetVersionRange,
		Enabled:            ur.Enabled,
		Priority:           ur.Priority,
		RolloutConfig:      ur.RolloutConfig,
		CreatedAt:          ur.CreatedAt,
		UpdatedAt:          ur.UpdatedAt,
	}
}
