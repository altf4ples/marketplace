package notification

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /users/{userID}/notifications", h.List)
}
