package models

import (
	"time"

	"gorm.io/gorm"
)

// Channel represents a global release channel in the database
type Channel struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:50;uniqueIndex" validate:"required"`
	DisplayName string         `json:"display_name" gorm:"not null;size:100" validate:"required"`
	Description string         `json:"description" gorm:"type:text"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	SortOrder   int            `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations - Many-to-many relationship with applications
	Applications []Application `json:"applications,omitempty" gorm:"many2many:application_channels;joinForeignKey:ChannelName;joinReferences:AppID;foreignKey:Name;references:AppID"`
}

// TableName returns the table name for Channel model
func (Channel) TableName() string {
	return "channels"
}

// ChannelRequest represents the request payload for creating/updating global channels
type ChannelRequest struct {
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

// ChannelResponse represents the response payload for global channel queries
type ChannelResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ApplicationChannel represents the many-to-many relationship between applications and channels
type ApplicationChannel struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	AppID             string         `json:"app_id" gorm:"size:32;uniqueIndex:idx_app_channel" validate:"required"`
	ChannelName       string         `json:"channel_name" gorm:"size:50;uniqueIndex:idx_app_channel" validate:"required"`
	IsEnabled         bool           `json:"is_enabled" gorm:"default:true"`
	AutoPublish       bool           `json:"auto_publish" gorm:"default:false"`
	RolloutPercentage int            `json:"rollout_percentage" gorm:"default:100" validate:"min=0,max=100"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Application Application `json:"application,omitempty" gorm:"foreignKey:AppID;references:AppID"`
	Channel     Channel     `json:"channel,omitempty" gorm:"foreignKey:ChannelName;references:Name"`
}

// TableName returns the table name for ApplicationChannel model
func (ApplicationChannel) TableName() string {
	return "application_channels"
}

// ApplicationChannelRequest represents the request payload for updating app-channel relationships
type ApplicationChannelRequest struct {
	AppID             string `json:"app_id" validate:"required"`
	ChannelName       string `json:"channel_name" validate:"required"`
	IsEnabled         bool   `json:"is_enabled"`
	AutoPublish       bool   `json:"auto_publish"`
	RolloutPercentage int    `json:"rollout_percentage" validate:"min=0,max=100"`
}

// ApplicationChannelResponse represents the response payload for app-channel relationships
type ApplicationChannelResponse struct {
	ID                 uint      `json:"id"`
	AppID              string    `json:"app_id"`
	ChannelName        string    `json:"channel_name"`
	ChannelDisplayName string    `json:"channel_display_name,omitempty"`
	IsEnabled          bool      `json:"is_enabled"`
	AutoPublish        bool      `json:"auto_publish"`
	RolloutPercentage  int       `json:"rollout_percentage"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ToResponse converts ApplicationChannel model to ApplicationChannelResponse
func (ac *ApplicationChannel) ToResponse() *ApplicationChannelResponse {
	resp := &ApplicationChannelResponse{
		ID:                ac.ID,
		AppID:             ac.AppID,
		ChannelName:       ac.ChannelName,
		IsEnabled:         ac.IsEnabled,
		AutoPublish:       ac.AutoPublish,
		RolloutPercentage: ac.RolloutPercentage,
		CreatedAt:         ac.CreatedAt,
		UpdatedAt:         ac.UpdatedAt,
	}

	// Include channel display name if available
	if ac.Channel.DisplayName != "" {
		resp.ChannelDisplayName = ac.Channel.DisplayName
	}

	return resp
}

// ToResponse converts Channel model to ChannelResponse
func (c *Channel) ToResponse() *ChannelResponse {
	return &ChannelResponse{
		ID:          c.ID,
		Name:        c.Name,
		DisplayName: c.DisplayName,
		Description: c.Description,
		IsActive:    c.IsActive,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
