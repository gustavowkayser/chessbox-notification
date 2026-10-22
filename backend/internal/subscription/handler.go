package subscription

import (
	"chessbox-notification/internal/middleware"
	"chessbox-notification/internal/utils"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest

	wr := utils.NewWriteReader[CreateSubscriptionRequest, CreateSubscriptionResponse](r, w)
	req = *wr.DecodeRequest(req)

	device := middleware.DeviceFromContext(r.Context())

	subscription, err := h.service.Subscribe(r.Context(), CreateSubscriptionInput{
		DeviceID:      device.ID,
		TournamentURL: req.TournamentURL,
	})

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Error trying to subscribe to tournament",
			err.Error(),
		)
		return
	}

	response, err := wr.EncodeResponse(CreateSubscriptionResponse{
		SubscriptionID: subscription.ID,
	})

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Error trying to subscribe to tournament",
			err.Error(),
		)
		return
	}

	wr.WriteResponse(
		http.StatusOK,
		"Subscribed to the tournament successfuly",
		response,
	)
}

func (h *Handler) RemoveSubscription(w http.ResponseWriter, r *http.Request) {
	wr := utils.NewWriteReader[RemoveSubscriptionRequest, RemoveSubscriptionResponse](r, w)
	id := chi.URLParam(r, "id")

	device := middleware.DeviceFromContext(r.Context())

	err := h.service.Unsubscribe(r.Context(), RemoveSubscriptionInput{
		DeviceID:       device.ID,
		SubscriptionID: id,
	})

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Could not unsubscribe from tournament",
			err.Error(),
		)
		return
	}
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {

	wr := utils.NewWriteReader[any, ListSubscriptionsResponse](r, w)
	device := middleware.DeviceFromContext(r.Context())

	subscriptions, err := h.service.ListSubscriptions(r.Context(), device.ID)

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Could not list subscriptions",
			err.Error(),
		)
		return
	}

	response := make(ListSubscriptionsResponse, 0)

	for _, sub := range *subscriptions {
		response = append(response, SubscriptionView{
			ID:                    sub.ID,
			TournamentID:          sub.TournamentID,
			TournamentName:        sub.TournamentName,
			TournamentURL:         sub.TournamentURL,
			TournamentRound:       sub.TournamentRound,
			TournamentTotalRounds: sub.TournamentTotalRounds,
		})
	}

	encoded, err := wr.EncodeResponse(response)

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Could not list subscriptions",
			err.Error(),
		)
		return
	}

	wr.WriteResponse(
		http.StatusOK,
		"Listed subscriptions successfuly",
		encoded,
	)
}
