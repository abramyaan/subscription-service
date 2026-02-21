package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/abramyaan/subscription-service/internal/domain"
)

var (
	ErrNotFound      = errors.New("subscription not found")
	ErrAlreadyExists = errors.New("subscription already exists")
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	List(ctx context.Context, limit, offset int) ([]domain.Subscription, int, error)
	Update(ctx context.Context, sub *domain.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	CalculateCost(ctx context.Context, params domain.CostQueryParams) (int, int, error)
}

type subscriptionRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewSubscriptionRepository(pool *pgxpool.Pool, logger *slog.Logger) SubscriptionRepository {
	return &subscriptionRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *subscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, service_name, user_id, cost, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	r.logger.Debug("creating subscription",
		slog.String("id", sub.ID.String()),
		slog.String("service_name", sub.ServiceName),
		slog.String("user_id", sub.UserID.String()),
	)

	_, err := r.pool.Exec(ctx, query,
		sub.ID,
		sub.ServiceName,
		sub.UserID,
		sub.Cost,
		sub.StartDate,
		sub.EndDate,
		sub.CreatedAt,
		sub.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create subscription",
			slog.String("error", err.Error()),
			slog.String("id", sub.ID.String()),
		)
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	r.logger.Info("subscription created successfully", slog.String("id", sub.ID.String()))
	return nil
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	query := `
		SELECT id, service_name, user_id, cost, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	r.logger.Debug("fetching subscription by id", slog.String("id", id.String()))

	var sub domain.Subscription
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.UserID,
		&sub.Cost,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("subscription not found", slog.String("id", id.String()))
			return nil, ErrNotFound
		}
		r.logger.Error("failed to fetch subscription",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
		)
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}

	r.logger.Debug("subscription fetched successfully", slog.String("id", id.String()))
	return &sub, nil
}

func (r *subscriptionRepository) List(ctx context.Context, limit, offset int) ([]domain.Subscription, int, error) {
	r.logger.Debug("fetching subscriptions list",
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	var total int
	countQuery := `SELECT COUNT(*) FROM subscriptions`
	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count subscriptions", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	query := `
		SELECT id, service_name, user_id, cost, start_date, end_date, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("failed to fetch subscriptions", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("failed to fetch subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.UserID,
			&sub.Cost,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan subscription", slog.String("error", err.Error()))
			return nil, 0, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating subscriptions", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("error iterating subscriptions: %w", err)
	}

	r.logger.Info("subscriptions fetched successfully",
		slog.Int("count", len(subscriptions)),
		slog.Int("total", total),
	)

	return subscriptions, total, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, sub *domain.Subscription) error {
	query := `
		UPDATE subscriptions
		SET service_name = $1, cost = $2, start_date = $3, end_date = $4, updated_at = $5
		WHERE id = $6
	`

	r.logger.Debug("updating subscription", slog.String("id", sub.ID.String()))

	result, err := r.pool.Exec(ctx, query,
		sub.ServiceName,
		sub.Cost,
		sub.StartDate,
		sub.EndDate,
		sub.UpdatedAt,
		sub.ID,
	)

	if err != nil {
		r.logger.Error("failed to update subscription",
			slog.String("error", err.Error()),
			slog.String("id", sub.ID.String()),
		)
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		r.logger.Debug("subscription not found for update", slog.String("id", sub.ID.String()))
		return ErrNotFound
	}

	r.logger.Info("subscription updated successfully", slog.String("id", sub.ID.String()))
	return nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	r.logger.Debug("deleting subscription", slog.String("id", id.String()))

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete subscription",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
		)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		r.logger.Debug("subscription not found for deletion", slog.String("id", id.String()))
		return ErrNotFound
	}

	r.logger.Info("subscription deleted successfully", slog.String("id", id.String()))
	return nil
}

func (r *subscriptionRepository) CalculateCost(ctx context.Context, params domain.CostQueryParams) (int, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("start_date <= $%d", argIdx))
	args = append(args, params.EndDate)
	argIdx++

	conditions = append(conditions, fmt.Sprintf("(end_date IS NULL OR end_date >= $%d)", argIdx))
	args = append(args, params.StartDate)
	argIdx++

	if params.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *params.UserID)
		argIdx++
	}

	if params.ServiceName != nil {
		conditions = append(conditions, fmt.Sprintf("service_name = $%d", argIdx))
		args = append(args, *params.ServiceName)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(cost), 0), COUNT(*)
		FROM subscriptions
		WHERE %s
	`, whereClause)

	r.logger.Debug("calculating subscription cost",
		slog.String("start_date", params.StartDate),
		slog.String("end_date", params.EndDate),
		slog.Any("user_id", params.UserID),
		slog.Any("service_name", params.ServiceName),
	)

	var totalCost int
	var count int

	err := r.pool.QueryRow(ctx, query, args...).Scan(&totalCost, &count)
	if err != nil {
		r.logger.Error("failed to calculate cost", slog.String("error", err.Error()))
		return 0, 0, fmt.Errorf("failed to calculate cost: %w", err)
	}

	r.logger.Info("subscription cost calculated",
		slog.Int("total_cost", totalCost),
		slog.Int("count", count),
	)

	return totalCost, count, nil
}

