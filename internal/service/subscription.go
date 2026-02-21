package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/abramyaan/subscription-service/internal/domain"
	"github.com/abramyaan/subscription-service/internal/repository"
)

type SubscriptionService interface {
	Create(ctx context.Context, req domain.CreateSubscriptionRequest) (*domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	List(ctx context.Context, page, pageSize int) (*domain.ListResponse, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateSubscriptionRequest) (*domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CalculateCost(ctx context.Context, params domain.CostQueryParams) (*domain.CostResponse, error)
}

type subscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *slog.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *slog.Logger) SubscriptionService {
	return &subscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *subscriptionService) Create(ctx context.Context, req domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	s.logger.Info("creating new subscription",
		slog.String("service_name", req.ServiceName),
		slog.String("user_id", req.UserID),
	)

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		s.logger.Warn("invalid user_id format", slog.String("user_id", req.UserID))
		return nil, fmt.Errorf("invalid user_id format: %w", err)
	}

	if err := validateDateFormat(req.StartDate); err != nil {
		s.logger.Warn("invalid start_date format", slog.String("start_date", req.StartDate))
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	if req.EndDate != nil {
		if err := validateDateFormat(*req.EndDate); err != nil {
			s.logger.Warn("invalid end_date format", slog.String("end_date", *req.EndDate))
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}

		if !isEndDateValid(req.StartDate, *req.EndDate) {
			s.logger.Warn("end_date is before start_date",
				slog.String("start_date", req.StartDate),
				slog.String("end_date", *req.EndDate),
			)
			return nil, fmt.Errorf("end_date must be greater than or equal to start_date")
		}
	}

	now := time.Now()
	subscription := &domain.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		UserID:      userID,
		Cost:        req.Cost,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, subscription); err != nil {
		s.logger.Error("failed to create subscription in repository", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	s.logger.Info("subscription created successfully", slog.String("id", subscription.ID.String()))
	return subscription, nil
}

func (s *subscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	s.logger.Debug("fetching subscription", slog.String("id", id.String()))

	subscription, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			s.logger.Debug("subscription not found", slog.String("id", id.String()))
			return nil, err
		}
		s.logger.Error("failed to fetch subscription", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}

	return subscription, nil
}

func (s *subscriptionService) List(ctx context.Context, page, pageSize int) (*domain.ListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	s.logger.Debug("fetching subscriptions list",
		slog.Int("page", page),
		slog.Int("page_size", pageSize),
	)

	subscriptions, total, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		s.logger.Error("failed to fetch subscriptions list", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to fetch subscriptions: %w", err)
	}

	if subscriptions == nil {
		subscriptions = []domain.Subscription{}
	}

	totalPages := (total + pageSize - 1) / pageSize

	response := &domain.ListResponse{
		Data:       subscriptions,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	s.logger.Info("subscriptions list fetched",
		slog.Int("count", len(subscriptions)),
		slog.Int("total", total),
	)

	return response, nil
}

func (s *subscriptionService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateSubscriptionRequest) (*domain.Subscription, error) {
	s.logger.Info("updating subscription", slog.String("id", id.String()))

	subscription, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			s.logger.Debug("subscription not found for update", slog.String("id", id.String()))
			return nil, err
		}
		s.logger.Error("failed to fetch subscription for update", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}

	if req.ServiceName != nil {
		subscription.ServiceName = *req.ServiceName
	}
	if req.Cost != nil {
		subscription.Cost = *req.Cost
	}
	if req.StartDate != nil {
		if err := validateDateFormat(*req.StartDate); err != nil {
			s.logger.Warn("invalid start_date format", slog.String("start_date", *req.StartDate))
			return nil, fmt.Errorf("invalid start_date: %w", err)
		}
		subscription.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		if err := validateDateFormat(*req.EndDate); err != nil {
			s.logger.Warn("invalid end_date format", slog.String("end_date", *req.EndDate))
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}

		if !isEndDateValid(subscription.StartDate, *req.EndDate) {
			s.logger.Warn("end_date is before start_date",
				slog.String("start_date", subscription.StartDate),
				slog.String("end_date", *req.EndDate),
			)
			return nil, fmt.Errorf("end_date must be greater than or equal to start_date")
		}

		subscription.EndDate = req.EndDate
	}

	subscription.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, subscription); err != nil {
		s.logger.Error("failed to update subscription in repository", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logger.Info("subscription updated successfully", slog.String("id", id.String()))
	return subscription, nil
}

func (s *subscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("deleting subscription", slog.String("id", id.String()))

	if err := s.repo.Delete(ctx, id); err != nil {
		if err == repository.ErrNotFound {
			s.logger.Debug("subscription not found for deletion", slog.String("id", id.String()))
			return err
		}
		s.logger.Error("failed to delete subscription", slog.String("error", err.Error()))
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	s.logger.Info("subscription deleted successfully", slog.String("id", id.String()))
	return nil
}

func (s *subscriptionService) CalculateCost(ctx context.Context, params domain.CostQueryParams) (*domain.CostResponse, error) {
	s.logger.Info("calculating subscription cost",
		slog.String("start_date", params.StartDate),
		slog.String("end_date", params.EndDate),
	)

	if err := validateDateFormat(params.StartDate); err != nil {
		s.logger.Warn("invalid start_date format", slog.String("start_date", params.StartDate))
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	if err := validateDateFormat(params.EndDate); err != nil {
		s.logger.Warn("invalid end_date format", slog.String("end_date", params.EndDate))
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}

	if !isEndDateValid(params.StartDate, params.EndDate) {
		s.logger.Warn("end_date is before start_date",
			slog.String("start_date", params.StartDate),
			slog.String("end_date", params.EndDate),
		)
		return nil, fmt.Errorf("end_date must be greater than or equal to start_date")
	}

	totalCost, count, err := s.repo.CalculateCost(ctx, params)
	if err != nil {
		s.logger.Error("failed to calculate cost", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to calculate cost: %w", err)
	}

	period := fmt.Sprintf("%s to %s", params.StartDate, params.EndDate)

	response := &domain.CostResponse{
		TotalCost: totalCost,
		Period:    period,
		Count:     count,
	}

	s.logger.Info("cost calculated successfully",
		slog.Int("total_cost", totalCost),
		slog.Int("count", count),
	)

	return response, nil
}

func validateDateFormat(dateStr string) error {
	_, err := time.Parse("01-2006", dateStr)
	if err != nil {
		return fmt.Errorf("date must be in format MM-YYYY (e.g., 07-2025)")
	}
	return nil
}

func isEndDateValid(startDate, endDate string) bool {
	start, err1 := time.Parse("01-2006", startDate)
	end, err2 := time.Parse("01-2006", endDate)

	if err1 != nil || err2 != nil {
		return false
	}

	return !end.Before(start)
}
