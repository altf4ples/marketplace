package order

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /orders", h.Checkout)
	mux.HandleFunc("GET /orders/{orderID}", h.Get)
	mux.HandleFunc("POST /orders/{orderID}/events", h.ApplyEvent)
	mux.HandleFunc("GET /users/{userID}/orders", h.List)
}
