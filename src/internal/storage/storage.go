package storage

import (
	"context"
	"effective-mobile/internal/models"
	"errors"
	"time"
)

// ErrSubscriptionNotFound is returned when a subscription is not found in the database.
var ErrSubscriptionNotFound = errors.New("subscription not found")

type SubscriptionsStorage interface {
	Add(ctx context.Context, s models.Subscription) error
	RemoveByID(ctx context.Context, id models.SubscriptionID) error
	Update(ctx context.Context, s models.Subscription) error
	FindByID(ctx context.Context, id models.SubscriptionID) (*models.Subscription, error)
	Find(ctx context.Context, f SubscriptionsFilter) ([]*models.Subscription, error)
}

type SubscriptionsFilter struct {
	ServiceName models.ServiceName
	OwnerID     models.PersonID
	StartTime   *time.Time
	EndTime     *time.Time
}
