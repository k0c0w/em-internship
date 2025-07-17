package service

import (
	"context"
	"effective-mobile/internal/models"
	"fmt"
	"time"
)

type ErrorCode string

const (
	ErrInvalidInput ErrorCode = "invalid_input"
	ErrNotFound     ErrorCode = "not_found"
	ErrInternal     ErrorCode = "internal"
)

type ServiceError struct {
	Code    ErrorCode
	Message string
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewInvalidInputError(message string) *ServiceError {
	return &ServiceError{Code: ErrInvalidInput, Message: message}
}

func NewNotFoundError(message string) *ServiceError {
	return &ServiceError{Code: ErrNotFound, Message: message}
}

func NewInternalError(message string) *ServiceError {
	return &ServiceError{Code: ErrInternal, Message: message}
}

type CreateNewSubscriptionArgs struct {
	UserID    models.PersonID
	Service   models.ServiceName
	StartTime time.Time
	EndTime   *time.Time
	PriceRUB  int64
}

type UpdateExistingSubscriptionArgs struct {
	models.SubscriptionID
	UserID    models.PersonID
	Service   models.ServiceName
	StartTime time.Time
	EndTime   *time.Time
	PriceRUB  int64
}

type totalSubscriptionsPrice struct {
	TotalPriceRUB int64
}

type SubscriptionService interface {
	CreateNewSubscription(ctx context.Context, c CreateNewSubscriptionArgs) (*models.Subscription, error)
	UpdateExistingSubscription(ctx context.Context, u UpdateExistingSubscriptionArgs) (*models.Subscription, error)
	FindSubscriptionByID(ctx context.Context, id models.SubscriptionID) (*models.Subscription, error)
	GetSubscriptions(ctx context.Context) []*models.Subscription
	CalculateTotalSubscriptionsPrice(ctx context.Context, userID models.PersonID, serviceName models.ServiceName, startTime *time.Time, endTime *time.Time) (totalSubscriptionsPrice, error)
	RemoveExistingSubscription(ctx context.Context, id models.SubscriptionID) error
}
