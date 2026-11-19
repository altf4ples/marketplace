package users

import (
	"errors"
	"github.com/altf4ples/marketplace/internal/platform"
	"net/http"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := platform.DecodeJSON(w, r, &req); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	u, err := h.svc.CreateUser(r.Context(), req.Name, req.PhoneNumber)
	switch {
	case errors.Is(err, ErrNameRequired), errors.Is(err, ErrPhoneInvalid):
		platform.WriteError(w, 400, "validation_error", err.Error())
	case errors.Is(err, ErrConflict):
		platform.WriteError(w, 409, "conflict", err.Error())
	case err != nil:
		platform.LoggerFrom(r.Context()).Error("create user", "error", err)
		platform.WriteError(w, 500, "internal_error", "internal server error")
	default:
		platform.WriteJSON(w, http.StatusCreated, UserResponse{ID: u.ID.String(), Name: u.Name, PhoneNumber: u.PhoneNumber})
	}
}
