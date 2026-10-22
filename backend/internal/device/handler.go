package device

import (
	"chessbox-notification/internal/utils"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterDeviceRequest

	wr := utils.NewWriteReader[RegisterDeviceRequest, RegisterDeviceResponse](r, w)
	req = *wr.DecodeRequest(req)

	device, err := h.service.RegisterDevice(r.Context(), req)

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Error trying to register device",
			err.Error(),
		)
		return
	}

	response, err := wr.EncodeResponse(RegisterDeviceResponse{
		DeviceID:    device.ID,
		DeviceToken: device.CredentialHash,
	})

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Error trying to register device",
			err.Error(),
		)
		return
	}

	wr.WriteResponse(
		http.StatusOK,
		"Device registered successfuly",
		response,
	)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateDeviceRequest

	wr := utils.NewWriteReader[UpdateDeviceRequest, UpdateDeviceResponse](r, w)
	req = *wr.DecodeRequest(req)

	device, err := h.service.UpdateDevice(r.Context(), req)

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Error trying to register device",
			err.Error(),
		)
		return
	}

	response, err := wr.EncodeResponse(UpdateDeviceResponse{
		DeviceID: device.ID,
	})

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Error trying to update device",
			err.Error(),
		)
		return
	}

	wr.WriteResponse(
		http.StatusOK,
		"Device updated successfuly",
		response,
	)
}

func (h *Handler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req UpsertDeviceRequest

	wr := utils.NewWriteReader[UpsertDeviceRequest, UpsertDeviceResponse](r, w)
	req = *wr.DecodeRequest(req)

	credentials := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	newDevice, err := h.service.UpsertDevice(r.Context(), req, &credentials)

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Error trying to register device",
			err.Error(),
		)
		return
	}

	response, err := wr.EncodeResponse(UpsertDeviceResponse{
		DeviceID:    newDevice.ID,
		DeviceToken: newDevice.Credentials,
	})

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"Error trying to update device",
			err.Error(),
		)
		return
	}

	wr.WriteResponse(
		http.StatusOK,
		"Device updated successfuly",
		response,
	)
}
