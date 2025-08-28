package models

import (
	"time"

	"gorm.io/gorm"
)

// Channel represents a release channel in the database
type Channel struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	AppID             string         `json:"app_id" gorm:"size:32;uniqueIndex:idx_app_channel_name" validate:"required"`
	Name              string         `json:"name" gorm:"not null;size:50;uniqueIndex:idx_app_channel_name" validate:"required"`
	DisplayName       string         `json:"display_name" gorm:"not null;size:100" validate:"required"`
	Description       string         `json:"description" gorm:"type:text"`
	IsActive          bool           `json:"is_active" gorm:"default:true"`
	AutoPublish       bool           `json:"auto_publish" gorm:"default:false"`
	RolloutPercentage int            `json:"rollout_percentage" gorm:"default:100" validate:"min=0,max=100"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Application Application `json:"application,omitempty" gorm:"foreignKey:AppID;references:AppID"`
}

// TableName returns the table name for Channel model
func (Channel) TableName() string {
	return "channels"
}

// ChannelRequest represents the request payload for creating/updating channels
type ChannelRequest struct {
	AppID             string `json:"app_id" validate:"required"`
	Name              string `json:"name" validate:"required"`
	DisplayName       string `json:"display_name" validate:"required"`
	Description       string `json:"description"`
	IsActive          bool   `json:"is_active"`
	AutoPublish       bool   `json:"auto_publish"`
	RolloutPercentage int    `json:"rollout_percentage" validate:"min=0,max=100"`
}

// ChannelResponse represents the response payload for channel queries
type ChannelResponse struct {
	ID                uint      `json:"id"`
	AppID             string    `json:"app_id"`
	Name              string    `json:"name"`
	DisplayName       string    `json:"display_name"`
	Description       string    `json:"description"`
	IsActive          bool      `json:"is_active"`
	AutoPublish       bool      `json:"auto_publish"`
	RolloutPercentage int       `json:"rollout_percentage"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ToResponse converts Channel model to ChannelResponse
func (c *Channel) ToResponse() *ChannelResponse {
	return &ChannelResponse{
		ID:                c.ID,
		AppID:             c.AppID,
		Name:              c.Name,
		DisplayName:       c.DisplayName,
		Description:       c.Description,
		IsActive:          c.IsActive,
		AutoPublish:       c.AutoPublish,
		RolloutPercentage: c.RolloutPercentage,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}
