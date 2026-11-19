package users

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) { mux.HandleFunc("POST /users", h.Create) }
