package service

import (
	"context"
	"effective-mobile/internal/models"
	"effective-mobile/internal/storage"
	"effective-mobile/pkg/logger/sl"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

func NewSubscriptionService(s storage.SubscriptionsStorage, log *slog.Logger) SubscriptionService {
	return subscriptionService{
		subscriptionsStorage: s,
		log:                  log.With(slog.String("component", "SubscriptionService")),
	}
}

type subscriptionService struct {
	subscriptionsStorage storage.SubscriptionsStorage
	log                  *slog.Logger
}

func (s subscriptionService) CreateNewSubscription(ctx context.Context, c CreateNewSubscriptionArgs) (*models.Subscription, error) {
	const op = "internal.service.impl.CreateNewSubscription"

	sub, err := models.NewSubscription(c.UserID, c.PriceRUB, c.Service, c.StartTime, c.EndTime)
	if err != nil {
		s.log.Debug("validation failed", slog.String("op", op), sl.Err(err))
		return nil, NewInvalidInputError(err.Error())
	}

	err = s.subscriptionsStorage.Add(ctx, *sub)
	if err != nil {
		s.log.Error("failed to add subscription", slog.String("op", op), sl.Err(err), slog.Any("subscription_id", sub.ID))
		return nil, NewInternalError("failed to create subscription")
	}

	s.log.Info("subscription created", slog.String("op", op), slog.Any("subscription_id", sub.ID))
	return sub, nil
}

func (s subscriptionService) UpdateExistingSubscription(ctx context.Context, u UpdateExistingSubscriptionArgs) (*models.Subscription, error) {
	const op = "internal.service.impl.UpdateExistingSubscription"

	sub, err := s.subscriptionsStorage.FindByID(ctx, u.SubscriptionID)
	if err != nil {
		if errors.Is(err, storage.ErrSubscriptionNotFound) {
			s.log.Warn("subscription not found", slog.String("op", op), slog.Any("subscription_id", u.SubscriptionID))
			return nil, NewNotFoundError("subscription not found")
		}
		s.log.Error("failed to find subscription", slog.String("op", op), sl.Err(err), slog.Any("subscription_id", u.SubscriptionID))
		return nil, NewInternalError("failed to update subscription")
	}

	if err := sub.ChangeOwner(u.UserID); err != nil {
		s.log.Warn("invalid user id", slog.String("op", op), sl.Err(err))
		return nil, NewInvalidInputError(err.Error())
	}
	sub.ResetEndTime()
	if err := sub.ChangeStartTime(u.StartTime); err != nil {
		s.log.Warn("invalid start time", slog.String("op", op), sl.Err(err))
		return nil, NewInvalidInputError(err.Error())
	}
	if u.EndTime != nil {
		if err := sub.ChangeEndTime(*u.EndTime); err != nil {
			s.log.Warn("invalid end time", slog.String("op", op), sl.Err(err))
			return nil, NewInvalidInputError(err.Error())
		}
	}
	if err := sub.ChangePrice(u.PriceRUB); err != nil {
		s.log.Warn("invalid price", slog.String("op", op), sl.Err(err))
		return nil, NewInvalidInputError(err.Error())
	}
	if err := sub.ChangeServiceName(u.Service); err != nil {
		s.log.Warn("invalid service name", slog.String("op", op), sl.Err(err))
		return nil, NewInvalidInputError(err.Error())
	}

	err = s.subscriptionsStorage.Update(ctx, *sub)
	if err != nil {
		s.log.Error("failed to update subscription", slog.String("op", op), sl.Err(err), slog.Any("subscription_id", sub.ID))
		return nil, NewInternalError("failed to update subscription")
	}

	s.log.Info("subscription updated", slog.String("op", op), slog.Any("subscription_id", sub.ID))
	return sub, nil
}

func (s subscriptionService) FindSubscriptionByID(ctx context.Context, id models.SubscriptionID) (*models.Subscription, error) {
	const op = "internal.service.impl.FindSubscriptionByID"

	sub, err := s.subscriptionsStorage.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrSubscriptionNotFound) {
			s.log.Warn("subscription not found", slog.String("op", op), slog.Any("subscription_id", id))
			return nil, NewNotFoundError("subscription not found")
		}
		s.log.Error("failed to find subscription", slog.String("op", op), sl.Err(err), slog.Any("subscription_id", id))
		return nil, NewInternalError("failed to fetch subscription")
	}

	s.log.Info("subscription fetched", slog.String("op", op), slog.Any("subscription_id", id))
	return sub, nil
}

func (s subscriptionService) GetSubscriptions(ctx context.Context) []*models.Subscription {
	const op = "internal.service.impl.GetSubscriptions"

	subs, err := s.subscriptionsStorage.Find(ctx, storage.SubscriptionsFilter{})
	if err != nil {
		s.log.Error("failed to fetch subscriptions", slog.String("op", op), sl.Err(err))
		return nil
	}

	s.log.Info("subscriptions fetched", slog.String("op", op), slog.Int("count", len(subs)))
	return subs
}

func (s subscriptionService) RemoveExistingSubscription(ctx context.Context, id models.SubscriptionID) error {
	const op = "internal.service.impl.RemoveExistingSubscription"

	err := s.subscriptionsStorage.RemoveByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrSubscriptionNotFound) {
			s.log.Warn("subscription not found", slog.String("op", op), slog.Any("subscription_id", id))
			return NewNotFoundError("subscription not found")
		}
		s.log.Error("failed to remove subscription", slog.String("op", op), sl.Err(err), slog.Any("subscription_id", id))
		return NewInternalError("failed to remove subscription")
	}

	s.log.Info("subscription removed", slog.String("op", op), slog.Any("subscription_id", id))
	return nil
}

func (s subscriptionService) CalculateTotalSubscriptionsPrice(ctx context.Context, userID models.PersonID, serviceName models.ServiceName, startTime, endTime *time.Time) (totalSubscriptionsPrice, error) {
	const op = "internal.service.impl.CalculateTotalSubscriptionsPrice"

	if userID == uuid.Nil {
		s.log.Warn("invalid input: user_id is empty", slog.String("op", op))
		return totalSubscriptionsPrice{}, NewInvalidInputError("user_id is required")
	}
	if serviceName == "" {
		s.log.Warn("invalid input: service_name is empty", slog.String("op", op))
		return totalSubscriptionsPrice{}, NewInvalidInputError("service_name is required")
	}
	if startTime != nil && endTime != nil && !endTime.After(*startTime) {
		s.log.Warn("invalid input: end time must be after start time", slog.String("op", op))
		return totalSubscriptionsPrice{}, NewInvalidInputError("end time must be after start time")
	}

	f := storage.SubscriptionsFilter{
		OwnerID:     userID,
		ServiceName: serviceName,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	subs, err := s.subscriptionsStorage.Find(ctx, f)
	if err != nil {
		s.log.Error("failed to fetch subscriptions", slog.String("op", op), sl.Err(err), slog.Any("owner_id", userID))
		return totalSubscriptionsPrice{}, NewInternalError("failed to calculate total cost")
	}

	var totalCost int64
	for _, sub := range subs {
		totalCost += sub.PriceRUB
	}

	s.log.Info("total cost calculated", slog.String("op", op), slog.Any("owner_id", userID), slog.Int64("total_cost", totalCost))
	return totalSubscriptionsPrice{TotalPriceRUB: totalCost}, nil
}
