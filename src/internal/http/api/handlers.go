package api

import (
	"context"
	"effective-mobile/internal/http/middleware"
	"effective-mobile/internal/models"
	"effective-mobile/internal/service"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type HandlersDependencies struct {
	Log                 *slog.Logger
	SubscriptionService service.SubscriptionService
}

func withReqIDLog(ctx context.Context, log *slog.Logger) *slog.Logger {
	reqID := ctx.Value(middleware.EnrichReqIDKey)
	if reqID == nil {
		reqID = "unknown"
	}

	return log.With(slog.String("request_id", reqID.(string)))
}

func handleServiceError(err error, log *slog.Logger, op string) error {
	if err == nil {
		return nil
	}

	if svcErr, ok := err.(*service.ServiceError); ok {
		switch svcErr.Code {
		case service.ErrInvalidInput:
			log.Warn("invalid input", slog.String("op", op), slog.String("error", svcErr.Message))
			return echo.NewHTTPError(http.StatusBadRequest, ErrorResponse{Error: svcErr.Message})
		case service.ErrNotFound:
			log.Warn("resource not found", slog.String("op", op), slog.String("error", svcErr.Message))
			return echo.NewHTTPError(http.StatusNotFound, ErrorResponse{Error: svcErr.Message})
		default:
			log.Error("internal error", slog.String("op", op), slog.String("error", svcErr.Message))
			return echo.NewHTTPError(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
	}

	log.Error("unexpected error", slog.String("op", op), slog.Any("error", err))
	return echo.NewHTTPError(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
}

func (h HandlersDependencies) GetSubscriptions(ctx context.Context, request GetSubscriptionsRequestObject) (GetSubscriptionsResponseObject, error) {
	const op = "internal.http.api.handlers.GetSubscriptions"
	log := withReqIDLog(ctx, h.Log)

	allSubs := h.SubscriptionService.GetSubscriptions(ctx)
	responseModels := make(GetSubscriptions200JSONResponse, len(allSubs))
	for i, sub := range allSubs {
		vm := toViewModel(sub)
		responseModels[i] = Subscription{
			Id:          vm.Id,
			Price:       vm.Price,
			ServiceName: vm.ServiceName,
			StartDate:   vm.StartDate,
			EndDate:     vm.EndDate,
			UserId:      vm.UserId,
		}
	}

	log.Info("subscriptions fetched", slog.String("op", op), slog.Int("count", len(responseModels)))
	return responseModels, nil
}

func (h HandlersDependencies) PostSubscriptions(ctx context.Context, request PostSubscriptionsRequestObject) (PostSubscriptionsResponseObject, error) {
	const op = "internal.http.api.handlers.PostSubscriptions"
	log := withReqIDLog(ctx, h.Log)

	if request.Body == nil {
		log.Warn("invalid request: body is nil", slog.String("op", op))
		return nil, echo.NewHTTPError(http.StatusBadRequest, ErrorResponse{Error: "request body is required"})
	}

	var endTime *time.Time
	if endDate := request.Body.EndDate; endDate != nil {
		endTime = &endDate.Time
	}

	args := service.CreateNewSubscriptionArgs{
		UserID:    request.Body.UserId,
		Service:   request.Body.ServiceName,
		StartTime: request.Body.StartDate.Time,
		EndTime:   endTime,
		PriceRUB:  request.Body.Price,
	}

	sub, err := h.SubscriptionService.CreateNewSubscription(ctx, args)
	if err != nil {
		return nil, handleServiceError(err, h.Log, op)
	}

	vm := toViewModel(sub)
	log.Info("subscription created", slog.String("op", op), slog.Any("subscription_id", sub.ID))
	return PostSubscriptions201JSONResponse{
		EndDate:     vm.EndDate,
		Id:          vm.Id,
		Price:       vm.Price,
		ServiceName: vm.ServiceName,
		StartDate:   vm.StartDate,
		UserId:      vm.UserId,
	}, nil
}

func (h HandlersDependencies) GetSubscriptionsId(ctx context.Context, request GetSubscriptionsIdRequestObject) (GetSubscriptionsIdResponseObject, error) {
	const op = "internal.http.api.handlers.GetSubscriptionsId"
	log := withReqIDLog(ctx, h.Log)

	sub, err := h.SubscriptionService.FindSubscriptionByID(ctx, request.Id)
	if err != nil {
		return nil, handleServiceError(err, h.Log, op)
	}

	vm := toViewModel(sub)
	log.Info("subscription fetched", slog.String("op", op), slog.Any("subscription_id", sub.ID))
	return GetSubscriptionsId200JSONResponse{
		Id:          vm.Id,
		EndDate:     vm.EndDate,
		Price:       vm.Price,
		ServiceName: vm.ServiceName,
		StartDate:   vm.StartDate,
		UserId:      vm.UserId,
	}, nil
}

func (h HandlersDependencies) GetSubscriptionsTotalCost(ctx context.Context, request GetSubscriptionsTotalCostRequestObject) (GetSubscriptionsTotalCostResponseObject, error) {
	const op = "internal.http.api.handlers.GetSubscriptionsTotalCost"
	log := withReqIDLog(ctx, h.Log)

	params := request.Params
	var countStartTime, countEndTime *time.Time
	if params.StartDate != nil {
		countStartTime = &params.StartDate.Time
	}
	if params.EndDate != nil {
		countEndTime = &params.EndDate.Time
	}

	totalPrice, err := h.SubscriptionService.CalculateTotalSubscriptionsPrice(ctx, params.UserId, params.ServiceName, countStartTime, countEndTime)
	if err != nil {
		return nil, handleServiceError(err, h.Log, op)
	}

	log.Info("total cost calculated", slog.String("op", op), slog.Any("user_id", params.UserId), slog.Int64("total_cost", totalPrice.TotalPriceRUB))
	return GetSubscriptionsTotalCost200JSONResponse{
		TotalCost: &totalPrice.TotalPriceRUB,
	}, nil
}

func (h HandlersDependencies) DeleteSubscriptionsId(ctx context.Context, request DeleteSubscriptionsIdRequestObject) (DeleteSubscriptionsIdResponseObject, error) {
	const op = "internal.http.api.handlers.DeleteSubscriptionsId"
	log := withReqIDLog(ctx, h.Log)

	err := h.SubscriptionService.RemoveExistingSubscription(ctx, request.Id)
	if err != nil {
		return nil, handleServiceError(err, h.Log, op)
	}

	log.Info("subscription deleted", slog.String("op", op), slog.Any("subscription_id", request.Id))
	return DeleteSubscriptionsId204Response{}, nil
}

func (h HandlersDependencies) PatchSubscriptionsId(ctx context.Context, request PatchSubscriptionsIdRequestObject) (PatchSubscriptionsIdResponseObject, error) {
	const op = "internal.http.api.handlers.PatchSubscriptionsId"
	log := withReqIDLog(ctx, h.Log)

	if request.Body == nil {
		log.Warn("invalid request: body is nil", slog.String("op", op))
		return nil, echo.NewHTTPError(http.StatusBadRequest, ErrorResponse{Error: "request body is required"})
	}

	var endTime *time.Time
	if request.Body.EndDate != nil {
		endTime = &request.Body.EndDate.Time
	}
	args := service.UpdateExistingSubscriptionArgs{
		SubscriptionID: request.Id,
		UserID:         request.Body.UserId,
		Service:        request.Body.ServiceName,
		StartTime:      request.Body.StartDate.Time,
		EndTime:        endTime,
		PriceRUB:       request.Body.Price,
	}

	sub, err := h.SubscriptionService.UpdateExistingSubscription(ctx, args)
	if err != nil {
		return nil, handleServiceError(err, h.Log, op)
	}

	vm := toViewModel(sub)
	log.Info("subscription updated", slog.String("op", op), slog.Any("subscription_id", sub.ID))
	return PatchSubscriptionsId200JSONResponse{
		EndDate:     vm.EndDate,
		Id:          vm.Id,
		Price:       vm.Price,
		ServiceName: vm.ServiceName,
		StartDate:   vm.StartDate,
		UserId:      vm.UserId,
	}, nil
}

func toViewModel(sub *models.Subscription) Subscription {
	var endDate *openapi_types.Date
	if sub.IsCompleted() {
		endDate = &openapi_types.Date{
			Time: (*sub.CompletedAt).UTC(),
		}
	}

	return Subscription{
		EndDate:     endDate,
		Id:          sub.ID,
		Price:       sub.PriceRUB,
		ServiceName: sub.ServiceName,
		StartDate:   openapi_types.Date{Time: sub.StartedAt.UTC()},
		UserId:      sub.Owner,
	}
}
