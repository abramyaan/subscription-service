package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/abramyaan/subscription-service/internal/domain"
	"github.com/abramyaan/subscription-service/internal/repository"
	"github.com/abramyaan/subscription-service/internal/service"
	"github.com/abramyaan/subscription-service/pkg/validator"
)

type SubscriptionHandler struct {
	service   service.SubscriptionService
	validator *validator.Validator
	logger    *slog.Logger
}

func NewSubscriptionHandler(service service.SubscriptionService, validator *validator.Validator, logger *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service:   service,
		validator: validator,
		logger:    logger,
	}
}

// Create создает новую подписку
// @Summary Create subscription
// @Description Create a new subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body domain.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} domain.Subscription
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateSubscriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		h.logger.Warn("validation failed", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, "Validation failed", validator.FormatValidationError(err))
		return
	}

	subscription, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create subscription", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to create subscription", err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, subscription)
}

// GetByID получает подписку по ID
// @Summary Get subscription by ID
// @Description Get a subscription by its ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Success 200 {object} domain.Subscription
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("invalid subscription id", slog.String("id", idStr))
		h.respondError(w, http.StatusBadRequest, "Invalid subscription ID", "ID must be a valid UUID")
		return
	}

	subscription, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			h.logger.Debug("subscription not found", slog.String("id", idStr))
			h.respondError(w, http.StatusNotFound, "Subscription not found", "")
			return
		}
		h.logger.Error("failed to fetch subscription", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch subscription", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, subscription)
}

// List получает список подписок с пагинацией
// @Summary List subscriptions
// @Description Get list of subscriptions with pagination
// @Tags subscriptions
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Success 200 {object} domain.ListResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	response, err := h.service.List(r.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to fetch subscriptions", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch subscriptions", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// Update обновляет подписку
// @Summary Update subscription
// @Description Update an existing subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Param subscription body domain.UpdateSubscriptionRequest true "Updated subscription data"
// @Success 200 {object} domain.Subscription
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("invalid subscription id", slog.String("id", idStr))
		h.respondError(w, http.StatusBadRequest, "Invalid subscription ID", "ID must be a valid UUID")
		return
	}

	var req domain.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		h.logger.Warn("validation failed", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, "Validation failed", validator.FormatValidationError(err))
		return
	}

	subscription, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		if err == repository.ErrNotFound {
			h.logger.Debug("subscription not found", slog.String("id", idStr))
			h.respondError(w, http.StatusNotFound, "Subscription not found", "")
			return
		}
		h.logger.Error("failed to update subscription", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to update subscription", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, subscription)
}

// Delete удаляет подписку
// @Summary Delete subscription
// @Description Delete a subscription by ID
// @Tags subscriptions
// @Param id path string true "Subscription ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("invalid subscription id", slog.String("id", idStr))
		h.respondError(w, http.StatusBadRequest, "Invalid subscription ID", "ID must be a valid UUID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if err == repository.ErrNotFound {
			h.logger.Debug("subscription not found", slog.String("id", idStr))
			h.respondError(w, http.StatusNotFound, "Subscription not found", "")
			return
		}
		h.logger.Error("failed to delete subscription", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to delete subscription", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CalculateCost рассчитывает суммарную стоимость подписок
// @Summary Calculate subscription cost
// @Description Calculate total cost of subscriptions for a period with optional filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID (UUID)"
// @Param service_name query string false "Service name"
// @Param start_date query string true "Start date (MM-YYYY)" example(01-2025)
// @Param end_date query string true "End date (MM-YYYY)" example(12-2025)
// @Success 200 {object} domain.CostResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/subscriptions/cost [get]
func (h *SubscriptionHandler) CalculateCost(w http.ResponseWriter, r *http.Request) {
	params := domain.CostQueryParams{
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
	}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.Warn("invalid user_id", slog.String("user_id", userIDStr))
			h.respondError(w, http.StatusBadRequest, "Invalid user_id", "user_id must be a valid UUID")
			return
		}
		params.UserID = &userID
	}

	if serviceName := r.URL.Query().Get("service_name"); serviceName != "" {
		params.ServiceName = &serviceName
	}

	if params.StartDate == "" {
		h.logger.Warn("missing start_date parameter")
		h.respondError(w, http.StatusBadRequest, "Missing start_date", "start_date is required")
		return
	}

	if params.EndDate == "" {
		h.logger.Warn("missing end_date parameter")
		h.respondError(w, http.StatusBadRequest, "Missing end_date", "end_date is required")
		return
	}

	response, err := h.service.CalculateCost(r.Context(), params)
	if err != nil {
		h.logger.Error("failed to calculate cost", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to calculate cost", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

func (h *SubscriptionHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error("failed to encode response", slog.String("error", err.Error()))
		}
	}
}

func (h *SubscriptionHandler) respondError(w http.ResponseWriter, status int, error, message string) {
	h.respondJSON(w, status, domain.ErrorResponse{
		Error:   error,
		Message: message,
	})
}
