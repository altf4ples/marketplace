package order

import (
	"errors"
	"github.com/altf4ples/marketplace/internal/platform"
	"github.com/google/uuid"
	"net/http"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := platform.DecodeJSON(w, r, &req); err != nil {
		platform.WriteError(w, 400, "invalid_request", err.Error())
		return
	}
	o, err := h.svc.Checkout(r.Context(), req.UserID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, o)
}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("orderID"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_order_id", "orderID must be a UUID")
		return
	}
	userID, err := uuid.Parse(r.URL.Query().Get("user_id"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_user_id", "user_id must be a UUID")
		return
	}
	o, err := h.svc.Get(r.Context(), id, userID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	platform.WriteJSON(w, 200, o)
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("userID"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_user_id", "userID must be a UUID")
		return
	}
	orders, err := h.svc.List(r.Context(), userID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	platform.WriteJSON(w, 200, map[string]any{"orders": orders})
}
func (h *Handler) ApplyEvent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("orderID"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_order_id", "orderID must be a UUID")
		return
	}
	var req EventRequest
	if err := platform.DecodeJSON(w, r, &req); err != nil {
		platform.WriteError(w, 400, "invalid_request", err.Error())
		return
	}
	o, err := h.svc.ApplyEvent(r.Context(), id, req.UserID, req.Event)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	platform.WriteJSON(w, 200, o)
}
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var restoreErr *CartRestoreError
	if errors.As(err, &restoreErr) {
		platform.LoggerFrom(r.Context()).Error("checkout cart restoration failed", "error", err)
		platform.WriteError(w, 500, "checkout_recovery_failed", "checkout failed and cart recovery needs attention")
		return
	}
	switch {
	case errors.Is(err, ErrCartEmpty), errors.Is(err, ErrInvalidQuantity):
		platform.WriteError(w, 400, "validation_error", err.Error())
	case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrProductNotFound), errors.Is(err, ErrOrderNotFound):
		platform.WriteError(w, 404, "not_found", err.Error())
	case errors.Is(err, ErrForbidden):
		platform.WriteError(w, 403, "forbidden", err.Error())
	case errors.Is(err, ErrInvalidTransition):
		platform.WriteError(w, 409, "invalid_transition", err.Error())
	default:
		platform.LoggerFrom(r.Context()).Error("order request failed", "error", err)
		platform.WriteError(w, 500, "internal_error", "internal server error")
	}
}
