package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PersonID = uuid.UUID
type ServiceName = string
type SubscriptionID = uuid.UUID
type currencyRUB = int64

type Subscription struct {
	ID          SubscriptionID
	ServiceName ServiceName
	PriceRUB    currencyRUB
	StartedAt   time.Time
	CompletedAt *time.Time
	Owner       PersonID
}

func NewSubscription(owner PersonID, priceRUB int64, service ServiceName, startTime time.Time, endTime *time.Time) (sub *Subscription, err error) {
	startTime = startTime.UTC()
	if time.Now().UTC().Before(startTime) {
		return nil, fmt.Errorf("can not create subscription: date has not come yet")
	}

	if owner == uuid.Nil {
		return nil, fmt.Errorf("can not create subscription: user id was not provided")
	}

	service = strings.Trim(service, " ")
	if service == "" {
		return nil, fmt.Errorf("can not create subscription: subscribed service is not provided")
	}

	if priceRUB < 0 {
		return nil, fmt.Errorf("can not create subscription: invalid subscription price")
	}

	if endTime != nil {
		endTimeUTC := endTime.UTC()
		endTime = &endTimeUTC
		if endTime.Before(startTime) {
			return nil, fmt.Errorf("can not create subscription: start time must be less than end time")
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("can not create subscription: could not generate subscription id")
	}

	return &Subscription{ID: id, ServiceName: service, PriceRUB: priceRUB, StartedAt: startTime, CompletedAt: endTime, Owner: owner}, nil
}

func (s *Subscription) IsCompleted() bool {
	return s.CompletedAt != nil
}

func (s *Subscription) ChangeServiceName(srvName string) error {
	if srvName != "" {
		s.ServiceName = srvName
		return nil
	} else {
		return fmt.Errorf("can not update subscription: subscribed service is not provided")
	}
}

func (s *Subscription) ResetEndTime() {
	s.CompletedAt = nil
}

func (s *Subscription) ChangeStartTime(startTime time.Time) error {
	if s.CompletedAt != nil && startTime.UTC().After((*s.CompletedAt)) {
		return fmt.Errorf("can not update subscription: start time must be less than end time")
	} else if time.Now().UTC().Before(startTime.UTC()) {
		return fmt.Errorf("can not update subscription: date has not come yet")
	}

	s.StartedAt = startTime.UTC()
	return nil
}

func (s *Subscription) ChangeEndTime(endTime time.Time) error {
	if s.StartedAt.After(endTime.UTC()) {
		return fmt.Errorf("can not update subscription: start time must be less than end time")
	}

	utcEndTime := endTime.UTC()
	s.CompletedAt = &utcEndTime

	return nil
}

func (s *Subscription) ChangeOwner(owner PersonID) error {
	if owner == uuid.Nil {
		return fmt.Errorf("can not update subscription: user id was not provided")
	}
	s.Owner = owner

	return nil
}

func (s *Subscription) ChangePrice(price int64) error {
	if price < 0 {
		return fmt.Errorf("can not update subscription: invalid subscription price")
	}

	s.PriceRUB = price

	return nil
}
