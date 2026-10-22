package tournament

import (
	"chessbox-notification/internal/utils"
	"fmt"
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

func (h *Handler) SearchTournaments(w http.ResponseWriter, r *http.Request) {
	wr := utils.NewWriteReader[any, SearchTournamentsResponse](r, w)
	
	query := r.URL.Query().Get("q")
	tournaments, err := h.service.SearchTournaments(r.Context(), query)

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"An error occured when searching tournament",
			err.Error(),
		)
		return
	}

	response := make(SearchTournamentsResponse, 0)

	for _, tournament := range *tournaments {
		response = append(response, tournament)
	}

	data, err := wr.EncodeResponse(response)

	if err != nil {
		wr.WriteError(
			http.StatusBadRequest,
			"An error occured when searching tournament",
			err.Error(),
		)
		return
	}

	wr.WriteResponse(
		http.StatusOK,
		fmt.Sprintf("Search results found %d tournaments", len(*tournaments)),
		data,
	)
}