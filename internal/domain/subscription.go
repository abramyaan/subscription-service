package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name" validate:"required,min=1,max=255"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id" validate:"required,uuid"`
	Cost        int        `json:"cost" db:"cost" validate:"required,min=0"`
	StartDate   string     `json:"start_date" db:"start_date" validate:"required,datetime=01-2026"`
	EndDate     *string    `json:"end_date,omitempty" db:"end_date" validate:"omitempty,datetime=01-2026"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" validate:"required,min=1,max=255"`
	UserID      string  `json:"user_id" validate:"required,uuid"`
	Cost        int     `json:"cost" validate:"required,min=0"`
	StartDate   string  `json:"start_date" validate:"required"`
	EndDate     *string `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty" validate:"omitempty,min=1,max=255"`
	Cost        *int    `json:"cost,omitempty" validate:"omitempty,min=0"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}

type CostQueryParams struct {
	UserID      *uuid.UUID `query:"user_id" validate:"omitempty,uuid"`
	ServiceName *string    `query:"service_name" validate:"omitempty,min=1"`
	StartDate   string     `query:"start_date" validate:"required"`
	EndDate     string     `query:"end_date" validate:"required"`
}

type CostResponse struct {
	TotalCost int    `json:"total_cost"`
	Period    string `json:"period"`
	Count     int    `json:"count"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type ListResponse struct {
	Data       []Subscription `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
