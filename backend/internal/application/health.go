package application

import (
	"chessbox-notification/internal/utils"
	"net/http"
)

type Blank struct {}

type HealthResponse struct {}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	wr := utils.NewWriteReader[Blank, HealthResponse](r, w)

	wr.WriteResponse(
		http.StatusOK, 
		"All good",
		nil,
	)
}