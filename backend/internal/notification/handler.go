package notification

import (
	"chessbox-notification/internal/middleware"
	"chessbox-notification/internal/utils"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	wr := utils.NewWriteReader[any, any](r, w)
	device := middleware.DeviceFromContext(r.Context())

	if device == nil {
		wr.WriteError(http.StatusUnauthorized, "Unauthorized", "Token not found")
		return
	}

	notifications, err := h.service.List(r.Context(), device.ID)
	if err != nil {
		wr.WriteError(http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	encoded, err := wr.EncodeResponse(notifications)
	if err != nil {
		wr.WriteError(http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	wr.WriteResponse(http.StatusOK, "Success", encoded)
}
