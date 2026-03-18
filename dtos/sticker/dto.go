package sticker

import (
	"time"

	"github.com/google/uuid"
)

// Request DTOs

type SendStickerRequest struct {
	ReceiverID    uuid.UUID `json:"receiver_id" required:"true" doc:"ID of the employee receiving the sticker"`
	StickerTypeID uuid.UUID `json:"sticker_type_id" required:"true" doc:"ID of the sticker type"`
	Message       string    `json:"message" maxLength:"255" doc:"Optional message with the sticker"`
}

type GetLeaderboardRequest struct {
	Limit        int        `query:"limit" minimum:"1" maximum:"100" default:"10" doc:"Number of results to return"`
	StartDate    *time.Time `query:"start_date" doc:"Filter by start date (RFC3339)"`
	EndDate      *time.Time `query:"end_date" doc:"Filter by end date (RFC3339)"`
	DepartmentID *uuid.UUID `query:"department_id" doc:"Filter by department ID"`
}

type UpdateGlobalConfigRequest struct {
	YearlyPoints int `json:"yearly_points" required:"true" minimum:"1" doc:"Points to allocate per year"`
	ResetMonth   int `json:"reset_month" required:"true" minimum:"1" maximum:"12" doc:"Month to reset points"`
	ResetDay     int `json:"reset_day" required:"true" minimum:"1" maximum:"31" doc:"Day to reset points"`
}

// Response DTOs

type PointBalanceResponse struct {
	EmployeeID    uuid.UUID `json:"employee_id"`
	Year          int       `json:"year"`
	InitialPoints int       `json:"initial_points"`
	CurrentPoints int       `json:"current_points"`
}

type LeaderboardResult struct {
	EmployeeID uuid.UUID `json:"employee_id"`
	FullName   string    `json:"full_name"`
	Total      int       `json:"total"`
}

type LeaderboardResponse struct {
	Body struct {
		Data []LeaderboardResult `json:"data"`
	}
}

type PointBalanceAPIResponse struct {
	Body struct {
		Data PointBalanceResponse `json:"data"`
	}
}

type SendStickerResponse struct {
	Body struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
}

type UpdateConfigResponse struct {
	Body struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
}

type StickerTypeResponse struct {
	ID           uuid.UUID `json:"sticker_type_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	PointCost    int       `json:"point_cost"`
	Category     string    `json:"category"`
	IconURL      string    `json:"icon_url"`
	IsActive     bool      `json:"is_active"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GetStickerTypesResponse struct {
	Body struct {
		Data []StickerTypeResponse `json:"data"`
	} `json:"body"`
}

type CreateStickerRequest struct {
	Name         string `json:"name" required:"true"`
	Description  string `json:"description"`
	PointCost    int    `json:"point_cost" required:"true"`
	Category     string `json:"category" required:"true"`
	IconURL      string `json:"icon_url" form:"icon_url"`
	DisplayOrder int    `json:"display_order"`
}

type CreateStickerResponse struct {
	Body struct {
		Success bool `json:"success"`
		Data    any  `json:"data"`
	} `json:"body"`
}
